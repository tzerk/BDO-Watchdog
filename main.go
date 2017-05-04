package main

import (
	"strconv"
	"github.com/andlabs/ui"
	"os"
	. "BDO-Watchdog/lib"
)

// Variables
const VERSION = "0.1.9"
var errmsg string

func main() {


	// SETTINGS
	exp, _ := os.Executable()
	config, err := Read_Settings(exp)
	if err != nil {
		errmsg = err.Error()
	}

	//// GUI
	//--------------------------------------------------------------------------------------------------------------
	ui := ui.Main(func() {
		window := ui.NewWindow("BDO Watchdog v" + VERSION, 300, 80, false)

		label_Process := ui.NewLabel("  Process: " + config.Process)
		label_Status := ui.NewLabel("  Initializing...")
		label_PID := ui.NewLabel("")
		label_Connection := ui.NewLabel("")
		label_Update := ui.NewLabel("")

		//box := ui.NewVerticalBox()
		box_main := ui.NewVerticalBox()
		box_settings := ui.NewVerticalBox()
		box_about := ui.NewVerticalBox()
		tab := ui.NewTab()

		pb := ui.NewProgressBar()

		tbtn := ui.NewButton("Send Telegram test message")
		tbtn.OnClicked(func(*ui.Button) {
			Send_TelegramMessage(config, label_Update, pb)
		})

		sep := ui.NewHorizontalSeparator()

		// Append all UI elements to the box container
		box_main.Append(label_Process, false)
		box_main.Append(label_Status, false)
		box_main.Append(label_PID, false)
		box_main.Append(label_Connection, false)
		box_main.Append(label_Update, false)
		box_main.Append(tbtn, false)
		box_main.Append(sep, false)
		box_main.Append(pb, true)

		// Append UI elements for settings tab
		box_settings.Append(ui.NewLabel("Token: " + config.Token), false)
		box_settings.Append(ui.NewLabel("Bot ID: " + config.Botid), false)
		box_settings.Append(ui.NewLabel("Chat ID: " + config.Chatid), false)
		box_settings.Append(ui.NewLabel("Timestamp: " + strconv.FormatBool(config.TimeStamp)), false)
		box_settings.Append(ui.NewLabel("Message: " + config.Message), false)
		box_settings.Append(ui.NewLabel("Stay alive: " + strconv.FormatBool(config.StayAlive)), false)
		box_settings.Append(ui.NewLabel("Process: " + config.Process), false)
		box_settings.Append(ui.NewLabel("Polling interval: " + strconv.Itoa(config.TimeBetweenChecksInS)), false)
		box_settings.Append(ui.NewLabel("Kill process after disconnect: " + strconv.FormatBool(config.KillOnDC)), false)
		box_settings.Append(ui.NewLabel("Shutdown PC after disconnect: " + strconv.FormatBool(config.ShutdownOnDC) ), false)
		box_settings.Append(ui.NewLabel("Kill CoherenUI_Host.exe: " + strconv.FormatBool(config.KillCoherentUI)), false)
		box_settings.Append(ui.NewLabel("Process Priority: " + config.ProcessPriority), false)

		// Append UI elements for about tab
		box_about.Append(ui.NewLabel("Version " + VERSION), false)
		box_about.Append(ui.NewLabel("Check https://github.com/tzerk/BDO-Watchdog/releases/ for updates"), false)


		tab.Append("Main", box_main)
		tab.Append("Settings", box_settings)
		tab.Append("About", box_about)


		// If this tool requires admin rights add a text label to inform the user
		if config.KillOnDC || config.KillCoherentUI {
			box_main.Append(ui.NewLabel("Make sure to have started this program with admin rights!"), true)
		}

		box := ui.NewHorizontalBox()
		box.Append(tab, false)

		window.SetChild(box)

		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()

		if (errmsg) != "" {
			label_Status.SetText("  " + errmsg)
		} else {
			go Watchdog(config, label_Status, label_PID, label_Connection, label_Update, pb)
		}
	})
	if ui != nil {
		panic(ui)
	}
}


