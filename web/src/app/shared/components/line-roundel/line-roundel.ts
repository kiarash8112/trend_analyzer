import { Component, input } from '@angular/core';

/** The numbered line badge; it is the keyboard button that opens and closes its row. */
@Component({
  selector: 'app-line-roundel',
  templateUrl: './line-roundel.html',
  styleUrl: './line-roundel.css',
})
export class LineRoundel {
  readonly number = input.required<number>();
  readonly expanded = input(false);
  readonly label = input.required<string>();
  readonly disabled = input(false);
}
