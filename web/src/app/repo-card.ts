import { Component, computed, input } from '@angular/core';
import { DatePipe, DecimalPipe } from '@angular/common';
import { RepoResult } from './models';

@Component({
  selector: 'app-repo-card',
  imports: [DatePipe, DecimalPipe],
  templateUrl: './repo-card.html',
  styleUrl: './repo-card.css',
})
export class RepoCard {
  readonly result = input.required<RepoResult>();
  readonly rank = input.required<number>();

  protected readonly repo = computed(() => this.result().evidence.repo);
  protected readonly insight = computed(() => this.result().insight);
  protected readonly whyPopular = computed(() => {
    const why = this.insight()?.why_popular;
    return why === undefined ? null : splitReasons(why);
  });

  protected readonly periodLabel = computed(() => {
    switch (this.repo().period) {
      case 'weekly':
        return 'this week';
      case 'monthly':
        return 'this month';
      default:
        return 'today';
    }
  });

  protected host(url: string): string {
    try {
      return new URL(url).hostname.replace(/^www\./, '');
    } catch {
      return url;
    }
  }
}

// Older reports store why_popular as one string with inline numbering such as
// "Intro. (1) First reason. (2) Second reason." Split that into an intro and the reasons.
function splitReasons(why: string[] | string): { intro: string; reasons: string[] } {
  if (Array.isArray(why)) {
    return { intro: '', reasons: why };
  }
  // Keep only markers numbered 1, 2, 3... in order, so dates like "Oct 26. " are skipped.
  const sequential: RegExpExecArray[] = [];
  for (const m of why.matchAll(/(?<=^|\s)\(?(\d{1,2})[.)]\s+/g)) {
    if (Number(m[1]) === sequential.length + 1) {
      sequential.push(m);
    }
  }
  if (sequential.length < 2) {
    return { intro: why, reasons: [] };
  }
  const reasons = sequential.map((m, i) => {
    const end = i + 1 < sequential.length ? sequential[i + 1].index : why.length;
    return why.slice(m.index + m[0].length, end).trim();
  });
  return { intro: why.slice(0, sequential[0].index).trim(), reasons };
}
