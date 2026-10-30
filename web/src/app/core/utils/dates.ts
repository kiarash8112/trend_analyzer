export const DAY_MS = 86_400_000;

/** Parses "YYYY-MM-DD" or "YYYY-MM" (mid-month) to a UTC timestamp; null when it is not a date. */
export function parseDay(value: string | undefined): number | null {
  const m = /^(\d{4})-(\d{2})(?:-(\d{2}))?/.exec(value ?? '');
  return m ? Date.UTC(+m[1], +m[2] - 1, m[3] ? +m[3] : 15) : null;
}

export function isoDay(time: number): string {
  return new Date(time).toISOString().slice(0, 10);
}
