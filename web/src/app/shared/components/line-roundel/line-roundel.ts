import { Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';

/** The numbered line badge; it links to the repo's slide. */
@Component({
  selector: 'app-line-roundel',
  imports: [RouterLink],
  templateUrl: './line-roundel.html',
  styleUrl: './line-roundel.css',
})
export class LineRoundel {
  readonly number = input.required<number>();
  /** Router link; null renders an inert badge (for a failed analysis). */
  readonly link = input<unknown[] | null>(null);
  readonly label = input.required<string>();
}
