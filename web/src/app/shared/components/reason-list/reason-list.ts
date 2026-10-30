import { Component, input } from '@angular/core';

/** Numbered reasons, each on its own full-width line. */
@Component({
  selector: 'app-reason-list',
  templateUrl: './reason-list.html',
  styleUrl: './reason-list.css',
})
export class ReasonList {
  readonly reasons = input.required<string[]>();
}
