package game

import (
	"time"

	"github.com/wmattei/go-snake/constants"
)

var (
	gameUpdateRate = time.Second / constants.GAME_SPEED
	frameRate      = time.Second / constants.FPS
)

func StartGameLoop(frameChannel chan []byte, commandChannel chan string, closeSignal chan bool) {
	frameTicker := time.NewTicker(frameRate)
	defer frameTicker.Stop()

	gameUpdateTicker := time.NewTicker(gameUpdateRate)
	defer gameUpdateTicker.Stop()

	gameState := NewGameState(constants.ROWS, constants.COLS)

	for {
		select {
		case <-frameTicker.C:
			handleFrameUpdate(gameState, frameChannel)
		case <-gameUpdateTicker.C:
			handleGameLogicUpdate(gameState, commandChannel, closeSignal)
		}
	}
}

func handleFrameUpdate(gameState *GameState, frameChannel chan []byte) {
	if constants.SHOULD_RENDER_FRAME {
		frame := RenderFrame(*gameState)
		frameChannel <- frame
	}
}

func handleGameLogicUpdate(gameState *GameState, commandChannel chan string, closeSignal chan bool) {
	select {
	case command := <-commandChannel:
		if !updateGameState(gameState, &command) {
			closeSignal <- true
		}
	default:
		if !updateGameState(gameState, nil) {
			closeSignal <- true
		}
	}
}
