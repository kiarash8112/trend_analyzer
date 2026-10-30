import { Component, input } from '@angular/core';

/** The research sources behind an analysis, behind a disclosure. */
@Component({
  selector: 'app-source-list',
  templateUrl: './source-list.html',
  styleUrl: './source-list.css',
})
export class SourceList {
  readonly sources = input.required<string[]>();
}
