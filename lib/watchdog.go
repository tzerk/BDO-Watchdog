package lib

import (
	"os/exec"
	"syscall"
	"log"
	"regexp"
	"strconv"
	"github.com/andlabs/ui"
	"github.com/mitchellh/go-ps"
	"os"
	"time"
)

var STATUS bool = false
var CONNECTION bool = false
var PID int
var PENALTY int


//--------------------------------------------------------------------------------------------------------------
// PROCESS
//--------------------------------------------------------------------------------------------------------------
func Watchdog(
	config Config,
	label_Status *ui.Label,
	label_PID *ui.Label,
	label_Connection *ui.Label,
	label_Update *ui.Label,
	pb *ui.ProgressBar) {

	// KILL CoherentUI_Host.exe
	if config.KillCoherentUI {
		err := killCoherentUI()
		if err != nil {
			label_Status.SetText(err.Error())
		}
	}

	// INFINITE MAIN LOOP
	for {
		label_Update.SetText("")

		//// EXIT CONDITION
		//-----------------
		// If the process is running, but no longer connected we trigger the following actions
		if STATUS && !CONNECTION {

			// Only procede with exit routine if we reached the fail threshold
			if PENALTY >= config.FailLimit {
				// Use the Telegram API to send a message
				Send_TelegramMessage(config)

				// Optional: shutdown the computer if the monitored process is disconnected
				if config.ShutdownOnDC {
					exec.Command("cmd", "/C", "shutdown", "/s").Run()
				}

				// Optional: kill the monitored process if it is disconnected
				// requires elevated rights --> start .exe as administrator
				if config.KillOnDC {

					label_Update.SetText("  Trying to kill PID " + strconv.Itoa(PID))
					time.Sleep(5 * time.Second)

					defer func() {
						if r := recover(); r != nil {
							log.Println("  Panicked while trying to kill the process.")
						}
					}()

					proc, err := os.FindProcess(PID)
					if err != nil {
						label_Update.SetText("  Error: " + err.Error())
						log.Println(err)
					}

					// Kill the process
					err = proc.Kill()
					if err != nil {
						label_Update.SetText("  Error: " + err.Error())
						log.Println(err)
					}

					time.Sleep(5 * time.Second)
				}

				// Optional (YAML file, default: false): keep ts program open even if
				// the process is disconnected
				if !config.StayAlive {
					os.Exit(1)
				}

			// The process is running and disconnected, but we haven't reached the threshold yet;
			// Hence, we increase the penalty counter.
			} else {
				PENALTY += 1
			}

		}

		// Reset the penalty counter if process is running and disconnected
		if CONNECTION {
			PENALTY = 0
		}

		//// PROCESS
		//----------
		p, err := ps.Processes()
		if err != nil {
			log.Fatal(err)
		}

		//// PID
		//------
		PID = 0
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

			wait(config, label_Update, pb, PENALTY)
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
		wait(config, label_Update, pb, PENALTY)
	}
}