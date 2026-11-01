import { Component, computed, input, linkedSignal } from '@angular/core';
import { HostPipe } from '../../../core/pipes/host.pipe';
import { RouteStripLayout } from '../../../core/utils/route-layout';

/** Movement before a press counts as dragging along the route instead of a tap. */
const DRAG_THRESHOLD_PX = 4;

let nextClipId = 0;

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

  private pressX: number | null = null;
  private dragging = false;

  protected go(index: number): void {
    this.index.set(Math.max(0, Math.min(this.count() - 1, index)));
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
    this.index.set(nearest);
  }
}
