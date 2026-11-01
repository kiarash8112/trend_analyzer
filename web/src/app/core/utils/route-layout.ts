import { Insight } from '../models/report';
import { DAY_MS, dayLabel, isoDay, parseDay, spanLabel } from './dates';

export interface RouteStop {
  x: number;
  y: number;
  /** Width of an interchange pill when several events fall on nearly the same day; 0 for a station. */
  span: number;
  now: boolean;
}

export interface RouteLabel {
  /** Left edge of the label column, as a percentage of the route's width. */
  left: number;
  side: 'above' | 'below';
  /** The column hangs left of its station when it would run past the right edge. */
  alignEnd: boolean;
  text: string;
  date: string;
  /** Connector from the station to the label row, in route units. */
  x: number;
  y1: number;
  y2: number;
}

export interface RouteMonth {
  x: number;
  label: string;
}

export interface RouteLayout {
  /** viewBox x origin and width; the route is drawn from 0 to width - 20. */
  viewX: number;
  viewWidth: number;
  /** Width of one label column, as a percentage of the route's width. */
  labelWidth: number;
  height: number;
  path: string;
  stops: RouteStop[];
  labels: RouteLabel[];
  months: RouteMonth[];
  start: string;
  eventCount: number;
}

/** One event on the phone's strip map, with what the reader sees when it is selected. */
export interface StripStation {
  x: number;
  y: number;
  date: string;
  event: string;
  sourceUrl: string;
  /** How long after the event before it. */
  since: string;
}

/** A repo's route small enough for a phone: the same line, one selectable station per event. */
export interface RouteStripLayout {
  width: number;
  height: number;
  path: string;
  stations: StripStation[];
  months: RouteMonth[];
}

const WIDTH = 1100;
const VIEW_X = -10;
const VIEW_WIDTH = WIDTH + 20;
const HEIGHT = 120;
const LEVELS = [92, 60, 28];
const STEP = 32;
/** A label column in route units; its text wraps inside it, so nothing is cut. */
const LABEL_WIDTH = 210;
const LABEL_GAP = 14;
const SURGE_DAYS = 30;

const STRIP_WIDTH = 358;
const STRIP_HEIGHT = 92;
const STRIP_LEVELS = [62, 44, 26];
const STRIP_STEP = 18;
/** Fewer month labels on longer routes, so they stay apart on a phone. */
const STRIP_QUARTERS_FROM_DAYS = 300;
const STRIP_YEARS_FROM_DAYS = 800;

/** The line shared by both layouts. It steps up at the launch and the current trending window. */
function routeTrack(
  insight: Insight,
  reportDay: number,
  width: number,
  levels: number[],
  stepSize: number,
) {
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
  const x = (t: number) => 10 + ((t - start) / span) * (width - 30);
  const surge = end - SURGE_DAYS * DAY_MS;

  const steps: number[] = [];
  // A launch step needs room to climb; a launch a few days after the first event starts launched.
  const launchStep = launch !== null && x(launch) - x(start) > stepSize * 1.5;
  if (launchStep) steps.push(x(launch));
  if (surge > start) steps.push(x(surge));
  const baseLevel = launchStep ? 0 : 1;

  let level = baseLevel;
  let path = `M${x(start)} ${levels[level]}`;
  for (const sx of steps) {
    path += ` H${Math.max(sx - stepSize, x(start))} L${sx} ${levels[level + 1]}`;
    level++;
  }
  path += ` H${x(end)}`;

  const yAt = (px: number) => {
    let lv = baseLevel;
    let y = levels[lv];
    for (const sx of steps) {
      if (px >= sx) y = levels[++lv];
      else if (px > sx - stepSize) y = levels[lv] - (px - (sx - stepSize));
    }
    return y;
  };

  return { events, start, end, x, yAt, path };
}

/** Timestamps of the first of each month after start, up to end. */
function monthStarts(start: number, end: number): number[] {
  const out: number[] = [];
  const first = new Date(start);
  for (let y = first.getUTCFullYear(), m = first.getUTCMonth() + 1; ; m++) {
    const t = Date.UTC(y, m, 1);
    if (t > end) return out;
    out.push(t);
  }
}

