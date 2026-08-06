function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

/**
 * Minimal regex-based XML syntax coloring for display only — not a
 * parser, and deliberately not one: this only ever needs to color text
 * mustangproject already validated as well-formed. Falls back to plain
 * escaped text on anything that doesn't look like a tag/attribute so a
 * malformed fragment never crashes the view.
 */
export function highlightXml(xml: string): string {
  const escaped = escapeHtml(xml);

  return escaped.replace(
    /(&lt;!--[\s\S]*?--&gt;)|(&lt;\/?[\w:.-]+)|([\w:.-]+)(=)("[^"]*")|(\/?&gt;)/g,
    (match, comment, tagOpen, attrName, attrEq, attrVal, tagClose) => {
      if (comment) return `<span class="xml-muted">${comment}</span>`;
      if (tagOpen) return `<span class="xml-tag">${tagOpen}</span>`;
      if (attrName && attrEq && attrVal) {
        return `<span class="xml-attr">${attrName}</span>${attrEq}<span class="xml-val">${attrVal}</span>`;
      }
      if (tagClose) return `<span class="xml-tag">${tagClose}</span>`;
      return match;
    },
  );
}
