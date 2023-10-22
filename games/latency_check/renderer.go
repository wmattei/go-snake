package latencycheck

import (
	"image/color"
	"time"

	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/shared/debugutil"
	"github.com/wmattei/go-snake/shared/encodingutil"
)

var (
	RED = color.RGBA{R: 255, G: 0, B: 0, A: 255}
)

const (
	bytesPerPixel = 3 // RGB: 3 bytes per pixel
	boxSize       = 40
)

func drawRectangle(rawRGBData []byte, col color.RGBA, idx, width int) {
	for i := 0; i < boxSize; i++ {
		for j := 0; j < boxSize; j++ {
			index := idx + (i*width+j)*bytesPerPixel
			if index >= len(rawRGBData) {
				continue
			}
			rawRGBData[index] = RED.R
			rawRGBData[index+1] = RED.G
			rawRGBData[index+2] = RED.B
		}
	}
}

func hasStateChanged(prev, curr *gameState) bool {
	if prev.mousePosition.X == curr.mousePosition.X && prev.mousePosition.Y == curr.mousePosition.Y {
		return false
	}
	return true
}

func renderFrame(gs *gameState, width, height int) []byte {
	// startedAt := time.Now()
	// defer logutil.LogTimeElapsed(startedAt, "Frame rendering")
	matrix := gs.GetMatrix()

	rawRGBData := make([]byte, bytesPerPixel*width*height)
	idx := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if matrix[y][x] == 1 {
				drawRectangle(rawRGBData, RED, idx, width)
			}
			idx += bytesPerPixel
		}
	}

	return rawRGBData
}

func startFrameRenderer(gameStateCh chan gameState, canvasCh chan<- *encodingutil.Canvas, width, height int, debugger *debugutil.Debugger) {
	var lastRenderedState *gameState
	statePerSecond := 0

	for {
		gameState := <-gameStateCh
		// fmt.Println(gameState.timeStamp)
		if lastRenderedState != nil && lastRenderedState.timeStamp.Second() == gameState.timeStamp.Second() {
			statePerSecond++
		} else {
			// fmt.Println("State per second:", statePerSecond)
			statePerSecond = 0
		}

		if lastRenderedState != nil && !hasStateChanged(lastRenderedState, &gameState) {
			continue
		}
		lastRenderedState = &gameState

		duration := gameState.timeStamp.Sub(time.Now())
		if duration > (time.Second / constants.FPS) {
			debugger.ReportSkippedFrame()
			continue
		}

		rawRGBData := renderFrame(&gameState, width, height)
		canvas := &encodingutil.Canvas{Data: rawRGBData, Timestamp: gameState.timeStamp}

		canvasCh <- canvas

	}
}
