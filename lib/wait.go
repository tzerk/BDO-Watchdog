package lib

import (
	"github.com/andlabs/ui"
	"strconv"
	"time"
)

// ---------------------------------------------------------------------------------------------------------------------
// A wrapper for time.Sleep() that also updates the UI label and progressbar
func wait(config Config, label_Update *ui.Label, pb *ui.ProgressBar, penalty int) {
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
		label_Update.SetText("  Next update in... " + strconv.Itoa(tstep - i) + " s" +
			" (failed checks: " + strconv.Itoa(penalty) + "/" + strconv.Itoa(config.FailLimit) + ")")
		time.Sleep(1 * time.Second)
	}
	pb.SetValue(0)
}