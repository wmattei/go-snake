package gamerunner

import (
	"fmt"
	"sync"
	"time"

	"github.com/wmattei/go-snake/lib/debugutil"
	"github.com/wmattei/go-snake/lib/encodingutil"
)

type Window struct {
	Width  int
	Height int
}

type gameRenderer struct {
	gameStateCh <-chan interface{}
	rawFrameCh  chan<- *encodingutil.Canvas
	closeSignal <-chan bool
	game        Game
	window      *Window
	debugger    *debugutil.Debugger
}

const numWorkers = 4

func (gr *gameRenderer) start() {

	var wg sync.WaitGroup
	workerCh := make(chan interface{})

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case gameState, ok := <-workerCh:
					if !ok {
						return
					}
					frame := gr.game.RenderFrame(&gameState, gr.window)
					if frame == nil {
						gr.debugger.ReportSkippedFrame()
						continue
					}
					gr.debugger.ReportRenderedCanvas()
					gr.rawFrameCh <- &encodingutil.Canvas{Data: frame, Timestamp: time.Now()}
				case _, ok := <-gr.closeSignal:
					if !ok {
						return
					}
				}
			}
		}()

	}

	for gameState := range gr.gameStateCh {
		workerCh <- gameState
	}

	fmt.Println("Closing game renderer")
	close(workerCh)
	wg.Wait()
	close(gr.rawFrameCh)

}
