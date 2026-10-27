import { Injectable } from '@angular/core';
import { httpResource } from '@angular/common/http';
import { ReportSummary } from './models';

@Injectable({ providedIn: 'root' })
export class ReportApi {
  /** All saved reports, newest first. */
  readonly list = httpResource<ReportSummary[]>(() => '/api/reports');
}
