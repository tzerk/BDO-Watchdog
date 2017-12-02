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
var ESTABLISHEDONCE bool = false
var EXIT bool = false


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

	// SET Process Priority
	if (config.ProcessPriority != "normal") {
		exec.Command("cmd", "/C", "wmic", "process",  "where", "name='" + config.Process + "'", "CALL", "setpriority", config.ProcessPriority).Run()
	}

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

		// Two exit conditions:
		// a) The process is running, but no longer connected
		// b) The process was online once, but killed
		if STATUS && !CONNECTION {
			EXIT = true
		}
		if !STATUS && !CONNECTION && ESTABLISHEDONCE {
			EXIT = true
		}

		// If the process is running, but no longer connected we trigger the following actions
		if EXIT {

			// Only proceed with exit routine if we reached the fail threshold
			if PENALTY >= config.FailLimit {

				// Optional: shutdown the computer if the monitored process is disconnected
				if config.ShutdownOnDC {
					exec.Command("cmd", "/C", "shutdown", "/s").Run()
				}

				// Optional: write disconnect msg to log file
				if config.Log {
					logger("Process (PID " + strconv.Itoa(PID) + ") disconnected\r\n")
				}

				// Optional: kill the monitored process if it is disconnected
				// requires elevated rights --> start .exe as administrator
				if config.KillOnDC && PID != 0 {

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

				// Use the Telegram API to send a message
				Send_TelegramMessage(config, label_Update, pb)

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

		// If the process was observed at least once and was also connected, we also want to notify the user
		// when the process unexpectedly crashed. We need this to differentiate this case with the case where
		// BDO-Watchdog was started BEFORE the observed process. We set this flag only once.
		if STATUS && CONNECTION && !ESTABLISHEDONCE {
			ESTABLISHEDONCE = true
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
				CONNECTION = false
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