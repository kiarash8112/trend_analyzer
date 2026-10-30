import { Component, input, output } from '@angular/core';
import { RepoResult } from '../../../core/models/report';
import { TimetableRow } from '../timetable-row/timetable-row';

export interface RankedResult {
  rank: number;
  result: RepoResult;
}

/** Today's repos as a departure timetable; one row can be open at a time. */
@Component({
  selector: 'app-timetable',
  imports: [TimetableRow],
  templateUrl: './timetable.html',
  styleUrl: './timetable.css',
})
export class Timetable {
  readonly rows = input.required<RankedResult[]>();
  readonly openRank = input<number | null>(null);
  readonly reportDay = input.required<number>();
  readonly toggle = output<number>();
}
