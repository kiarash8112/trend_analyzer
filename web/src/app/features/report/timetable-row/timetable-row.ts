import {
  Component,
  ElementRef,
  afterRenderEffect,
  computed,
  inject,
  input,
  output,
} from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { RepoResult } from '../../../core/models/report';
import { splitReasons } from '../../../core/utils/reasons';
import { shortChannel } from '../../../core/utils/text';
import { LineRoundel } from '../../../shared/components/line-roundel/line-roundel';
import { RepoDetail } from '../repo-detail/repo-detail';

/** One repo in the timetable; clicking the row (or its roundel) opens the detail below it. */
@Component({
  selector: 'app-timetable-row',
  imports: [DecimalPipe, LineRoundel, RepoDetail],
  templateUrl: './timetable-row.html',
  styleUrl: './timetable-row.css',
  host: { '[class.open]': 'open()' },
})
export class TimetableRow {
  readonly result = input.required<RepoResult>();
  readonly rank = input.required<number>();
  readonly open = input(false);
  readonly reportDay = input.required<number>();
  readonly toggle = output<void>();

  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);
  private scrollOnOpen = false;

  protected readonly repo = computed(() => this.result().evidence.repo);
  protected readonly insight = computed(() => this.result().insight);
  protected readonly channel = computed(() => shortChannel(this.insight()?.launch_channel ?? ''));
  protected readonly lead = computed(() => {
    const i = this.insight();
    if (!i) {
      return `Analysis failed: ${this.result().error}`;
    }
    const why = splitReasons(i.why_popular);
    return why.reasons[0] ?? why.intro;
  });

  constructor() {
    // Opening a row brings its start to the top, even when the row above just collapsed.
    afterRenderEffect(() => {
      if (this.open() && this.scrollOnOpen) {
        this.scrollOnOpen = false;
        const reduce = matchMedia('(prefers-reduced-motion: reduce)').matches;
        this.host.nativeElement.scrollIntoView({
          block: 'start',
          behavior: reduce ? 'auto' : 'smooth',
        });
      }
    });
  }

  protected onRowClick(event: MouseEvent): void {
    if (!this.insight() || (event.target as HTMLElement).closest('a')) {
      return;
    }
    this.scrollOnOpen = !this.open();
    this.toggle.emit();
  }
}
