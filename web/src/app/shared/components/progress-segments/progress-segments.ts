import { Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';

/** One segment per repo of the day; each segment links to its slide. */
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

  protected readonly positions = computed(() =>
    Array.from({ length: this.count() }, (_, i) => i + 1),
  );
}
