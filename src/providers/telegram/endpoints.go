package telegram

// BaseURL is the Telegram Bot API.
const BaseURL = "https://api.telegram.org"

// MethodURL builds the URL of a Bot API method. It contains the bot token, so never log it.
func MethodURL(token, method string) string {
	return BaseURL + "/bot" + token + "/" + method
}
