package telegram

import "encoding/json"

// ChatTypePrivate is a one-to-one chat with a user. Web App buttons only work there.
const ChatTypePrivate = "private"

// Update is one incoming event from getUpdates. Message is nil for other event types.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

// Message is an incoming message.
type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

// Chat is the conversation a message belongs to.
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// response is the envelope around every Bot API result.
type response struct {
	OK          bool            `json:"ok"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type getUpdatesRequest struct {
	Offset         int64    `json:"offset"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates"`
}

type sendMessageRequest struct {
	ChatID      int64          `json:"chat_id"`
	Text        string         `json:"text"`
	ReplyMarkup inlineKeyboard `json:"reply_markup"`
}

type inlineKeyboard struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

type inlineButton struct {
	Text   string `json:"text"`
	WebApp webApp `json:"web_app"`
}

type webApp struct {
	URL string `json:"url"`
}

type setMenuButtonRequest struct {
	MenuButton menuButton `json:"menu_button"`
}

type menuButton struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	WebApp webApp `json:"web_app"`
}
