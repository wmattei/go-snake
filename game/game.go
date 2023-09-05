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
		case <-closeSignal:
			return // Exiting the loop on close signal
		case <-frameTicker.C:
			handleFrameUpdate(gameState, frameChannel)
		case <-gameUpdateTicker.C:
			handleGameLogicUpdate(gameState, commandChannel)
		}
	}
}

func handleFrameUpdate(gameState *GameState, frameChannel chan []byte) {
	if constants.SHOULD_RENDER_FRAME {
		go RenderFrame(*gameState, frameChannel)
	}
}

func handleGameLogicUpdate(gameState *GameState, commandChannel chan string) {
	select {
	case command := <-commandChannel:
		updateGameState(gameState, &command)
	default:
		updateGameState(gameState, nil)
	}
}
