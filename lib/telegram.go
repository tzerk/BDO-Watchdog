package lib

import (
	"net/http"
	"time"
)

var t string

// ---------------------------------------------------------------------------------------------------------------------
// Send a telegram message using a query URL
func Send_TelegramMessage(config Config) {
	// Timestamp
	if config.TimeStamp {
		t = "[" + string(time.Now().Format("2006/01/02 15:04:05")) + "] %0A"
	}
	// Learn how to setup a telegram bot: https://core.telegram.org/bots
	resp, _ := http.Get("https://api.telegram.org/bot" + config.Botid +
		":" + config.Token +
		"/sendMessage?chat_id=" + config.Chatid +
		"&text=" + t + config.Message)
	defer resp.Body.Close()
}
