import { Component, computed, inject, input } from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { Router } from '@angular/router';
import { RepoResult } from '../../../core/models/report';
import { splitReasons } from '../../../core/utils/reasons';
import { shortChannel } from '../../../core/utils/text';
import { LineRoundel } from '../../../shared/components/line-roundel/line-roundel';

/** One repo in the timetable; clicking the row (or its roundel) opens the repo's slide. */
@Component({
  selector: 'app-timetable-row',
  imports: [DecimalPipe, LineRoundel],
  templateUrl: './timetable-row.html',
  styleUrl: './timetable-row.css',
})
export class TimetableRow {
  readonly result = input.required<RepoResult>();
  readonly rank = input.required<number>();
  /** Router link to this repo's slide. */
  readonly link = input.required<unknown[]>();

  private readonly router = inject(Router);

  protected readonly repo = computed(() => this.result().evidence.repo);
  protected readonly insight = computed(() => this.result().insight);
  protected readonly channel = computed(() => shortChannel(this.insight()?.launch_channel ?? ''));
  protected readonly lead = computed(() => {
    const i = this.insight();
    if (!i) {
      return `Analysis failed: ${this.result().error}`;
    }
    // The repo's own description; repos without one fall back to the first reason.
    const why = splitReasons(i.why_popular);
    return this.repo().description || why.reasons[0] || why.intro;
  });

  protected onRowClick(event: MouseEvent): void {
    if (!this.insight() || (event.target as HTMLElement).closest('a')) {
      return;
    }
    this.router.navigate(this.link());
  }
}
