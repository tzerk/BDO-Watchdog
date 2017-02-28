package lib

import "net/http"

// ---------------------------------------------------------------------------------------------------------------------
// Send a telegram message using a query URL
func Send_TelegramMessage(config Config) {
	// Learn how to setup a telegram bot: https://core.telegram.org/bots
	resp, _ := http.Get("https://api.telegram.org/bot" + config.Botid +
		":" + config.Token +
		"/sendMessage?chat_id=" + config.Chatid +
		"&text=" + config.Message)
	defer resp.Body.Close()
}
