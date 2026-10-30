import { Insight } from '../models/report';
import { DAY_MS, isoDay, parseDay } from './dates';
import { truncate } from './text';

export interface RouteStop {
  x: number;
  y: number;
  /** Width of an interchange pill when several events fall on nearly the same day; 0 for a station. */
  span: number;
  now: boolean;
}

export interface RouteLabel {
  x: number;
  y: number;
  anchor: 'start' | 'end';
  text: string;
  full: string;
  date: string;
  /** Connector from the label to its station. */
  y1: number;
  y2: number;
}

export interface RouteMonth {
  x: number;
  label: string;
}

export interface RouteLayout {
  width: number;
  height: number;
  path: string;
  stops: RouteStop[];
  labels: RouteLabel[];
  months: RouteMonth[];
  start: string;
  events: { date: string; event: string }[];
}

const WIDTH = 1100;
const LEVELS = [128, 96, 64];
const STEP = 32;
const LABEL_TOP = 22;
const LABEL_BOTTOM = 170;
const CHAR_WIDTH = 6.6;
const SURGE_DAYS = 30;

/**
 * Lays a repo's timeline out as a metro route spaced by real dates. The line steps up
 * only at real phase changes: the launch and the start of the current trending window.
 */
export function routeLayout(insight: Insight, reportDay: number): RouteLayout | null {
  const events = insight.timeline
    .map((e) => ({ ...e, t: parseDay(e.date) }))
    .filter((e): e is typeof e & { t: number } => e.t !== null)
    .sort((a, b) => a.t - b.t);
  if (!events.length) {
    return null;
  }

  const launch = parseDay(insight.launch_date);
  const start = Math.min(events[0].t, launch ?? events[0].t);
  const end = reportDay;
  const span = Math.max(end - start, SURGE_DAYS * DAY_MS);
  const x = (t: number) => 10 + ((t - start) / span) * (WIDTH - 30);
  const surge = end - SURGE_DAYS * DAY_MS;

  const steps: number[] = [];
  if (launch !== null && launch > start) steps.push(x(launch));
  if (surge > start) steps.push(x(surge));
  const baseLevel = launch !== null && start < launch ? 0 : 1;

  let level = baseLevel;
  let path = `M${x(start)} ${LEVELS[level]}`;
  for (const sx of steps) {
    path += ` H${Math.max(sx - STEP, x(start))} L${sx} ${LEVELS[level + 1]}`;
    level++;
  }
  path += ` H${x(end)}`;

  const yAt = (px: number) => {
    let lv = baseLevel;
    let y = LEVELS[lv];
    for (const sx of steps) {
      if (px >= sx) y = LEVELS[++lv];
      else if (px > sx - STEP) y = LEVELS[lv] - (px - (sx - STEP));
    }
    return y;
  };

  // Events closer than a station's width share one interchange.
  const groups: (typeof events)[] = [];
  for (const e of events) {
    const g = groups.at(-1);
    if (g && x(e.t) - x(g[0].t) < 14) g.push(e);
    else groups.push([e]);
  }

  const stops: RouteStop[] = [];
  const labels: RouteLabel[] = [];
  const lastRight = { up: -Infinity, down: -Infinity };
  groups.forEach((g, i) => {
    const sx = x(g[0].t);
    const sy = yAt(sx);
    stops.push({
      x: sx,
      y: sy,
      span: g.length > 1 ? x(g.at(-1)!.t) - sx : 0,
      now: i === groups.length - 1,
    });

    const full = g.map((e) => e.event).join(' · ');
    const text = truncate(full, 34);
    const w = Math.max(text.length * CHAR_WIDTH, 90);
    const anchor = sx + w > WIDTH ? 'end' : 'start';
    const left = anchor === 'end' ? sx - w : sx;
    const side = left > lastRight.up + 12 ? 'up' : left > lastRight.down + 12 ? 'down' : null;
    if (!side) return;
    lastRight[side] = anchor === 'end' ? sx : sx + w;
    const y = side === 'up' ? LABEL_TOP : LABEL_BOTTOM;
    labels.push({
      x: sx,
      y,
      anchor,
      text,
      full,
      date: g[0].date,
      y1: side === 'up' ? y + 20 : sy + 10,
      y2: side === 'up' ? sy - 10 : y - 14,
    });
  });

  const months: RouteMonth[] = [];
  const first = new Date(start);
  for (let y = first.getUTCFullYear(), m = first.getUTCMonth() + 1; ; m++) {
    const t = Date.UTC(y, m, 1);
    if (t > end) break;
    const d = new Date(t);
    const name = d.toLocaleString('en', { month: 'short', timeZone: 'UTC' }).toLowerCase();
    months.push({ x: x(t), label: d.getUTCMonth() === 0 ? `${name} ${d.getUTCFullYear()}` : name });
  }

  return {
    width: WIDTH,
    height: 218,
    path,
    stops,
    labels,
    months,
    start: isoDay(start),
    events: events.map(({ date, event }) => ({ date, event })),
  };
}
