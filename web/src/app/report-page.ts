import { Component, computed, inject, input, signal } from '@angular/core';
import { httpResource } from '@angular/common/http';
import { RepoResult } from './models';
import { ReportApi } from './report-api';
import { RepoCard } from './repo-card';

@Component({
  selector: 'app-report-page',
  imports: [RepoCard],
  templateUrl: './report-page.html',
  styleUrl: './report-page.css',
})
export class ReportPage {
  /** Route param; when absent the newest report is shown. */
  readonly name = input<string>();

  private readonly api = inject(ReportApi);

  protected readonly current = computed(() => this.name() ?? this.api.list.value()?.[0]?.name);
  protected readonly summary = computed(() => this.api.list.value()?.find((r) => r.name === this.current()));
  protected readonly listLoading = this.api.list.isLoading;

  protected readonly report = httpResource<RepoResult[]>(() => {
    const name = this.current();
    return name ? `/api/reports/${encodeURIComponent(name)}` : undefined;
  });

  protected readonly query = signal('');

  protected readonly results = computed(() => {
    const q = this.query().trim().toLowerCase();
    const all = this.report.value() ?? [];
    if (!q) return all;
    return all.filter((r) => JSON.stringify(r).toLowerCase().includes(q));
  });
}
