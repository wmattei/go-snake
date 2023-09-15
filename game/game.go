package game

import (
	"time"

	"github.com/wmattei/go-snake/constants"
)

var (
	frameRate = time.Second / constants.FPS
)

func StartGameLoop(commandChannel chan string, gameStateCh chan *GameState, closeSignal chan bool) {
	frameTicker := time.NewTicker(frameRate)
	defer frameTicker.Stop()

	gameState := NewGameState(constants.ROWS, constants.COLS)

	for {
		select {
		case command := <-commandChannel:
			gameOver := !updateGameState(gameState, &command)
			if gameOver {
				frameTicker.Stop()
				closeSignal <- true
				break
			}
			gameStateCh <- gameState

		case <-frameTicker.C:
			gameOver := !updateGameState(gameState, nil)
			if gameOver {
				frameTicker.Stop()
				closeSignal <- true
				break
			}
			gameStateCh <- gameState
		}
	}
}
