import { Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ReportSummary } from '../../core/models/report';
import { DateTabs } from '../../shared/components/date-tabs/date-tabs';

@Component({
  selector: 'app-top-bar',
  imports: [RouterLink, DateTabs],
  templateUrl: './top-bar.html',
  styleUrl: './top-bar.css',
})
export class TopBar {
  readonly reports = input<ReportSummary[]>([]);
  readonly current = input<string>();
}
