package debugutil

import (
	"fmt"
	"runtime"
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
		fps := d.getFpsStat()
		mem := d.getMemoryStats()
		cpu := d.getCPUStats()

		logStat(fmt.Sprintf("%s | %s | %s", fps, mem, cpu))
	}
}

func (d *Debugger) getMemoryStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Allocated memory: %v MB", m.Alloc/1024/1024)
}

func (d *Debugger) getCPUStats() string {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	return fmt.Sprintf("Goroutines: %v ", runtime.NumGoroutine())

}

func (d *Debugger) getFpsStat() string {
	frameRate := float64(d.frameCounter) * (1000 / float64(d.debugInterval))
	var colorCode string
	if frameRate < constants.FPS*0.90 {
		colorCode = "\x1b[31m"
	} else {
		colorCode = "\x1b[32m"
	}

	resetColor := "\x1b[0m"

	d.frameCounter = 0
	return fmt.Sprintf("%sFrame rate: %v / %v FPS%s ", colorCode, frameRate, constants.FPS, resetColor)
}

func logStat(stat string) {
	fmt.Printf("\r%v", stat)
}
