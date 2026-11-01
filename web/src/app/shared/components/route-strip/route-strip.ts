import {
  Component,
  DestroyRef,
  ElementRef,
  afterNextRender,
  computed,
  inject,
  input,
  linkedSignal,
  signal,
  viewChild,
} from '@angular/core';
import { HostPipe } from '../../../core/pipes/host.pipe';
import { RouteStripLayout } from '../../../core/utils/route-layout';

/** Movement before a press counts as dragging along the route instead of a tap. */
const DRAG_THRESHOLD_PX = 4;

/** Remembers across repos and visits that the reader has moved along a route. */
const LEARNED_KEY = 'route-strip-learned';
/** How much of the route the demonstration travels, as a share of its width. */
const DEMO_REACH = 0.3;
const DEMO_DELAY_MS = 400;
const DEMO_DURATION_MS = 1500;
const DEMO_ROUNDS = 2;
/** Most of the route has to be on screen before it demonstrates itself. */
const DEMO_VISIBLE_SHARE = 0.75;

let nextClipId = 0;
let learnedThisVisit = false;

function hasLearned(): boolean {
  try {
    return learnedThisVisit || localStorage.getItem(LEARNED_KEY) !== null;
  } catch {
    return learnedThisVisit;
  }
}

function rememberLearned(): void {
  learnedThisVisit = true;
  try {
    localStorage.setItem(LEARNED_KEY, '1');
  } catch {
    // Without storage the hint only comes back on the next visit.
  }
}

/**
 * The route on a phone: pinned, to scale, and read one event at a time. Earlier and Later move
 * through time, so they stay apart from the slide's round arrows, which change repo.
 */
@Component({
  selector: 'app-route-strip',
  imports: [HostPipe],
  templateUrl: './route-strip.html',
  styleUrl: './route-strip.css',
})
export class RouteStrip {
  readonly layout = input.required<RouteStripLayout>();

  protected readonly clipId = `route-strip-past-${nextClipId++}`;
  protected readonly index = linkedSignal<RouteStripLayout, number>({
    source: this.layout,
    computation: () => 0,
  });
  protected readonly count = computed(() => this.layout().stations.length);
  protected readonly current = computed(() => this.layout().stations[this.index()]);
  protected readonly next = computed(() => this.layout().stations[this.index() + 1]);

  /** The hint and the moving ghost show until the reader first moves along a route. */
  protected readonly learned = signal(hasLearned());
  /** True while the route shows it can move, which also nudges the Later chevron. */
  protected readonly demonstrating = signal(false);

  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);
  private readonly ghost = viewChild<ElementRef<SVGCircleElement>>('ghost');
  private demo: Animation | null = null;

  private pressX: number | null = null;
  private dragging = false;

  constructor() {
    const destroyRef = inject(DestroyRef);
    afterNextRender(() => {
      if (this.learned() || this.count() < 2) return;
      // A hidden route never intersects, so this only fires on phones, each time it comes into view.
      const observer = new IntersectionObserver(
        ([entry]) => {
          if (entry.isIntersecting) this.demonstrate();
        },
        { threshold: DEMO_VISIBLE_SHARE },
      );
      observer.observe(this.host.nativeElement.querySelector('.pinned')!);
      destroyRef.onDestroy(() => observer.disconnect());
    });
    destroyRef.onDestroy(() => this.demo?.cancel());
  }

  protected go(index: number): void {
    this.moveTo(Math.max(0, Math.min(this.count() - 1, index)));
  }

  protected onKey(event: KeyboardEvent): void {
    const target =
      event.key === 'ArrowLeft'
        ? this.index() - 1
        : event.key === 'ArrowRight'
          ? this.index() + 1
          : event.key === 'Home'
            ? 0
            : event.key === 'End'
              ? this.count() - 1
              : null;
    if (target === null) return;
    // On the route these keys move events; the slide page would otherwise change repo.
    event.preventDefault();
    event.stopPropagation();
    this.go(target);
  }

  protected onPress(event: PointerEvent): void {
    this.pressX = event.clientX;
    this.dragging = false;
    (event.currentTarget as Element).setPointerCapture(event.pointerId);
  }

  protected onMove(event: PointerEvent): void {
    if (this.pressX === null) return;
    if (Math.abs(event.clientX - this.pressX) >= DRAG_THRESHOLD_PX) this.dragging = true;
    if (this.dragging) this.pick(event);
  }

  /** A press only selects once it is a tap or a sideways drag, so scrolling the page never does. */
  protected onRelease(event: PointerEvent): void {
    if (this.pressX !== null) this.pick(event);
    this.pressX = null;
  }

  protected onCancel(): void {
    this.pressX = null;
  }

  private pick(event: PointerEvent): void {
    const box = (event.currentTarget as Element).getBoundingClientRect();
    const { stations, width } = this.layout();
    const x = ((event.clientX - box.left) / box.width) * width;
    let nearest = 0;
    stations.forEach((s, i) => {
      if (Math.abs(s.x - x) < Math.abs(stations[nearest].x - x)) nearest = i;
    });
    this.moveTo(nearest);
  }

  /** Only a real move teaches anything, so tapping the selected stop keeps the hint. */
  private moveTo(index: number): void {
    if (index === this.index()) return;
    this.index.set(index);
    this.stopDemonstration();
    if (this.learned()) return;
    this.learned.set(true);
    rememberLearned();
  }

  private stopDemonstration(): void {
    this.demo?.cancel();
    this.demonstrating.set(false);
  }

  /** A ghost of the selected stop runs along the line, showing that the route can be moved. */
  private demonstrate(): void {
    this.stopDemonstration();
    const ghost = this.ghost()?.nativeElement;
    if (!ghost || this.learned() || matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    const { stations, width } = this.layout();
    const from = stations[0];
    const reach = from.x + width * DEMO_REACH;
    const lastStop = stations.findIndex((s, i) => i > 0 && s.x >= reach);
    const stops = stations.slice(0, lastStop === -1 ? undefined : lastStop + 1);
    const distance = stops.at(-1)!.x - from.x || 1;
    const place = (s: { x: number; y: number }) => `translate(${s.x}px, ${s.y}px)`;

    this.demonstrating.set(true);
    this.demo = ghost.animate(
      [
        { offset: 0, opacity: 0, transform: place(from) },
        { offset: 0.1, opacity: 0.75, transform: place(from) },
        ...stops.slice(1).map((s) => ({
          offset: 0.1 + ((s.x - from.x) / distance) * 0.65,
          opacity: 0.75,
          transform: place(s),
        })),
        { offset: 1, opacity: 0, transform: place(stops.at(-1)!) },
      ],
      {
        delay: DEMO_DELAY_MS,
        duration: DEMO_DURATION_MS,
        iterations: DEMO_ROUNDS,
        easing: 'ease-in-out',
      },
    );
    this.demo.onfinish = () => this.demonstrating.set(false);
  }
}
