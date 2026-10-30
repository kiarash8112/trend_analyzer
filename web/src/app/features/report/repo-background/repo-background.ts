import { Component, computed, input } from '@angular/core';
import { DatePipe, DecimalPipe } from '@angular/common';
import { RepoResult } from '../../../core/models/report';
import { HostPipe } from '../../../core/pipes/host.pipe';
import { EventList } from '../../../shared/components/event-list/event-list';
import { SourceList } from '../../../shared/components/source-list/source-list';

const CONFIDENCE: Record<string, string> = {
  high: 'High: most claims have a direct source.',
  medium: 'Medium: some steps are inferred from indirect sources.',
  low: 'Low: mostly inferred; treat it as a hypothesis.',
};

/** The rest of a repo's story: what it is, how it spread, facts, events and sources. */
@Component({
  selector: 'app-repo-background',
  imports: [DatePipe, DecimalPipe, HostPipe, EventList, SourceList],
  templateUrl: './repo-background.html',
  styleUrl: './repo-background.css',
})
export class RepoBackground {
  readonly result = input.required<RepoResult>();

  protected readonly repo = computed(() => this.result().evidence.repo);
  protected readonly insight = computed(() => this.result().insight!);
  protected readonly confidence = computed(
    () => CONFIDENCE[this.insight().confidence] ?? this.insight().confidence,
  );
}
