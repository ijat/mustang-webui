package dev.mustangwebui.sidecar;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpExchange;
import org.mustangproject.CalculatedInvoice;
import org.mustangproject.TradeParty;
import org.mustangproject.ZUGFeRD.IZUGFeRDExportableItem;
import org.mustangproject.ZUGFeRD.IZUGFeRDExportableProduct;
import org.mustangproject.ZUGFeRD.ZUGFeRDImporter;
import org.mustangproject.util.ByteArraySearcher;
import org.mustangproject.validator.ValidationContext;
import org.mustangproject.validator.ValidationResultItem;
import org.mustangproject.validator.ZUGFeRDValidator;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.math.BigDecimal;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
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
        String profile;
        List<FindingDto> findings;
        try {
            ZUGFeRDValidator validator = new ZUGFeRDValidator();
            validator.validate(pdfBytes, filename);
            ValidationContext context = validator.getContext();
            valid = context.isValid();
            profile = context.getProfile();
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

        InspectResponse response = new InspectResponse(
                filename, pdfBytes.length, valid, profile, findings, invoiceDto, rawXml);
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

        return new InvoiceDto(
                invoice.getNumber(),
                formatDate(invoice.getIssueDate()),
                formatDate(invoice.getDueDate()),
                invoice.getCurrency(),
                invoice.getPaymentTermDescription(),
                toPartyDto(invoice.getSender()),
                toPartyDto(invoice.getRecipient()),
                lineItems,
                totals);
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

    record InvoiceDto(String number, String issueDate, String dueDate, String currency, String paymentTerms,
                       PartyDto seller, PartyDto buyer, List<LineItemDto> lineItems, TotalsDto totals) {
    }

    record InspectResponse(String filename, long sizeBytes, boolean valid, String profile,
                            List<FindingDto> findings, InvoiceDto invoice, String rawXml) {
    }

    private InspectHandler() {
    }
}
