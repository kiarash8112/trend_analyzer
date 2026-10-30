import { Component, computed, inject, input, signal } from '@angular/core';
import { ReportApi } from '../../core/services/report-api';
import { parseDay } from '../../core/utils/dates';
import { RankedResult, Timetable } from '../../features/report/timetable/timetable';
import { TopBar } from '../../layout/top-bar/top-bar';
import { SearchField } from '../../shared/components/search-field/search-field';

@Component({
  selector: 'app-report-page',
  imports: [TopBar, SearchField, Timetable],
  templateUrl: './report-page.html',
  styleUrl: './report-page.css',
})
export class ReportPage {
  /** Route param; when absent the newest report is shown. */
  readonly name = input<string>();

  private readonly api = inject(ReportApi);

  protected readonly reports = computed(() => this.api.list.value() ?? []);
  protected readonly listLoading = this.api.list.isLoading;
  protected readonly listError = this.api.list.error;
  protected readonly current = computed(() => this.name() ?? this.reports()[0]?.name);
  protected readonly summary = computed(() =>
    this.reports().find((r) => r.name === this.current()),
  );
  protected readonly report = this.api.report(this.current);

  protected readonly reportDay = computed(
    () => parseDay(this.summary()?.date ?? this.current()) ?? Date.now(),
  );
  protected readonly title = computed(() =>
    new Date(this.reportDay()).toLocaleString('en', {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      timeZone: 'UTC',
    }),
  );
  protected readonly query = signal('');

  protected readonly rows = computed<RankedResult[]>(() => {
    const all = (this.report.value() ?? []).map((result, i) => ({ rank: i + 1, result }));
    const q = this.query().trim().toLowerCase();
    return q ? all.filter((r) => JSON.stringify(r.result).toLowerCase().includes(q)) : all;
  });
}
