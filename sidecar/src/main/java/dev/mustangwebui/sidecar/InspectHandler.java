package dev.mustangwebui.sidecar;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpExchange;
import org.apache.pdfbox.Loader;
import org.apache.pdfbox.pdmodel.PDDocument;
import org.apache.pdfbox.pdmodel.PDDocumentInformation;
import org.apache.pdfbox.pdmodel.PDDocumentNameDictionary;
import org.apache.pdfbox.pdmodel.PDEmbeddedFilesNameTreeNode;
import org.apache.pdfbox.pdmodel.common.PDMetadata;
import org.apache.pdfbox.pdmodel.common.filespecification.PDComplexFileSpecification;
import org.mustangproject.BankDetails;
import org.mustangproject.CalculatedInvoice;
import org.mustangproject.TradeParty;
import org.mustangproject.ZUGFeRD.IAbsoluteValueProvider;
import org.mustangproject.ZUGFeRD.IZUGFeRDAllowanceCharge;
import org.mustangproject.ZUGFeRD.IZUGFeRDExportableItem;
import org.mustangproject.ZUGFeRD.IZUGFeRDExportableProduct;
import org.mustangproject.ZUGFeRD.TransactionCalculator;
import org.mustangproject.ZUGFeRD.ZUGFeRDImporter;
import org.mustangproject.util.ByteArraySearcher;
import org.mustangproject.validator.ValidationContext;
import org.mustangproject.validator.ValidationResultItem;
import org.mustangproject.validator.ZUGFeRDValidator;
import org.verapdf.features.FeatureFactory;
import org.verapdf.gf.foundry.VeraGreenfieldFoundryProvider;
import org.verapdf.metadata.fixer.FixerFactory;
import org.verapdf.pdfa.flavours.PDFAFlavour;
import org.verapdf.pdfa.results.ValidationResult;
import org.verapdf.pdfa.validation.validators.ValidatorFactory;
import org.verapdf.processor.ItemProcessor;
import org.verapdf.processor.ProcessorConfig;
import org.verapdf.processor.ProcessorFactory;
import org.verapdf.processor.ProcessorResult;
import org.verapdf.processor.TaskType;
import org.verapdf.processor.plugins.PluginsCollectionConfig;
import org.verapdf.processor.reports.ItemDetails;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.math.BigDecimal;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Calendar;
import java.util.Date;
import java.util.EnumSet;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.logging.Level;
import java.util.logging.Logger;

/**
 * Handles {@code POST /api/inspect}: runs an uploaded PDF through
 * mustangproject's validator and (if present) its ZUGFeRD/Factur-X/XRechnung
 * importer, and shapes the result into this app's own JSON contract.
 *
 * This class calls mustang's public library/validator API directly and does
 * nothing else — no PDF parsing, no XML parsing, no rule logic of its own.
 */
final class InspectHandler {

    private static final Logger LOGGER = Logger.getLogger(InspectHandler.class.getName());
    private static final long MAX_BODY_BYTES = 50L * 1024 * 1024;
    private static final byte[] PDF_SIGNATURE = {'%', 'P', 'D', 'F'};
    private static final ObjectMapper MAPPER = new ObjectMapper()
            .configure(JsonGenerator.Feature.WRITE_BIGDECIMAL_AS_PLAIN, true);

