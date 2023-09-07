package game

import (
	"fmt"
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
		case command := <-commandChannel:
			fmt.Println("Updating game state with command:", command)
			if !updateGameState(gameState, &command) {
				closeSignal <- true
			}

		case <-frameTicker.C:
			handleFrameUpdate(gameState, frameChannel)

		case <-gameUpdateTicker.C:
			if !updateGameState(gameState, nil) {
				closeSignal <- true
			}
		}
	}
}

func handleFrameUpdate(gameState *GameState, frameChannel chan []byte) {
	if constants.SHOULD_RENDER_FRAME {
		frame := RenderFrame(*gameState)
		frameChannel <- frame
	}
}
