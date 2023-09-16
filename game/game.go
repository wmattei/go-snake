package game

import (
	"time"

	"github.com/wmattei/go-snake/constants"
)

type GameLoop struct {
	gameState      *GameState
	commandChannel chan string
	gameStateCh    chan *GameState
	closeSignal    chan bool
	frameTicker    *time.Ticker
}

type GameLoopInit struct {
	CommandChannel   chan string
	GameStateChannel chan *GameState
	CloseSignal      chan bool
}

func NewGameLoop(options *GameLoopInit) *GameLoop {
	return &GameLoop{
		gameState:      NewGameState(constants.ROWS, constants.COLS),
		commandChannel: options.CommandChannel,
		gameStateCh:    options.GameStateChannel,
		closeSignal:    options.CloseSignal,
		frameTicker:    time.NewTicker(time.Second / constants.FPS),
	}
}

func (gl *GameLoop) Start() {
	defer gl.frameTicker.Stop()

	for {
		select {
		case command := <-gl.commandChannel:
			gl.handleCommand(command)

		case <-gl.frameTicker.C:
			gl.updateGameState(nil)
			gl.gameStateCh <- gl.gameState
		}
	}
}

func (gl *GameLoop) Close() {
	gl.closeSignal <- true
}

func (gl *GameLoop) handleCommand(command string) error {
	return gl.updateGameState(&command)
}

func (gl *GameLoop) updateGameState(command *string) error {
	gameOver := !gl.gameState.handleCommand(command)
	if gameOver {
		gl.Close()
	}
	return nil
}
