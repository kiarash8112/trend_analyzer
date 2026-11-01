package mocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"trend_analyzer/src/providers/telegram"
	"trend_analyzer/src/services"
)

const testToken = "123:secret"

// roundTrip fakes the Telegram API behind http.DefaultTransport.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func stubTransport(t *testing.T, f roundTrip) {
	t.Helper()
	old := http.DefaultTransport
	http.DefaultTransport = f
	t.Cleanup(func() { http.DefaultTransport = old })
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}
}

// sent records sendMessage calls; refuse maps a chat ID to the API error it answers with.
type sent struct {
	texts  map[int64]string
	refuse map[int64]string
}

func stubSend(t *testing.T) *sent {
	t.Helper()
	s := &sent{texts: map[int64]string{}, refuse: map[int64]string{}}
	stubTransport(t, func(r *http.Request) (*http.Response, error) {
		var msg struct {
			ChatID int64  `json:"chat_id"`
			Text   string `json:"text"`
		}
		json.NewDecoder(r.Body).Decode(&msg)
		if body, ok := s.refuse[msg.ChatID]; ok {
			return jsonResponse(http.StatusBadRequest, body), nil
		}
		s.texts[msg.ChatID] = msg.Text
		return jsonResponse(http.StatusOK, `{"ok":true,"result":{}}`), nil
	})
	return s
}

func newBot(t *testing.T) (*services.TelegramBot, *services.Subscribers, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data", "subs.json")
	subs, err := services.OpenSubscribers(path)
	if err != nil {
		t.Fatal(err)
	}
	return services.NewTelegramBot(telegram.New(testToken), subs, dir, "https://example.com"), subs, path
}

func at(day, hour int) time.Time { return time.Date(2026, 11, day, hour, 0, 0, 0, time.UTC) }

func start(chat int64, chatType, text string) telegram.Update {
	return telegram.Update{Message: &telegram.Message{Chat: telegram.Chat{ID: chat, Type: chatType}, Text: text}}
}

func TestStartThenRemindOncePerDay(t *testing.T) {
	sends := stubSend(t)
	bot, _, _ := newBot(t)
	ctx := context.Background()

	bot.HandleUpdate(ctx, start(1, telegram.ChatTypePrivate, "/start"), at(1, 12))
	delete(sends.texts, 1)

	bot.Remind(ctx, at(1, 13))
	if _, ok := sends.texts[1]; ok {
		t.Error("subscribing after the reminder hour should wait for the next day")
	}
	bot.Remind(ctx, at(2, 8))
	if _, ok := sends.texts[1]; ok {
		t.Error("reminded before the reminder hour")
	}
	bot.Remind(ctx, at(2, 9))
	if _, ok := sends.texts[1]; !ok {
		t.Fatal("expected a reminder at the reminder hour")
	}
	delete(sends.texts, 1)
	bot.Remind(ctx, at(2, 20))
	if _, ok := sends.texts[1]; ok {
		t.Error("reminded twice in one day")
	}
}

func TestRemindCatchesUpAfterDowntimeAndRestart(t *testing.T) {
	sends := stubSend(t)
	bot, subs, path := newBot(t)
	ctx := context.Background()
	if err := subs.Add(1, "2026-10-20"); err != nil {
		t.Fatal(err)
	}

	bot.Remind(ctx, at(2, 15))
	if _, ok := sends.texts[1]; !ok {
		t.Fatal("expected one reminder after downtime")
	}

	reopened, err := services.OpenSubscribers(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reopened.Has(1) || len(reopened.Due("2026-11-02")) != 0 {
		t.Error("the sent slot should survive a restart")
	}
}

func TestStopAndBlockedChatsLeaveTheList(t *testing.T) {
	sends := stubSend(t)
	bot, subs, _ := newBot(t)
	ctx := context.Background()
	subs.Add(1, "")
	subs.Add(2, "")
	subs.Add(3, "")

	bot.HandleUpdate(ctx, start(1, telegram.ChatTypePrivate, "/stop"), at(1, 12))
	if subs.Has(1) {
		t.Error("/stop should unsubscribe")
	}

	sends.refuse[2] = `{"ok":false,"error_code":403,"description":"Forbidden: bot was blocked by the user"}`
	sends.refuse[3] = `{"ok":false,"error_code":500,"description":"Internal Server Error"}`
	bot.Remind(ctx, at(1, 12))
	if subs.Has(2) {
		t.Error("a blocked chat should be dropped")
	}
	if !subs.Has(3) || len(subs.Due("2026-11-01")) != 1 {
		t.Error("a temporary failure should keep the chat due for a retry")
	}
}

func TestLogsEventsButNotMessageText(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	sends := stubSend(t)
	bot, subs, _ := newBot(t)
	ctx := context.Background()
	subs.Add(1, "")
	subs.Add(2, "")
	sends.refuse[2] = `{"ok":false,"error_code":403,"description":"Forbidden: bot was blocked by the user"}`

	bot.HandleUpdate(ctx, start(1, telegram.ChatTypePrivate, "my private note"), at(1, 12))
	bot.Remind(ctx, at(1, 12))

	out := logs.String()
	for _, want := range []string{"chat 1: other message", "sending to 2 chats", "chat 2: removed", "sent 1 of 2"} {
		if !strings.Contains(out, want) {
			t.Errorf("logs miss %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "private note") || strings.Contains(out, "secret") {
		t.Errorf("logs leak message text or the token:\n%s", out)
	}
}

func TestIgnoresGroupChats(t *testing.T) {
	sends := stubSend(t)
	bot, subs, _ := newBot(t)

	bot.HandleUpdate(context.Background(), start(-5, "group", "/start"), at(1, 12))
	if subs.Has(-5) || len(sends.texts) != 0 {
		t.Error("group chats should be ignored: Web App buttons only work in private chats")
	}
}

func TestReminderNamesTheLatestReport(t *testing.T) {
	sends := stubSend(t)
	dir := t.TempDir()
	subs, err := services.OpenSubscribers(filepath.Join(dir, "subs.json"))
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"2026-10-29-daily.json": `[{}]`,
		"2026-10-30-daily.json": `[{},{},{}]`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	subs.Add(1, "")

	services.NewTelegramBot(telegram.New(testToken), subs, dir, "https://example.com").Remind(context.Background(), at(1, 12))
	if !strings.Contains(sends.texts[1], "2026-10-30, 3 repos") {
		t.Errorf("reminder = %q", sends.texts[1])
	}
}

func TestProviderSendsWebAppButton(t *testing.T) {
	var path, body string
	stubTransport(t, func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		path, body = r.URL.Path, string(b)
		return jsonResponse(http.StatusOK, `{"ok":true,"result":{}}`), nil
	})

	err := telegram.New(testToken).SendWebAppMessage(context.Background(), 7, "hi", "Open", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if path != "/bot"+testToken+"/sendMessage" {
		t.Errorf("path = %q", path)
	}
	want := `"reply_markup":{"inline_keyboard":[[{"text":"Open","web_app":{"url":"https://example.com"}}]]}`
	if !strings.Contains(body, want) {
		t.Errorf("body = %s", body)
	}
}

func TestProviderErrorsHideTheToken(t *testing.T) {
	stubTransport(t, func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})

	_, err := telegram.New(testToken).GetUpdates(context.Background(), 0)
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("error leaks the token: %v", err)
	}
}