    static void handle(HttpExchange exchange) throws IOException {
        try {
            doHandle(exchange);
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, "unexpected failure inspecting uploaded PDF", e);
            Main.sendJson(exchange, 500, MAPPER.writeValueAsString(new ErrorResponse("Internal error while inspecting the file.")));
        }
    }

    private static void doHandle(HttpExchange exchange) throws IOException {
        String filename = decodeFilenameHeader(exchange.getRequestHeaders().getFirst("X-Filename"));

        byte[] pdfBytes;
        try {
            pdfBytes = readBoundedBody(exchange.getRequestBody());
        } catch (BodyTooLargeException e) {
            Main.sendJson(exchange, 413, MAPPER.writeValueAsString(new ErrorResponse("The uploaded file exceeds the 50 MB size limit.")));
            return;
        }

        boolean valid;
        FormatDto format;
        List<FindingDto> findings;
        try {
            ZUGFeRDValidator validator = new ZUGFeRDValidator();
            validator.validate(pdfBytes, filename);
            ValidationContext context = validator.getContext();
            valid = context.isValid();
            format = new FormatDto(context.getProfile(), context.getGeneration(), context.getFormat());
            findings = new ArrayList<>();
            for (ValidationResultItem item : context.getResults()) {
                findings.add(new FindingDto(
                        item.getID(),
                        item.getSeverity() == null ? null : item.getSeverity().name(),
                        item.getPart() == null ? null : item.getPart().name(),
                        item.getSection(),
                        item.getMessage()));
            }
        } catch (Exception e) {
            LOGGER.log(Level.INFO, "validator could not open uploaded file as a PDF", e);
            Main.sendJson(exchange, 400, MAPPER.writeValueAsString(new ErrorResponse("The uploaded file could not be validated as a PDF.")));
            return;
        }

        String rawXml = null;
        InvoiceDto invoiceDto = null;
        // ZUGFeRDInvoiceImporter falls back to treating any non-"%PDF"-prefixed
        // input as a raw standalone XML invoice (a legitimate use of that class
        // elsewhere in mustang), which would swallow this endpoint's raw bytes
        // into rawXML for non-PDF uploads. This endpoint's contract is PDF-only,
        // so only attempt extraction on input that actually looks like a PDF.
        if (ByteArraySearcher.startsWith(pdfBytes, PDF_SIGNATURE)) {
            try {
                ZUGFeRDImporter importer = new ZUGFeRDImporter(new ByteArrayInputStream(pdfBytes));
                byte[] rawXmlBytes = importer.getRawXML();
                if (rawXmlBytes != null) {
                    rawXml = new String(rawXmlBytes, StandardCharsets.UTF_8);
                    try {
                        CalculatedInvoice invoice = (CalculatedInvoice) importer.extractInto(new CalculatedInvoice());
                        invoiceDto = toInvoiceDto(invoice);
                    } catch (Exception e) {
                        LOGGER.log(Level.INFO, "embedded XML present but could not be parsed into a structured invoice", e);
                    }
                }
            } catch (Exception e) {
                LOGGER.log(Level.INFO, "no embedded e-invoice XML could be extracted from uploaded PDF", e);
            }
        }

        PdfMetadataDto metadata = extractPdfMetadata(pdfBytes, filename);

        InspectResponse response = new InspectResponse(
                filename, pdfBytes.length, valid, format, findings, invoiceDto, rawXml, metadata);
        Main.sendJson(exchange, 200, MAPPER.writeValueAsString(response));
    }

    private static String decodeFilenameHeader(String headerValue) {
        if (headerValue == null || headerValue.isEmpty()) {
            return "upload.pdf";
        }
        try {
            return URLDecoder.decode(headerValue, StandardCharsets.UTF_8);
        } catch (IllegalArgumentException e) {
            return headerValue;
        }
    }

    private static byte[] readBoundedBody(InputStream requestBody) throws IOException, BodyTooLargeException {
        ByteArrayOutputStream buffer = new ByteArrayOutputStream();
        byte[] chunk = new byte[8192];
        long total = 0;
        int read;
        while ((read = requestBody.read(chunk)) != -1) {
            total += read;
            if (total > MAX_BODY_BYTES) {
                throw new BodyTooLargeException();
            }
            buffer.write(chunk, 0, read);
        }
        return buffer.toByteArray();
    }

    private static InvoiceDto toInvoiceDto(CalculatedInvoice invoice) {
        List<LineItemDto> lineItems = new ArrayList<>();
        IZUGFeRDExportableItem[] items = invoice.getZFItems();
        if (items != null) {
            for (IZUGFeRDExportableItem item : items) {
                IZUGFeRDExportableProduct product = item.getProduct();
                String description = product == null ? null : product.getName();
                BigDecimal vatPercent = product == null ? null : product.getVATPercent();
                lineItems.add(new LineItemDto(
                        description, item.getQuantity(), item.getPrice(), vatPercent, item.getLineTotalAmount()));
            }
        }

        TotalsDto totals = new TotalsDto(invoice.getTaxBasis(), invoice.getVATtotal(), invoice.getGrandTotal());

        // getTaxBasis()/getVATtotal()/getGrandTotal() above trigger
        // CalculatedInvoice.calculate() if it hasn't run yet, so
        // getCalculation() is guaranteed non-null here. It's the
        // IAbsoluteValueProvider mustang's own XML writers pass to
        // IZUGFeRDAllowanceCharge.getTotalAmount() to resolve
        // percentage-based allowances/charges against the line total —
        // reusing it here instead of redoing that math ourselves.
        TransactionCalculator basis = invoice.getCalculation();
        List<AllowanceChargeDto> allowances = toAllowanceChargeDtos(invoice.getZFAllowances(), basis, false);
        List<AllowanceChargeDto> charges = toAllowanceChargeDtos(invoice.getZFCharges(), basis, true);

        List<String> notes = invoice.getNotes() == null ? List.of() : Arrays.asList(invoice.getNotes());

        TradeParty paymentReceiver = invoice.getPayee() != null ? invoice.getPayee() : invoice.getSender();
        PaymentMeansDto paymentMeans = toPaymentMeansDto(paymentReceiver);

        return new InvoiceDto(
                invoice.getNumber(),
                formatDate(invoice.getIssueDate()),
                formatDate(invoice.getDueDate()),
                formatDate(invoice.getDeliveryDate()),
                invoice.getCurrency(),
                invoice.getPaymentTermDescription(),
                invoice.getReferenceNumber(),
                invoice.getPaymentReference(),
                toPartyDto(invoice.getSender()),
                toPartyDto(invoice.getRecipient()),
                lineItems,
                totals,
                allowances,
                charges,
                notes,
                paymentMeans);
    }

    private static List<AllowanceChargeDto> toAllowanceChargeDtos(
            IZUGFeRDAllowanceCharge[] items, IAbsoluteValueProvider basis, boolean isCharge) {
        if (items == null) {
            return List.of();
        }
        List<AllowanceChargeDto> result = new ArrayList<>();
        for (IZUGFeRDAllowanceCharge item : items) {
            // getTotalAmount() returns the allowance/charge's own pre-parsed
            // ActualAmount when the source XML stated one directly, without
            // even touching `basis` — only a percent-only allowance/charge
            // needs `basis` to resolve an absolute amount, and that's null
            // for imported invoices whose totals came straight from the XML
            // rather than mustang's own recalculation (see toInvoiceDto's
            // comment on getCalculation()). Degrade to null rather than
            // force a recalculation that could diverge from what the
            // document itself states.
            BigDecimal amount;
            try {
                amount = item.getTotalAmount(basis);
            } catch (RuntimeException e) {
                amount = null;
            }
            result.add(new AllowanceChargeDto(
                    isCharge, item.getReasonCode(), item.getReason(), item.getPercent(),
                    item.getBasisAmount(), amount, item.getTaxCategoryCode(), item.getTaxRateApplicablePercent()));
        }
        return result;
    }

    private static PaymentMeansDto toPaymentMeansDto(TradeParty party) {
        if (party == null) {
            return null;
        }
        List<BankDetails> details = party.getBankDetails();
        if (details == null || details.isEmpty()) {
            return null;
        }
        BankDetails bd = details.get(0);
        if (bd.getIBAN() == null && bd.getBIC() == null) {
            return null;
        }
        return new PaymentMeansDto(bd.getIBAN(), bd.getBIC(), bd.getAccountName());
    }

    private static PartyDto toPartyDto(TradeParty party) {
        if (party == null) {
            return null;
        }
        return new PartyDto(
                party.getName(), party.getStreet(), party.getZIP(), party.getLocation(),
                party.getCountry(), party.getVATID());
    }

    private static String formatDate(Date date) {
        if (date == null) {
            return null;
        }
        return new SimpleDateFormat("yyyy-MM-dd").format(date);
    }

    private static String formatDate(Calendar calendar) {
        return calendar == null ? null : formatDate(calendar.getTime());
    }

    private static PdfMetadataDto extractPdfMetadata(byte[] pdfBytes, String filename) {
        try {
            return doExtractPdfMetadata(pdfBytes, filename);
        } catch (Exception e) {
            LOGGER.log(Level.INFO, "could not extract PDF metadata from uploaded file", e);
            return null;
        }
    }

    private static PdfMetadataDto doExtractPdfMetadata(byte[] pdfBytes, String filename) throws IOException {
        int pageCount;
        String pdfVersion;
        boolean encrypted;
        String producer;
        String creator;
        String creationDate;
        boolean hasXmpMetadata;
        List<String> embeddedFiles = new ArrayList<>();

        try (PDDocument doc = Loader.loadPDF(pdfBytes)) {
            pageCount = doc.getNumberOfPages();
            pdfVersion = String.valueOf(doc.getVersion());
            encrypted = doc.isEncrypted();

            PDDocumentInformation info = doc.getDocumentInformation();
            producer = info == null ? null : info.getProducer();
            creator = info == null ? null : info.getCreator();
            creationDate = info == null ? null : formatDate(info.getCreationDate());

            PDMetadata xmp = doc.getDocumentCatalog().getMetadata();
            hasXmpMetadata = xmp != null;

            PDEmbeddedFilesNameTreeNode embeddedTree =
                    new PDDocumentNameDictionary(doc.getDocumentCatalog()).getEmbeddedFiles();
            if (embeddedTree != null) {
                Map<String, PDComplexFileSpecification> names = embeddedTree.getNames();
                if (names != null) {
                    embeddedFiles.addAll(names.keySet());
                }
            }
        }

        // mustang's own ZUGFeRDValidator already runs this exact veraPDF pass
        // internally (PDFValidator), but ValidationContext.clear() discards the
        // typed PDFAFlavour/compliance result before validate() returns — only a
        // pass/fail ValidationResultItem survives. veraPDF is the same library
        // mustang delegates PDF/A conformance checking to, so calling it directly
        // here (at the cost of running PDF/A validation twice per request) reads
        // the typed result rather than reimplementing PDF/A logic ourselves.
        VeraGreenfieldFoundryProvider.initialise();
        ProcessorConfig processorConfig = ProcessorFactory.fromValues(
                ValidatorFactory.defaultConfig(), FeatureFactory.defaultConfig(),
                PluginsCollectionConfig.defaultConfig(), FixerFactory.defaultConfig(),
                EnumSet.of(TaskType.VALIDATE));

        String pdfaFlavour = null;
        boolean pdfaCompliant = false;
        try (ItemProcessor processor = ProcessorFactory.createProcessor(processorConfig)) {
            ProcessorResult result = processor.process(
                    ItemDetails.fromValues(filename), new ByteArrayInputStream(pdfBytes));
            List<ValidationResult> results = result.getValidationResults();
            if (!results.isEmpty()) {
                ValidationResult vr = results.get(0);
                pdfaCompliant = vr.isCompliant();
                PDFAFlavour flavour = vr.getPDFAFlavour();
                if (flavour != null && flavour != PDFAFlavour.NO_FLAVOUR) {
                    Integer partNumber = flavour.getPart().getPartNumber();
                    String levelCode = flavour.getLevel().getCode();
                    if (partNumber != null) {
                        pdfaFlavour = "PDF/A-" + partNumber
                                + (levelCode == null ? "" : levelCode.toUpperCase(Locale.ROOT));
                    }
                }
            }
        }

        return new PdfMetadataDto(pageCount, pdfVersion, encrypted, producer, creator, creationDate,
                hasXmpMetadata, embeddedFiles, pdfaFlavour, pdfaCompliant);
    }

    private static final class BodyTooLargeException extends Exception {
    }

    record ErrorResponse(String error) {
    }

    record FindingDto(String id, String severity, String part, int section, String message) {
    }

    record PartyDto(String name, String street, String zip, String city, String country, String vatId) {
    }

    record LineItemDto(String description, BigDecimal quantity, BigDecimal unitPrice, BigDecimal vatPercent,
                        BigDecimal lineTotal) {
    }

    record TotalsDto(BigDecimal netTotal, BigDecimal vatTotal, BigDecimal grossTotal) {
    }

    record AllowanceChargeDto(boolean charge, String reasonCode, String reason, BigDecimal percent,
                               BigDecimal basisAmount, BigDecimal amount, String taxCategoryCode,
                               BigDecimal taxRatePercent) {
    }

    record PaymentMeansDto(String iban, String bic, String accountName) {
    }

    record InvoiceDto(String number, String issueDate, String dueDate, String deliveryDate, String currency,
                       String paymentTerms, String buyerReference, String paymentReference, PartyDto seller,
                       PartyDto buyer, List<LineItemDto> lineItems, TotalsDto totals,
                       List<AllowanceChargeDto> allowances, List<AllowanceChargeDto> charges, List<String> notes,
                       PaymentMeansDto paymentMeans) {
    }

    record PdfMetadataDto(int pageCount, String pdfVersion, boolean encrypted, String producer, String creator,
                           String creationDate, boolean hasXmpMetadata, List<String> embeddedFiles,
                           String pdfaFlavour, boolean pdfaCompliant) {
    }

    /**
     * The three identifiers mustang's validator derives from the embedded XML
     * itself: {@code specificationId} is the verbatim
     * GuidelineSpecifiedDocumentContextParameter/ID URN (the one thing that
     * actually names the standard and profile — e.g.
     * {@code urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic}),
     * {@code generation} is the ZUGFeRD generation ("1"/"2"), {@code syntax} is
     * "CII" or "UBL". Naming them is the frontend's job, not this module's.
     */
    record FormatDto(String specificationId, String generation, String syntax) {
    }

    record InspectResponse(String filename, long sizeBytes, boolean valid, FormatDto format,
                            List<FindingDto> findings, InvoiceDto invoice, String rawXml, PdfMetadataDto metadata) {
    }

    private InspectHandler() {
    }
}
