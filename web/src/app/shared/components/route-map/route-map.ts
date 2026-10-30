import { Component, computed, input } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { Insight } from '../../../core/models/report';
import { routeLayout } from '../../../core/utils/route-layout';

/**
 * A repo's timeline drawn as a metro route spaced by real dates. Phones get the same
 * events as a vertical line.
 */
@Component({
  selector: 'app-route-map',
  imports: [NgTemplateOutlet],
  templateUrl: './route-map.html',
  styleUrl: './route-map.css',
})
export class RouteMap {
  readonly insight = input.required<Insight>();
  /** The report's date as a UTC timestamp; the route ends there. */
  readonly reportDay = input.required<number>();

  protected readonly layout = computed(() => routeLayout(this.insight(), this.reportDay()));
}
