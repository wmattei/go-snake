package latencycheck

import (
	"time"

	"github.com/wmattei/go-snake/constants"
)

type gameLoop struct {
	gameState      *gameState
	commandChannel chan position
	gameStateCh    chan<- gameState
	closeSignal    chan bool
	frameTicker    *time.Ticker
}

type gameLoopInit struct {
	CommandChannel   chan position
	GameStateChannel chan<- gameState
	CloseSignal      chan bool
	Width            int
	Height           int
}

func newGameLoop(options *gameLoopInit) *gameLoop {
	return &gameLoop{
		gameState:      newGameState(options.Height, options.Width),
		commandChannel: options.CommandChannel,
		gameStateCh:    options.GameStateChannel,
		closeSignal:    options.CloseSignal,
		frameTicker:    time.NewTicker(time.Second / constants.FPS),
	}
}

func (gl *gameLoop) start() {
	defer gl.frameTicker.Stop()

	for {
		select {
		case <-gl.closeSignal:
			return
		case command := <-gl.commandChannel:
			gl.updateGameState(command)
		case <-gl.frameTicker.C:
			gl.gameStateCh <- *gl.gameState
		}
	}
}

func (gl *gameLoop) close() {
	gl.closeSignal <- true
}

func (gl *gameLoop) updateGameState(command position) {
	gl.gameState.handleCommand(command)
}
