import { Component, computed, input } from '@angular/core';
import { DatePipe, DecimalPipe } from '@angular/common';
import { RepoResult } from '../../../core/models/report';
import { HostPipe } from '../../../core/pipes/host.pipe';
import { splitReasons } from '../../../core/utils/reasons';
import { EventList } from '../../../shared/components/event-list/event-list';
import { ReasonList } from '../../../shared/components/reason-list/reason-list';
import { RouteMap } from '../../../shared/components/route-map/route-map';
import { SourceList } from '../../../shared/components/source-list/source-list';

const CONFIDENCE: Record<string, string> = {
  high: 'High: most claims have a direct source.',
  medium: 'Medium: some steps are inferred from indirect sources.',
  low: 'Low: mostly inferred; treat it as a hypothesis.',
};

/** The opened row: route, every reason, background, facts and sources. */
@Component({
  selector: 'app-repo-detail',
  imports: [DatePipe, DecimalPipe, HostPipe, RouteMap, ReasonList, EventList, SourceList],
  templateUrl: './repo-detail.html',
  styleUrl: './repo-detail.css',
})
export class RepoDetail {
  readonly result = input.required<RepoResult>();
  readonly reportDay = input.required<number>();

  protected readonly repo = computed(() => this.result().evidence.repo);
  protected readonly insight = computed(() => this.result().insight!);
  protected readonly why = computed(() => splitReasons(this.insight().why_popular));
  protected readonly confidence = computed(
    () => CONFIDENCE[this.insight().confidence] ?? this.insight().confidence,
  );
}
