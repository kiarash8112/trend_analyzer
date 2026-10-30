import { Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';

/** A round keycap arrow that links to the previous or next slide; inert at either end. */
@Component({
  selector: 'app-arrow-link',
  imports: [RouterLink],
  templateUrl: './arrow-link.html',
  styleUrl: './arrow-link.css',
})
export class ArrowLink {
  readonly direction = input.required<'previous' | 'next'>();
  /** Router link; null at the end of the list. */
  readonly link = input<unknown[] | null>(null);
  readonly label = input.required<string>();
}
