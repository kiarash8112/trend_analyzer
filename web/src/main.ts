import { bootstrapApplication } from '@angular/platform-browser';
import { appConfig } from './app/app.config';
import { App } from './app/app';
import { startTelegramMiniApp } from './app/core/utils/telegram';

startTelegramMiniApp()
  .then(() => bootstrapApplication(App, appConfig))
  .catch((err) => console.error(err));
