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
	"strings"
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
	KillCoherentUI bool
}

// Variables
var STATUS bool = false
var CONNECTION bool = false
var PID int

func main() {

	////  SETTINGS
	//--------------------------------------------------------------------------------------------------------------
	// YAML PARSING
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	newex := strings.Replace(ex, "BDO-Watchdog.exe", "config.yml", -1)
	// This is necessary for dynamic builds in Jetbrains Gogland IDE
	//newex = strings.Replace(ex, "Application.exe", "config.yml", -1)

	var config Config
	source, err := ioutil.ReadFile(newex)

	if err != nil {
		// in theory, using yml.Marshal() would be more elegant, but we want to preserve the yaml comments
		// as well as set some default values/hints
		defconf :=
				"## Get updates here: https://github.com/tzerk/BDO-Watchdog/releases/\r\n" +
				"## Telegram Bot Settings\r\n" +
				"token: \r\n" +
				"botid: \r\n" +
				"chatid: \r\n" +
				"message: BDO disconnected \r\n" +
				"\r\n" +
				"## Program Settings\r\n" +
				"stayalive: false\r\n" +
				"process: BlackDesert64.exe\r\n" +
				"timebetweenchecksins: 60\r\n" +
				"\r\n" +
				"# These settings require the .exe to be run with admin rights! \r\n" +
				"killondc: true\r\n" +
				"shutdownondc: false\r\n" +
				"killcoherentui: false"
		ioutil.WriteFile("config.yml", []byte(defconf), os.FileMode(int(0777)))
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}


	//// GUI
	//--------------------------------------------------------------------------------------------------------------
	ui := ui.Main(func() {
		window := ui.NewWindow("BDO Watchdog v0.1.4", 300, 80, false)

		label_Process := ui.NewLabel("  Process: " + config.Process)
		label_Status := ui.NewLabel("  Initializing...")
		label_PID := ui.NewLabel("-")
		label_Connection := ui.NewLabel("-")
		label_Update := ui.NewLabel("")

		box := ui.NewVerticalBox()

		tbtn := ui.NewButton("Send Telegram test message")
		tbtn.OnClicked(func(*ui.Button) {
			send_TelegramMessage(config)
		})

		sep := ui.NewHorizontalSeparator()
		pb := ui.NewProgressBar()

		// Append all UI elements to the box container
		box.Append(label_Process, false)
		box.Append(label_Status, false)
		box.Append(label_PID, false)
		box.Append(label_Connection, false)
		box.Append(label_Update, false)
		box.Append(tbtn, false)
		box.Append(sep, false)
		box.Append(pb, true)

		// If this tool requires admin rights add a text label to inform the user
		if config.KillOnDC || config.KillCoherentUI {
			box.Append(ui.NewLabel("Make sure to have started this program with admin rights!"), true)
		}

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

	// KILL CoherentUI_Host.exe
	if config.KillCoherentUI {

		// Find process(es)
		chp, err := ps.Processes()
		if err != nil {
			log.Fatal(err)
		}

		// Find PID and kill
		for _, v := range chp {
			if v.Executable() == "CoherentUI_Host.exe" {
				proc, err := os.FindProcess(v.Pid())

				if err != nil {
					log.Println(err)
				}

				defer func() {
					recover() // 1
					return
				}()
				// Kill the process
				proc.Kill()
			}
		}
	}

	// INFINITE MAIN LOOP
	for {
		label_Update.SetText("")

		//// EXIT CONDITION
		//-----------------
		// If the process is running, but no longer connected we trigger the following actions
		if STATUS && !CONNECTION {

			// Use the Telegram API to send a message
			send_TelegramMessage(config)

			// Optional: shutdown the computer if the monitored process is disconnected
			if config.ShutdownOnDC {
				exec.Command("cmd", "/C", "shutdown", "/s").Run()
			}

			// Optional: kill the monitored process if it is disconnected
			// requires elevated rights --> start .exe as administrator
			if config.KillOnDC {

				proc, err := os.FindProcess(PID)

				if err != nil {
					log.Println(err)
				}

				defer func() {
					recover() // 1
					return
				}()
				// Kill the process
				proc.Kill()

				time.Sleep(5 * time.Second)
			}

			// Optional (YAML file, default: false): keep ts program open even if
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
	var pbVal int

	if tstep <= 0 {
		tstep = 1
	} // otherwise division by 0
	for i := 0; i <= tstep; i++ {
		pbVal = int(100/float32(tstep) * float32(i))
		if pbVal > 100 {
			pbVal = 100
		}
		pb.SetValue(pbVal)
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
