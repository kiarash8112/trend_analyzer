import { Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ReportSummary } from '../../../core/models/report';
import { parseDay } from '../../../core/utils/dates';

/** Keycap tabs, one per saved report; the current report is pressed. */
@Component({
  selector: 'app-date-tabs',
  imports: [RouterLink],
  templateUrl: './date-tabs.html',
  styleUrl: './date-tabs.css',
})
export class DateTabs {
  readonly reports = input.required<ReportSummary[]>();
  readonly current = input<string>();

  protected label(date: string): string {
    const t = parseDay(date);
    return t === null
      ? date
      : new Date(t).toLocaleString('en', {
          weekday: 'short',
          day: 'numeric',
          month: 'short',
          timeZone: 'UTC',
        });
  }
}
