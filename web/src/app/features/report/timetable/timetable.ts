import { Component, input } from '@angular/core';
import { RepoResult } from '../../../core/models/report';
import { TimetableRow } from '../timetable-row/timetable-row';

export interface RankedResult {
  rank: number;
  result: RepoResult;
}

/** Today's repos as a departure timetable; each row opens its repo's slide. */
@Component({
  selector: 'app-timetable',
  imports: [TimetableRow],
  templateUrl: './timetable.html',
  styleUrl: './timetable.css',
})
export class Timetable {
  readonly rows = input.required<RankedResult[]>();
  /** The report the rows belong to, for the slide links. */
  readonly reportName = input.required<string>();
}
