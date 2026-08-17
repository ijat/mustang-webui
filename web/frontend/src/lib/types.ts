export type Severity = 'notice' | 'warning' | 'error' | 'fatal' | 'exception';

export type Part = 'fx' | 'ox' | 'xr' | 'pdf' | 'none';

export interface Finding {
  id: string;
  severity: Severity;
  part: Part;
  section: number;
  message: string;
}

export interface Party {
  name: string;
  street: string;
  zip: string;
  city: string;
  country: string;
  vatId: string;
}

export interface LineItem {
  description: string;
  quantity: number;
  unitPrice: number;
  vatPercent: number;
  lineTotal: number;
}

export interface Totals {
  netTotal: number;
  vatTotal: number;
  grossTotal: number;
}

export interface AllowanceCharge {
  charge: boolean;
  reasonCode: string | null;
  reason: string | null;
  percent: number | null;
  basisAmount: number | null;
  amount: number | null;
  taxCategoryCode: string | null;
  taxRatePercent: number | null;
}

export interface PaymentMeans {
  iban: string | null;
  bic: string | null;
  accountName: string | null;
}

export interface Invoice {
  number: string;
  issueDate: string;
  dueDate: string | null;
  deliveryDate: string | null;
  currency: string;
  paymentTerms: string | null;
  buyerReference: string | null;
  paymentReference: string | null;
  seller: Party;
  buyer: Party;
  lineItems: LineItem[];
  totals: Totals;
  allowances: AllowanceCharge[];
  charges: AllowanceCharge[];
  notes: string[];
  paymentMeans: PaymentMeans | null;
}

export interface PdfMetadata {
  pageCount: number;
  pdfVersion: string;
  encrypted: boolean;
  producer: string | null;
  creator: string | null;
  creationDate: string | null;
  hasXmpMetadata: boolean;
  embeddedFiles: string[];
  pdfaFlavour: string | null;
  pdfaCompliant: boolean;
}

export interface DocumentFormat {
  specificationId: string | null;
  generation: string | null;
  syntax: string | null;
}

export interface InspectResponse {
  filename: string;
  sizeBytes: number;
  valid: boolean;
  format: DocumentFormat | null;
  findings: Finding[];
  invoice: Invoice | null;
  rawXml: string | null;
  metadata: PdfMetadata | null;
}

export interface ApiError {
  error: string;
}
