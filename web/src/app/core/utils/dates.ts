export const DAY_MS = 86_400_000;

/** Parses "YYYY-MM-DD" or "YYYY-MM" (mid-month) to a UTC timestamp; null when it is not a date. */
export function parseDay(value: string | undefined): number | null {
  const m = /^(\d{4})-(\d{2})(?:-(\d{2}))?/.exec(value ?? '');
  return m ? Date.UTC(+m[1], +m[2] - 1, m[3] ? +m[3] : 15) : null;
}

export function isoDay(time: number): string {
  return new Date(time).toISOString().slice(0, 10);
}

/** "Aug 31, 2025", or "Oct 2025" when only the month is known; "~" marks an approximate date. */
export function dayLabel(value: string): string {
  const m = /^(\d{4})-(\d{2})(?:-(\d{2}))?/.exec(value);
  if (!m) return value;
  const month = new Date(Date.UTC(2000, +m[2] - 1, 1)).toLocaleString('en', {
    month: 'short',
    timeZone: 'UTC',
  });
  const approx = value.includes('approx') ? '~' : '';
  return m[3] ? `${approx}${month} ${+m[3]}, ${m[1]}` : `${approx}${month} ${m[1]}`;
}

/** A length of time in the largest unit that reads naturally: "9 days", "3 weeks", "1 yr 2 mo". */
export function spanLabel(ms: number): string {
  const days = Math.round(ms / DAY_MS);
  if (days < 14) return plural(days, 'day');
  if (days < 45) return plural(Math.round(days / 7), 'week');
  const months = Math.round(days / 30.44);
  if (months < 12) return plural(months, 'month');
  const years = Math.floor(months / 12);
  return months % 12 ? `${years} yr ${months % 12} mo` : plural(years, 'year');
}

function plural(n: number, unit: string): string {
  return `${n} ${unit}${n === 1 ? '' : 's'}`;
}
