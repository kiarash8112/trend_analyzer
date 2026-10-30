import { Component, computed, inject, input, linkedSignal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { ReportApi } from '../../core/services/report-api';
import { parseDay } from '../../core/utils/dates';
import { RepoSlide } from '../../features/report/repo-slide/repo-slide';
import { TopBar } from '../../layout/top-bar/top-bar';
import { ArrowLink } from '../../shared/components/arrow-link/arrow-link';
import { ProgressSegments } from '../../shared/components/progress-segments/progress-segments';

/** One repo of a report at a time, with previous/next, ← → keys and Esc back to the list. */
@Component({
  selector: 'app-slide-page',
  imports: [RouterLink, TopBar, ArrowLink, ProgressSegments, RepoSlide],
  templateUrl: './slide-page.html',
  styleUrl: './slide-page.css',
  host: { '(document:keydown)': 'onKey($event)' },
})
export class SlidePage {
  /** Route params. */
  readonly name = input.required<string>();
  readonly rank = input.required<string>();

  private readonly api = inject(ReportApi);
  private readonly router = inject(Router);

  protected readonly reports = computed(() => this.api.list.value() ?? []);
  protected readonly report = this.api.report(this.name);
  protected readonly results = computed(() => this.report.value() ?? []);
  protected readonly total = computed(() => this.results().length);
  protected readonly position = computed(() => Number(this.rank()) || 1);
  protected readonly current = computed(() => this.results()[this.position() - 1]);
  protected readonly names = computed(() =>
    this.results().map((r) => `${r.evidence.repo.owner}/${r.evidence.repo.name}`),
  );
  protected readonly reportDay = computed(() => {
    const summary = this.reports().find((r) => r.name === this.name());
    return parseDay(summary?.date ?? this.name()) ?? Date.now();
  });

  /** Slides enter from the side you are moving towards. */
  protected readonly direction = linkedSignal<number, 'next' | 'previous'>({
    source: this.position,
    computation: (now, prev) => (prev && now < prev.source ? 'previous' : 'next'),
  });

  protected readonly listLink = computed(() => ['/reports', this.name()]);
  protected readonly linkFor = (n: number) => ['/reports', this.name(), 'repos', n];
  protected readonly previousLink = computed(() =>
    this.position() > 1 ? this.linkFor(this.position() - 1) : null,
  );
  protected readonly nextLink = computed(() =>
    this.position() < this.total() ? this.linkFor(this.position() + 1) : null,
  );

  protected onKey(event: KeyboardEvent): void {
    const target = event.target as HTMLElement;
    if (event.altKey || event.ctrlKey || event.metaKey || target.closest('input, textarea')) {
      return;
    }
    const link =
      event.key === 'ArrowLeft'
        ? this.previousLink()
        : event.key === 'ArrowRight'
          ? this.nextLink()
          : event.key === 'Escape'
            ? this.listLink()
            : null;
    if (link) {
      event.preventDefault();
      this.router.navigate(link);
    }
  }
}
