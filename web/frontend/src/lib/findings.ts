import type { Part, Severity } from './types';

/**
 * Buckets the five wire severities into the three tiers the design's
 * summary strip and stripe colors distinguish: an all-clear ("notice"
 * findings are informational, not failures), a warning, or a failure
 * (error/fatal/exception are all "the document is broken," just at
 * different layers of mustang's pipeline).
 */
export type SeverityTier = 'notice' | 'warning' | 'critical';

export function severityTier(severity: Severity): SeverityTier {
  switch (severity) {
    case 'notice':
      return 'notice';
    case 'warning':
      return 'warning';
    default:
      return 'critical';
  }
}

export const severityLabel: Record<Severity, string> = {
  notice: 'Notice',
  warning: 'Warning',
  error: 'Error',
  fatal: 'Fatal',
  exception: 'Exception',
};

// Best-effort labels for mustangproject's per-part breakdown. `pdf` and
// `xr` are documented in the contract; `fx`/`ox` are inferred from
// mustangproject's own profile naming (Factur-X / Order-X) since the
// contract only says "others are XML-related" — flagged as an
// assumption in the handoff notes.
export const partLabel: Partial<Record<Part, string>> = {
  fx: 'Factur-X',
  ox: 'Order-X',
  xr: 'XRechnung',
  pdf: 'PDF/A',
};
