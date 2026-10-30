import { Injectable } from '@angular/core';
import { httpResource } from '@angular/common/http';
import { ReportSummary, RepoResult } from '../models/report';

@Injectable({ providedIn: 'root' })
export class ReportApi {
  /** All saved reports, newest first. */
  readonly list = httpResource<ReportSummary[]>(() => '/api/reports');

  /** The results of one report; call from an injection context. Skips the request while name is undefined. */
  report(name: () => string | undefined) {
    return httpResource<RepoResult[]>(() => {
      const n = name();
      return n ? `/api/reports/${encodeURIComponent(n)}` : undefined;
    });
  }
}