/**
 * Lays a repo's timeline out as a metro route spaced by real dates. The line steps up
 * only at real phase changes: the launch and the start of the current trending window.
 */
export function routeLayout(insight: Insight, reportDay: number): RouteLayout | null {
  const track = routeTrack(insight, reportDay, WIDTH, LEVELS, STEP);
  if (!track) {
    return null;
  }
  const { events, start, end, x, yAt, path } = track;

  // Events closer than a station's width share one interchange.
  const groups: (typeof events)[] = [];
  for (const e of events) {
    const g = groups.at(-1);
    if (g && x(e.t) - x(g[0].t) < 14) g.push(e);
    else groups.push([e]);
  }

  const stops: RouteStop[] = [];
  const labels: RouteLabel[] = [];
  const lastRight = { above: -Infinity, below: -Infinity };
  const pct = (units: number) => ((units - VIEW_X) / VIEW_WIDTH) * 100;
  groups.forEach((g, i) => {
    const sx = x(g[0].t);
    const sy = yAt(sx);
    stops.push({
      x: sx,
      y: sy,
      span: g.length > 1 ? x(g.at(-1)!.t) - sx : 0,
      now: i === groups.length - 1,
    });

    const alignEnd = sx + LABEL_WIDTH > WIDTH;
    const left = alignEnd ? sx - LABEL_WIDTH : sx;
    const side =
      left > lastRight.above + LABEL_GAP
        ? 'above'
        : left > lastRight.below + LABEL_GAP
          ? 'below'
          : null;
    if (!side) return;
    lastRight[side] = left + LABEL_WIDTH;
    labels.push({
      left: pct(left),
      side,
      alignEnd,
      text: g.map((e) => e.event).join(' · '),
      date: g[0].date,
      x: sx,
      y1: side === 'above' ? 0 : sy + 10,
      y2: side === 'above' ? sy - 10 : HEIGHT,
    });
  });

  const months = monthStarts(start, end).map((t) => {
    const d = new Date(t);
    const name = d.toLocaleString('en', { month: 'short', timeZone: 'UTC' }).toLowerCase();
    return { x: x(t), label: d.getUTCMonth() === 0 ? `${name} ${d.getUTCFullYear()}` : name };
  });

  return {
    viewX: VIEW_X,
    viewWidth: VIEW_WIDTH,
    labelWidth: (LABEL_WIDTH / VIEW_WIDTH) * 100,
    height: HEIGHT,
    path,
    stops,
    labels,
    months,
    start: isoDay(start),
    eventCount: events.length,
  };
}

/** Lays the same route out for a phone, where the reader moves through it one event at a time. */
export function routeStripLayout(insight: Insight, reportDay: number): RouteStripLayout | null {
  // The line ends 20 units short of the width passed in, which would waste a phone's narrow route.
  const track = routeTrack(insight, reportDay, STRIP_WIDTH + 10, STRIP_LEVELS, STRIP_STEP);
  if (!track) {
    return null;
  }
  const { events, start, end, x, yAt, path } = track;

  const days = (end - start) / DAY_MS;
  const every = days > STRIP_YEARS_FROM_DAYS ? 12 : days > STRIP_QUARTERS_FROM_DAYS ? 3 : 1;
  const months = monthStarts(start, end)
    .filter((t) => new Date(t).getUTCMonth() % every === 0)
    .map((t) => {
      const d = new Date(t);
      const name = d.toLocaleString('en', { month: 'short', timeZone: 'UTC' });
      return { x: x(t), label: d.getUTCMonth() === 0 ? String(d.getUTCFullYear()) : name };
    });

  return {
    width: STRIP_WIDTH,
    height: STRIP_HEIGHT,
    path,
    months,
    stations: events.map((e, i) => {
      const before = events[i - 1];
      return {
        x: x(e.t),
        y: yAt(x(e.t)),
        date: dayLabel(e.date),
        event: e.event,
        sourceUrl: e.source_url,
        since: !before
          ? 'First event'
          : e.t === before.t
            ? 'Same day as the last event'
            : `${spanLabel(e.t - before.t)} after the last event`,
      };
    }),
  };
}
