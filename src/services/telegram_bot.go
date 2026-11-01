package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"trend_analyzer/src/providers/telegram"
)

// Bot behaviour. Telegram does not share a user's time zone, so everyone gets the same hour.
const (
	reminderHourUTC = 9
	remindEvery     = time.Minute
	pollRetryDelay  = 5 * time.Second
	sendGap         = 50 * time.Millisecond // Telegram allows about 30 messages a second.
	openButton      = "Open Trend Analyzer"
	menuButtonText  = "Open"
)

// TelegramBot opens the site as a Mini App and reminds subscribers to read it every day.
type TelegramBot struct {
	api        *telegram.Provider
	subs       *Subscribers
	reportsDir string
	webAppURL  string
}

// NewTelegramBot creates a TelegramBot.
func NewTelegramBot(api *telegram.Provider, subs *Subscribers, reportsDir, webAppURL string) *TelegramBot {
	return &TelegramBot{api: api, subs: subs, reportsDir: reportsDir, webAppURL: webAppURL}
}

// Run answers messages and sends reminders until ctx is cancelled.
func (b *TelegramBot) Run(ctx context.Context) error {
	if err := b.api.SetMenuButton(ctx, menuButtonText, b.webAppURL); err != nil {
		return err
	}
	log.Printf("menu button set, %d subscribers, reminders daily at %02d:00 UTC", b.subs.Count(), reminderHourUTC)

	var wg sync.WaitGroup
	wg.Go(func() { b.remindLoop(ctx) })
	wg.Go(func() { b.pollLoop(ctx) })
	wg.Wait()
	log.Print("bot stopped")
	return nil
}

func (b *TelegramBot) pollLoop(ctx context.Context) {
	var offset int64
	failing := false
	for ctx.Err() == nil {
		updates, err := b.api.GetUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("poll: %v", err)
				failing = true
			}
			sleep(ctx, pollRetryDelay)
			continue
		}
		if failing {
			log.Print("poll: reconnected")
			failing = false
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			b.HandleUpdate(ctx, u, time.Now())
		}
	}
}

// remindLoop checks often instead of sleeping until the hour, so a chat that failed or a bot
// that was down catches up on its own.
func (b *TelegramBot) remindLoop(ctx context.Context) {
	for ctx.Err() == nil {
		b.Remind(ctx, time.Now())
		sleep(ctx, remindEvery)
	}
}

// HandleUpdate answers one incoming message.
func (b *TelegramBot) HandleUpdate(ctx context.Context, u telegram.Update, now time.Time) {
	if u.Message == nil {
		return
	}
	chat := u.Message.Chat.ID
	if u.Message.Chat.Type != telegram.ChatTypePrivate {
		log.Printf("chat %d: ignored, %s chat", chat, u.Message.Chat.Type)
		return
	}

	var text string
	switch command(u.Message.Text) {
	case "/start":
		log.Printf("chat %d: /start", chat)
		if err := b.subs.Add(chat, reminderSlot(now)); err != nil {
			log.Printf("chat %d: subscribe: %v", chat, err)
		}
		text = fmt.Sprintf("See what's trending on GitHub and why. I'll remind you every day at %02d:00 UTC. Send /stop to turn reminders off.", reminderHourUTC)
	case "/stop":
		log.Printf("chat %d: /stop", chat)
		if err := b.subs.Remove(chat); err != nil {
			log.Printf("chat %d: unsubscribe: %v", chat, err)
		}
		text = "Reminders are off. Send /start to turn them back on."
	default:
		// Only the fact: what people write to the bot stays out of the logs.
		log.Printf("chat %d: other message", chat)
		text = "Send /start to get a daily reminder, or /stop to turn it off."
	}
	if err := b.send(ctx, chat, text); err != nil {
		log.Printf("chat %d: reply: %v", chat, err)
	}
}

// Remind sends the current slot's reminder to every chat that has not received it.
func (b *TelegramBot) Remind(ctx context.Context, now time.Time) {
	slot := reminderSlot(now)
	due := b.subs.Due(slot)
	if len(due) == 0 {
		return
	}

	log.Printf("reminder %s: sending to %d chats", slot, len(due))
	text := b.reminderText()
	sent := 0
	for _, chat := range due {
		if ctx.Err() != nil {
			break
		}
		if err := b.send(ctx, chat, text); err != nil {
			log.Printf("chat %d: remind: %v", chat, err)
		} else {
			sent++
			if err := b.subs.MarkReminded(chat, slot); err != nil {
				log.Printf("chat %d: remind: %v", chat, err)
			}
		}
		sleep(ctx, sendGap)
	}
	log.Printf("reminder %s: sent %d of %d", slot, sent, len(due))
}

// send delivers text with the Mini App button and drops chats that can no longer receive it.
func (b *TelegramBot) send(ctx context.Context, chat int64, text string) error {
	err := b.api.SendWebAppMessage(ctx, chat, text, openButton, b.webAppURL)
	if errors.Is(err, telegram.ErrChatGone) {
		log.Printf("chat %d: removed, it can no longer receive messages", chat)
		if rmErr := b.subs.Remove(chat); rmErr != nil {
			log.Printf("chat %d: unsubscribe: %v", chat, rmErr)
		}
	}
	return err
}

func (b *TelegramBot) reminderText() string {
	text := "Time to check what's trending on GitHub."
	reports, err := ListReports(b.reportsDir)
	if err != nil {
		log.Printf("list reports: %v", err)
		return text
	}
	if len(reports) == 0 {
		return text
	}
	return fmt.Sprintf("%s\nLatest report: %s, %d repos.", text, reports[0].Date, reports[0].RepoCount)
}

// reminderSlot names the most recent reminder time at or before now: the UTC date, which only
// turns over at the reminder hour. Chats that subscribe get the current slot, so their first
// reminder comes at the next reminder hour.
func reminderSlot(now time.Time) string {
	return now.UTC().Add(-reminderHourUTC * time.Hour).Format(time.DateOnly)
}

// command returns the first word of a message.
func command(text string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(text), " ")
	return first
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
