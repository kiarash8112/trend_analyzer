import { Routes } from '@angular/router';
import { ReportPage } from './report-page';

export const routes: Routes = [
  { path: '', component: ReportPage },
  { path: 'reports/:name', component: ReportPage },
  { path: '**', redirectTo: '' },
];
