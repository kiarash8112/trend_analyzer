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
