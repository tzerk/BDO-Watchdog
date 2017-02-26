package main

import (
	ps "github.com/mitchellh/go-ps"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"github.com/andlabs/ui"
	"time"
	"net/http"
	"os"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"syscall"
)

// Telegram and program settings (config.yml)
// Guide: http://sweetohm.net/article/go-yaml-parsers.en.html
type Config struct {
	Token string
	Botid string
	Chatid string
	Message string
	StayAlive bool
	Process string
	TimeBetweenChecksInS int
	KillOnDC bool
	ShutdownOnDC bool
}

// Variables
var STATUS bool = false
var CONNECTION bool = false
var PID int

func main() {

	////  SETTINGS
	//--------------------------------------------------------------------------------------------------------------
	// YAML PARSING
	var config Config
	source, err := ioutil.ReadFile("config.yml")
	if err != nil {
		// in theory, using yml.Marshal() would be more elegant, but we want to preserve the yaml comments
		// as well as set some default values/hints
		defconf :=
				"## Telegram Bot Settings\r\n" +
				"token: \r\n" +
				"botid: \r\n" +
				"chatid: \r\n" +
				"message: BDO disconnected \r\n" +
				"\r\n" +
				"## Program Settings\r\n" +
				"stayalive: false\r\n" +
				"process: BlackDesert64.exe\r\n" +
				"timebetweenchecksins: 60" +
				"killondc: true" +
				"shutdownondc: false"
		ioutil.WriteFile("config.yml", []byte(defconf), os.FileMode(int(0666)))
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}


	//// GUI
	//--------------------------------------------------------------------------------------------------------------
	ui := ui.Main(func() {
		window := ui.NewWindow("BDO Watchdog", 300, 80, false)

		label_Process := ui.NewLabel("  Process: " + config.Process)
		label_Status := ui.NewLabel("  Initializing...")
		label_PID := ui.NewLabel("-")
		label_Connection := ui.NewLabel("-")
		label_Update := ui.NewLabel("")

		box := ui.NewVerticalBox()
		sep := ui.NewHorizontalSeparator()
		pb := ui.NewProgressBar()

		// Append all UI elements to the box container
		box.Append(label_Process, false)
		box.Append(label_Status, false)
		box.Append(label_PID, false)
		box.Append(label_Connection, false)
		box.Append(label_Update, false)
		box.Append(sep, false)
		box.Append(pb, true)

		window.SetChild(box)
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
		go observer(config, label_Status, label_PID, label_Connection, label_Update, pb)
	})
	if ui != nil {
		panic(ui)
	}
}

//--------------------------------------------------------------------------------------------------------------
// PROCESS
//--------------------------------------------------------------------------------------------------------------
func observer(
	config Config,
	label_Status *ui.Label,
	label_PID *ui.Label,
	label_Connection *ui.Label,
	label_Update *ui.Label,
	pb *ui.ProgressBar) {

	for {
		label_Update.SetText("")

		//// EXIT CONDITION
		//-----------------
		// If the process is running, but no longer connected we trigger the following actions
		if STATUS && !CONNECTION {

			// Use the Telegram API to send a message
			send_TelegramMessage(config)

			// Optional: kill the monitored process if it is disconnected
			/*
			if config.KillOnDC {
				fmt.Println(PID)
				proc, err := os.FindProcess(PID)
				fmt.Println(proc)
				if err != nil {
					log.Println(err)
				}
				// Kill the process
				kill_err := proc.Kill()
				if kill_err != nil {
					log.Println(err)
				}

				time.Sleep(5 * time.Second)
			}
			*/

			// Optional: shutdown the computer if the monitored process is disconnected
			if config.ShutdownOnDC {
				exec.Command("cmd", "/C", "shutdown", "/s").Run()
			}

			// Optional (YAML file, default: false): keep this program open even if
			// the process is disconnected
			if !config.StayAlive {
				os.Exit(1)
			}
		}

		//// PROCESS
		//----------
		p, err := ps.Processes()
		if err != nil {
			log.Fatal(err)
		}

		//// PID
		//------
		for _, v := range p {
			if v.Executable() == config.Process {
				PID = v.Pid()
			}
		}
		if (PID == 0) {
			ui.QueueMain(func () {
				STATUS = false
				label_Status.SetText("  Status: not running")
				label_PID.SetText("  PID: -")
				label_Connection.SetText("  Connection: -" )
			})

			wait(config, label_Update, pb)
			continue
		} else {

			ui.QueueMain(func () {
				STATUS = true
				label_Status.SetText("  Status: running")
				label_PID.SetText("  PID: " + strconv.Itoa(PID))
			})
		}

		//// CONNECTION STATUS
		//--------------------
		// NETSTAT
		// the syscall.SysProcAttr trick found here:
		// https://www.reddit.com/r/golang/comments/2c1g3x/build_golang_app_reverse_shell_to_run_in_windows/
		cmd := exec.Command("cmd.exe", "/C netstat -aon")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}


		// RegEx matching; try to find the PID in the netstat output
		re := regexp.MustCompile(strconv.Itoa(PID))
		byteIndex := re.FindIndex([]byte(out))

		if (len(byteIndex) == 0) {
			ui.QueueMain(func () {
				CONNECTION = false
				label_Connection.SetText("  Connection: Offline" )
			})
		} else {
			// Update labels
			ui.QueueMain(func () {
				CONNECTION = true
				label_Connection.SetText("  Connection: online")
			})
		}

		// Wait x seconds before next iteration
		wait(config, label_Update, pb)
	}
}


// ---------------------------------------------------------------------------------------------------------------------
// A wrapper for time.Sleep() that also updates the UI label and progressbar
func wait(config Config, label_Update *ui.Label, pb *ui.ProgressBar) {
	tstep := config.TimeBetweenChecksInS
	if tstep <= 0 {
		tstep = 1
	} // otherwise division by 0
	for i := 0; i <= tstep; i++ {
		pb.SetValue(100/tstep * i)
		label_Update.SetText("  Next update in... " + strconv.Itoa(tstep - i) + " s")
		time.Sleep(1 * time.Second)
	}
	pb.SetValue(0)
}

// ---------------------------------------------------------------------------------------------------------------------
// Send a telegram message using a query URL
func send_TelegramMessage(config Config) {
	// Learn how to setup a telegram bot: https://core.telegram.org/bots
	resp, _ := http.Get("https://api.telegram.org/bot" + config.Botid +
		":" + config.Token +
		"/sendMessage?chat_id=" + config.Chatid +
		"&text=" + config.Message)
	defer resp.Body.Close()
}
