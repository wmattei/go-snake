package gamerunner

import (
	"fmt"
	"time"

	"github.com/wmattei/go-snake/constants"
)

type gameLoop struct {
	gameState      *interface{}
	closeSignal    <-chan bool
	game           Game
	gameStateCh    chan<- interface{}
	commandChannel chan interface{}
}

func (gl *gameLoop) start() {
	frameDuration := time.Second / constants.FPS
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-gl.closeSignal:
			if !ok {
				break
			}
		case command := <-gl.commandChannel:
			gl.game.UpdateGameState(gl.gameState, command, frameDuration.Milliseconds())
		case <-ticker.C:
			gl.game.UpdateGameState(gl.gameState, nil, frameDuration.Milliseconds())
			gl.gameStateCh <- *gl.gameState
		}
	}

	fmt.Println("Stopping game loop")
	close(gl.gameStateCh)
	close(gl.commandChannel)
}
