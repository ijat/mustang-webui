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

export interface Invoice {
  number: string;
  issueDate: string;
  dueDate: string | null;
  currency: string;
  paymentTerms: string | null;
  seller: Party;
  buyer: Party;
  lineItems: LineItem[];
  totals: Totals;
}

export interface InspectResponse {
  filename: string;
  sizeBytes: number;
  valid: boolean;
  profile: string | null;
  findings: Finding[];
  invoice: Invoice | null;
  rawXml: string | null;
}

export interface ApiError {
  error: string;
}
