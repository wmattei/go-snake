package debugutil

import (
	"fmt"
	"runtime"
	"time"

	"github.com/wmattei/go-snake/constants"
)

const RESET_COLOR = "\x1b[0m"

type Debugger struct {
	frameCounter        int
	droppedFrameCounter int
	skippedFrameCounter int
	debugInterval       int
}

func (d *Debugger) ReportFrameStream() {
	d.frameCounter++
}

func (d *Debugger) ReportDroppedFrame() {
	d.droppedFrameCounter++
}

func (d *Debugger) ReportSkippedFrame() {
	d.skippedFrameCounter++
}

func NewDebugger(intervalInMs int) *Debugger {
	return &Debugger{
		debugInterval: intervalInMs,
	}
}

func (d *Debugger) StartDebugger() {
	ticker := time.NewTicker(time.Duration(d.debugInterval) * time.Millisecond)
	defer ticker.Stop()

	lines := make([]string, 6) // Store the lines of statistics

	for {
		<-ticker.C

		memStats := d.getMemoryStats()
		lines[0] = d.getFpsStat()
		lines[1] = d.getDroppedFramesStat()
		lines[2] = d.getSkippedFrameCounter()
		lines[3] = memStats[0]
		lines[4] = memStats[1]
		lines[5] = d.getCPUStats()

		clearScreen()
		for _, line := range lines {
			fmt.Println(line)
		}
	}
}

func (d *Debugger) getMemoryStats() [2]string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	orangeCode := "\x1b[33m"
	alloc := fmt.Sprintf("Memory Allocated: %s%v MB%s", orangeCode, m.Alloc/1024/1024, RESET_COLOR)
	heapAlloc := fmt.Sprintf("Heap: %s%v MB%s", orangeCode, m.HeapAlloc/1024/1024, RESET_COLOR)

	return [2]string{alloc, heapAlloc}
}

func (d *Debugger) getCPUStats() string {
	orangeCode := "\x1b[33m"
	return fmt.Sprintf("Goroutines: %s%v%s", orangeCode, runtime.NumGoroutine(), RESET_COLOR)
}

func (d *Debugger) getFpsStat() string {
	frameRate := float64(d.frameCounter) * (1000 / float64(d.debugInterval))
	var colorCode string
	if frameRate < constants.FPS*0.90 {
		colorCode = "\x1b[31m"
	} else {
		colorCode = "\x1b[32m"
	}

	d.frameCounter = 0
	return fmt.Sprintf("Frame rate: %s%.2f%s / %v FPS", colorCode, frameRate, RESET_COLOR, constants.FPS)
}

func (d *Debugger) getDroppedFramesStat() string {
	droppedFrameRate := float64(d.droppedFrameCounter) * (1000 / float64(d.debugInterval))
	var colorCode string
	if droppedFrameRate > constants.FPS*0.10 {
		colorCode = "\x1b[31m"
	} else {
		colorCode = "\x1b[32m"
	}

	resetColor := "\x1b[0m"

	d.droppedFrameCounter = 0
	return fmt.Sprintf("Dropped frames: %s%v%s", colorCode, droppedFrameRate, resetColor)
}

func (d *Debugger) getSkippedFrameCounter() string {
	skippedFrameRate := float64(d.skippedFrameCounter) * (1000 / float64(d.debugInterval))
	var colorCode string
	if skippedFrameRate > constants.FPS*0.05 {
		colorCode = "\x1b[31m"
	} else {
		colorCode = "\x1b[32m"
	}

	resetColor := "\x1b[0m"

	d.skippedFrameCounter = 0
	return fmt.Sprintf("Skipped frames: %s%v%s", colorCode, skippedFrameRate, resetColor)
}

func clearScreen() {
	fmt.Print("\x1b[2J\x1b[H") // ANSI escape codes to clear the screen and move the cursor to the top-left corner
}
