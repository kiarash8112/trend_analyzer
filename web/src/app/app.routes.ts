import { Routes } from '@angular/router';
import { ReportPage } from './pages/report-page/report-page';

export const routes: Routes = [
  { path: '', component: ReportPage },
  { path: 'reports/:name', component: ReportPage },
  { path: '**', redirectTo: '' },
];
