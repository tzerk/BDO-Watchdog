package lib

import (
	"net/http"
	"time"
	"github.com/andlabs/ui"
	"strconv"
)

var t string
var sent bool
var pbVal int

// ---------------------------------------------------------------------------------------------------------------------
// Send a telegram message using a query URL
func Send_TelegramMessage(config Config, label_Update *ui.Label, pb *ui.ProgressBar) {

	// Timestamp
	if config.TimeStamp {
		t = "[" + string(time.Now().Format("2006/01/02 15:04:05")) + "] %0A"
	}

	for {
		// Update status message
		label_Update.SetText("  Sending Telegram message...")
		time.Sleep(2 * time.Second)

		// Learn how to setup a telegram bot: https://core.telegram.org/bots
		resp, err := http.Get("https://api.telegram.org/bot" + config.Botid +
			":" + config.Token +
			"/sendMessage?chat_id=" + config.Chatid +
			"&text=" + t + config.Message)

		// http.Get will return an error if there is no internet connection
		if err != nil {
			for i := 0; i <= 15; i++ {
				pbVal = int(100/float32(15) * float32(i))
				if pbVal > 100 {
					pbVal = 100
				}
				pb.SetValue(pbVal)
				label_Update.SetText("  No internet connection. Retry in... " + strconv.Itoa(15 - i) + " min")
				time.Sleep(1 * time.Minute)
			}
		// if we do have a connection, break the loop and exit this function
		} else {
			defer resp.Body.Close()
			break
		}
	}

}
