package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTP client settings. The timeout has to outlast the long poll.
const (
	pollSeconds = 30
	timeout     = (pollSeconds + 10) * time.Second
)

// ErrChatGone means the chat can no longer receive messages, e.g. the user blocked the bot.
var ErrChatGone = errors.New("chat is gone")

// Provider talks to the Telegram Bot API.
type Provider struct {
	http  *http.Client
	token string
}

// New creates a Provider for a bot token from @BotFather.
func New(token string) *Provider {
	return &Provider{http: &http.Client{Timeout: timeout}, token: token}
}

// GetUpdates long-polls for messages with an update ID of at least offset.
func (p *Provider) GetUpdates(ctx context.Context, offset int64) ([]Update, error) {
	var out []Update
	err := p.call(ctx, "getUpdates", getUpdatesRequest{
		Offset:         offset,
		Timeout:        pollSeconds,
		AllowedUpdates: []string{"message"},
	}, &out)
	return out, err
}

// SendWebAppMessage sends text with a button that opens the Mini App at webAppURL.
func (p *Provider) SendWebAppMessage(ctx context.Context, chatID int64, text, button, webAppURL string) error {
	return p.call(ctx, "sendMessage", sendMessageRequest{
		ChatID: chatID,
		Text:   text,
		ReplyMarkup: inlineKeyboard{InlineKeyboard: [][]inlineButton{{
			{Text: button, WebApp: webApp{URL: webAppURL}},
		}}},
	}, nil)
}

// SetMenuButton makes the chat's menu button open the Mini App at webAppURL.
func (p *Provider) SetMenuButton(ctx context.Context, text, webAppURL string) error {
	return p.call(ctx, "setChatMenuButton", setMenuButtonRequest{
		MenuButton: menuButton{Type: "web_app", Text: text, WebApp: webApp{URL: webAppURL}},
	}, nil)
}

func (p *Provider) call(ctx context.Context, method string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, MethodURL(p.token, method), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram %s: bad request", method)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		// A *url.Error quotes the URL, which holds the token.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fmt.Errorf("telegram %s: %w", method, err)
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("telegram %s: status %d: %w", method, resp.StatusCode, err)
	}
	if !r.OK {
		if r.ErrorCode == http.StatusForbidden || strings.Contains(r.Description, "chat not found") {
			return fmt.Errorf("telegram %s: %s: %w", method, r.Description, ErrChatGone)
		}
		return fmt.Errorf("telegram %s: %d %s", method, r.ErrorCode, r.Description)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(r.Result, out); err != nil {
		return fmt.Errorf("telegram %s: %w", method, err)
	}
	return nil
}
