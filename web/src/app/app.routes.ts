import { Routes } from '@angular/router';
import { ReportPage } from './pages/report-page/report-page';
import { SlidePage } from './pages/slide-page/slide-page';

export const routes: Routes = [
  { path: '', component: ReportPage },
  { path: 'reports/:name', component: ReportPage },
  { path: 'reports/:name/repos/:rank', component: SlidePage },
  { path: '**', redirectTo: '' },
];
