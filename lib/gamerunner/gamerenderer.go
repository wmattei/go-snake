package gamerunner

import (
	"time"

	"github.com/wmattei/go-snake/lib/artemisia"
	"github.com/wmattei/go-snake/lib/debugutil"
	"github.com/wmattei/go-snake/lib/encodingutil"
)

type gameRenderer struct {
	rawFrameCh chan<- *encodingutil.Canvas

	// closeSignal <-chan bool
	game     Game
	debugger *debugutil.Debugger
}

func newGameRenderer(game Game, rawFrameCh chan<- *encodingutil.Canvas) *gameRenderer {
	return &gameRenderer{
		game:       game,
		rawFrameCh: rawFrameCh,
	}
}

func (gr *gameRenderer) render(width, height int) {
	frame := artemisia.NewFrame(width, height)
	gr.game.RenderFrame(frame)

	if gr.debugger != nil {
		gr.debugger.ReportRenderedCanvas()
	}

	bytes := frame.Bytes()
	gr.rawFrameCh <- &encodingutil.Canvas{Data: bytes, Timestamp: time.Now()}
}
