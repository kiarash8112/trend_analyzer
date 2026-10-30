const LEADING_NUMBER = /^\s*\(?\d{1,2}[.)]\s+/;

export interface Reasons {
  intro: string;
  reasons: string[];
}

/**
 * Splits why_popular into an intro and a list of reasons. New reports store a list (items
 * may still start with "1) "); older ones store one string with inline numbering.
 */
export function splitReasons(why: string[] | string): Reasons {
  if (Array.isArray(why)) {
    return { intro: '', reasons: why.map((r) => r.replace(LEADING_NUMBER, '')) };
  }
  // Keep only markers numbered 1, 2, 3... in order, so dates like "Oct 26. " are skipped.
  const marks: RegExpExecArray[] = [];
  for (const m of why.matchAll(/(?<=^|\s)\(?(\d{1,2})[.)]\s+/g)) {
    if (Number(m[1]) === marks.length + 1) {
      marks.push(m);
    }
  }
  if (marks.length < 2) {
    return { intro: why, reasons: [] };
  }
  const reasons = marks.map((m, i) => {
    const end = i + 1 < marks.length ? marks[i + 1].index : why.length;
    return why.slice(m.index + m[0].length, end).trim();
  });
  return { intro: why.slice(0, marks[0].index).trim(), reasons };
}
