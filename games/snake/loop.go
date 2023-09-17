package snake

import (
	"time"

	"github.com/wmattei/go-snake/constants"
)

type gameLoop struct {
	gameState      *GameState
	commandChannel chan string
	gameStateCh    chan *GameState
	closeSignal    chan bool
	frameTicker    *time.Ticker
}

type gameLoopInit struct {
	CommandChannel   chan string
	GameStateChannel chan *GameState
	CloseSignal      chan bool
}

func newGameLoop(options *gameLoopInit) *gameLoop {
	return &gameLoop{
		gameState:      newGameState(constants.ROWS, constants.COLS),
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
			gl.handleCommand(command)

		case <-gl.frameTicker.C:
			gl.updateGameState(nil)
			gl.gameStateCh <- gl.gameState
		}
	}
}

func (gl *gameLoop) close() {
	gl.closeSignal <- true
}

func (gl *gameLoop) handleCommand(command string) error {
	return gl.updateGameState(&command)
}

func (gl *gameLoop) updateGameState(command *string) error {
	gameOver := !gl.gameState.handleCommand(command)
	if gameOver {
		gl.close()
	}
	return nil
}
