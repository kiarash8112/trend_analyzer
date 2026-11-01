// Command telegram_bot opens the site as a Telegram Mini App and reminds subscribers
// to read it every day. It needs TELEGRAM_BOT_TOKEN and an HTTPS TELEGRAM_WEB_APP_URL.
//
//	go run ./cmd/telegram_bot
package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"trend_analyzer/src/config"
	"trend_analyzer/src/providers/telegram"
	"trend_analyzer/src/services"
)

func main() {
	cfg := config.Load()
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set; create a bot with @BotFather")
	}
	// Telegram refuses to open a Mini App from anything but HTTPS.
	if u, err := url.Parse(cfg.TelegramWebAppURL); err != nil || u.Scheme != "https" || u.Host == "" {
		log.Fatal("TELEGRAM_WEB_APP_URL must be the site's public https:// address")
	}

	subs, err := services.OpenSubscribers(cfg.SubscribersFile)
	if err != nil {
		log.Fatalf("subscribers: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot := services.NewTelegramBot(telegram.New(cfg.TelegramBotToken), subs, cfg.ReportsDir, cfg.TelegramWebAppURL)
	log.Printf("bot running, opening %s", cfg.TelegramWebAppURL)
	if err := bot.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
