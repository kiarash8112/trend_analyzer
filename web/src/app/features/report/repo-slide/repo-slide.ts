import { Component, computed, input } from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { RepoResult } from '../../../core/models/report';
import { splitReasons } from '../../../core/utils/reasons';
import { shortChannel } from '../../../core/utils/text';
import { ReasonList } from '../../../shared/components/reason-list/reason-list';
import { RouteMap } from '../../../shared/components/route-map/route-map';
import { ProblemAnswer } from '../problem-answer/problem-answer';
import { RepoBackground } from '../repo-background/repo-background';

/** One repo as a slide: what it is, its route and every reason, then how it spread. */
@Component({
  selector: 'app-repo-slide',
  imports: [DecimalPipe, ProblemAnswer, RouteMap, ReasonList, RepoBackground],
  templateUrl: './repo-slide.html',
  styleUrl: './repo-slide.css',
})
export class RepoSlide {
  readonly result = input.required<RepoResult>();
  readonly reportDay = input.required<number>();

  protected readonly repo = computed(() => this.result().evidence.repo);
  protected readonly insight = computed(() => this.result().insight);
  protected readonly why = computed(() => {
    const i = this.insight();
    return i ? splitReasons(i.why_popular) : { intro: '', reasons: [] };
  });
  protected readonly channel = computed(() => shortChannel(this.insight()?.launch_channel ?? ''));
}
