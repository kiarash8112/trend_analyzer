const SDK_URL = 'https://telegram.org/js/telegram-web-app.js';

type TelegramWindow = Window & { Telegram?: { WebApp?: { ready(): void; expand(): void } } };

/**
 * Inside Telegram, loads its SDK and opens the Mini App at full height. Elsewhere it does nothing,
 * so the public site never depends on telegram.org. The SDK reads the launch data from the URL
 * hash, so this has to finish before the router starts. The site still runs if the SDK fails.
 */
export function startTelegramMiniApp(): Promise<void> {
  if (!location.hash.includes('tgWebApp')) return Promise.resolve();

  return new Promise((resolve) => {
    const script = document.createElement('script');
    script.src = SDK_URL;
    script.onload = () => {
      const app = (window as TelegramWindow).Telegram?.WebApp;
      app?.ready();
      app?.expand();
      resolve();
    };
    script.onerror = () => resolve();
    document.head.appendChild(script);
  });
}
