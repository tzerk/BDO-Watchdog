package lib

import (
	"github.com/mitchellh/go-ps"
	"log"
	"os"
	"errors"
)

func killCoherentUI() (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Unable to kill 'CoherentUI_Host.exe', need admin rights!'")
		}
	}()

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

	return err
}
