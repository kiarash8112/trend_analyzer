import { Component, input } from '@angular/core';
import { TimelineEvent } from '../../../core/models/report';
import { HostPipe } from '../../../core/pipes/host.pipe';

/** Every dated event with its source, behind a disclosure. */
@Component({
  selector: 'app-event-list',
  imports: [HostPipe],
  templateUrl: './event-list.html',
  styleUrl: './event-list.css',
})
export class EventList {
  readonly events = input.required<TimelineEvent[]>();
}
