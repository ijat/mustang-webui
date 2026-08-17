import type { InspectResponse } from './types';

/**
 * Names a document's standard from the identifiers mustangproject reports.
 *
 * The specification identifier
 * (GuidelineSpecifiedDocumentContextParameter/ID in CII) is the only
 * authoritative statement of *what* the embedded XML is — everything below
 * reads that one URN. The attachment filename is a separate, weaker signal
 * that matters for exactly one distinction: Factur-X and ZUGFeRD 2.x are the
 * same standard and carry byte-identical specification identifiers, so the
 * filename its issuer chose is the only in-document hint of which of the two
 * names they were working under.
 */
export interface FormatDescription {
  standard: string;
  profileLevel: string | null;
  jurisdiction: string | null;
  syntax: string | null;
  attachment: AttachmentHint | null;
  specificationId: string | null;
  en16931: En16931Status;
  recognised: boolean;
}

/**
 * Whether the document claims conformity with EN 16931 — the practical gate
 * for the German and French B2B mandates, and a stronger signal than the
 * standard's brand name. It is not a judgement of our own: the specification
 * identifier either carries the EN 16931 URN or it doesn't. MINIMUM and
 * BASIC WL notably don't, since neither is a complete invoice.
 */
export type En16931Status = 'compliant' | 'extension' | 'no' | 'not-applicable' | 'unknown';

export interface AttachmentHint {
  filename: string;
  naming: string;
}

const ATTACHMENT_NAMING: Record<string, string> = {
  'factur-x.xml': 'Factur-X naming (France)',
  'zugferd-invoice.xml': 'ZUGFeRD naming (Germany)',
  'xrechnung.xml': 'XRechnung naming (Germany)',
  'order-x.xml': 'Order-X naming',
};

export function describeFormat(result: InspectResponse): FormatDescription | null {
  const specificationId = result.format?.specificationId?.trim() || null;
  const attachment = findAttachment(result.metadata?.embeddedFiles ?? []);
  // The syntax defaults to "CII" even for a plain PDF that carries no XML at
  // all, so it can't stand in for "this is an e-invoice" on its own.
  const syntax = result.format?.syntax?.trim() || null;

  if (!specificationId) {
    return result.rawXml || attachment
      ? {
          standard: 'Unidentified e-invoice',
          profileLevel: null,
          jurisdiction: null,
          syntax,
          attachment,
          specificationId: null,
          en16931: 'unknown',
          recognised: false,
        }
      : null;
  }

  const urn = specificationId.toLowerCase();
  const base = {
    profileLevel: profileLevel(urn),
    syntax,
    attachment,
    specificationId,
    en16931: en16931Status(urn),
    recognised: true,
  };

  if (urn.includes('xrechnung')) {
    return {
      ...base,
      standard: `XRechnung${xrechnungVersion(urn) ?? ''}`,
      jurisdiction: 'Germany — public-sector (B2G) CIUS of EN 16931',
    };
  }
  if (urn.includes('order-x.eu')) {
    return {
      ...base,
      standard: 'Order-X 1.0',
      jurisdiction: 'France & Germany — purchase order, not an invoice',
    };
  }
  if (urn.includes('factur-x.eu') || urn === 'urn:cen.eu:en16931:2017') {
    return {
      ...base,
      standard: 'Factur-X 1.0 / ZUGFeRD 2.x',
      jurisdiction: 'France (Factur-X) and Germany (ZUGFeRD) — one identical standard under two names',
    };
  }
  if (urn.includes('zugferd.de:2p0')) {
    return { ...base, standard: 'ZUGFeRD 2.0', jurisdiction: 'Germany' };
  }
  if (urn.includes('urn:ferd:crossindustrydocument')) {
    return {
      ...base,
      standard: 'ZUGFeRD 1.0',
      jurisdiction: 'Germany — predates EN 16931',
    };
  }

  return {
    ...base,
    standard: 'Unrecognised specification',
    jurisdiction: null,
    recognised: false,
  };
}

function findAttachment(embeddedFiles: string[]): AttachmentHint | null {
  for (const filename of embeddedFiles) {
    const naming = ATTACHMENT_NAMING[filename.toLowerCase()];
    if (naming) return { filename, naming };
  }
  return null;
}

function en16931Status(urn: string): En16931Status {
  if (urn.includes('order-x.eu')) return 'not-applicable';
  if (!urn.includes('urn:cen.eu:en16931:2017')) return 'no';
  return urn.includes('conformant') ? 'extension' : 'compliant';
}

function profileLevel(urn: string): string | null {
  if (urn.includes('minimum')) return 'MINIMUM';
  if (urn.includes('basicwl')) return 'BASIC WL';
  if (urn.includes('basic')) return 'BASIC';
  if (urn.includes('extended')) return 'EXTENDED';
  if (urn.includes('comfort')) return 'COMFORT';
  if (urn.includes('en16931')) return 'EN 16931';
  return null;
}

function xrechnungVersion(urn: string): string | null {
  const match = urn.match(/xrechnung_(\d+(?:\.\d+)*)/);
  return match ? ` ${match[1]}` : null;
}
