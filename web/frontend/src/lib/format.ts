export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const kb = bytes / 1024;
  if (kb < 1024) return `${kb.toFixed(kb < 10 ? 1 : 0)} KB`;
  const mb = kb / 1024;
  return `${mb.toFixed(mb < 10 ? 1 : 0)} MB`;
}

const currencyCache = new Map<string, Intl.NumberFormat>();

export function formatMoney(value: number, currency: string | null | undefined): string {
  const code = currency || 'EUR';
  let fmt = currencyCache.get(code);
  if (!fmt) {
    try {
      fmt = new Intl.NumberFormat('en-US', { style: 'currency', currency: code });
    } catch {
      fmt = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'EUR' });
    }
    currencyCache.set(code, fmt);
  }
  return fmt.format(value);
}

export function formatDate(iso: string | null | undefined): string {
  if (!iso) return '—';
  return iso;
}
