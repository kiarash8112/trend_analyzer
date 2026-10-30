import { Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';

/** One segment per repo of the day; each links to its slide, and the current one fills as you read. */
@Component({
  selector: 'app-progress-segments',
  imports: [RouterLink],
  templateUrl: './progress-segments.html',
  styleUrl: './progress-segments.css',
})
export class ProgressSegments {
  readonly count = input.required<number>();
  /** 1-based position of the current slide. */
  readonly current = input.required<number>();
  /** Builds the router link for a 1-based position. */
  readonly linkFor = input.required<(position: number) => unknown[]>();
  readonly labels = input<string[]>([]);
  /** How far through the current position the reader is, 0 to 1. */
  readonly progress = input(0);

  protected readonly positions = computed(() =>
    Array.from({ length: this.count() }, (_, i) => i + 1),
  );
}
