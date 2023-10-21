package debugutil

import (
	"fmt"
	"time"

	"github.com/wmattei/go-snake/constants"
)

type Debugger struct {
	frameCounter  int
	debugInterval int
}

func (d *Debugger) ReportFrameStream() {
	d.frameCounter++
}

func NewDebugger(intervalInMs int) *Debugger {
	return &Debugger{
		debugInterval: intervalInMs,
	}
}

func (d *Debugger) StartDebugger() {
	ticker := time.NewTicker(time.Duration(d.debugInterval) * time.Millisecond)
	defer ticker.Stop()
	defer fmt.Println("")

	for {
		<-ticker.C
		d.printFps()
	}
}

func (d *Debugger) printFps() {
	frameRate := float64(d.frameCounter) * (1000 / float64(d.debugInterval))
	var colorCode string
	if frameRate < constants.FPS*0.90 {
		colorCode = "\x1b[31m"
	} else {
		colorCode = "\x1b[32m"
	}

	resetColor := "\x1b[0m"

	fmt.Printf("\r%sFrame rate: %v / %v FPS%s ", colorCode, frameRate, constants.FPS, resetColor)
	d.frameCounter = 0
}
