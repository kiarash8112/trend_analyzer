import { Component, computed, input } from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { RepoResult } from '../../../core/models/report';
import { splitReasons } from '../../../core/utils/reasons';
import { shortChannel } from '../../../core/utils/text';
import { ReasonList } from '../../../shared/components/reason-list/reason-list';
import { RouteMap } from '../../../shared/components/route-map/route-map';
import { RepoBackground } from '../repo-background/repo-background';

/** How many reasons sit on the slide itself; the rest continue below it. */
const ON_SLIDE = 3;

/** One repo as a slide: title, route and the top reasons on one screen, the full story below. */
@Component({
  selector: 'app-repo-slide',
  imports: [DecimalPipe, RouteMap, ReasonList, RepoBackground],
  templateUrl: './repo-slide.html',
  styleUrl: './repo-slide.css',
})
export class RepoSlide {
  readonly result = input.required<RepoResult>();
  readonly reportDay = input.required<number>();

  protected readonly onSlide = ON_SLIDE;
  protected readonly repo = computed(() => this.result().evidence.repo);
  protected readonly insight = computed(() => this.result().insight);
  protected readonly why = computed(() => {
    const i = this.insight();
    return i ? splitReasons(i.why_popular) : { intro: '', reasons: [] };
  });
  protected readonly channel = computed(() => shortChannel(this.insight()?.launch_channel ?? ''));
}
