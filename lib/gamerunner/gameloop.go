package gamerunner

import (
	"time"

	"github.com/wmattei/go-snake/constants"
)

type gameLoop struct {
	closeSignal  <-chan bool
	game         Game
	gameContext  *GameContext
	gameRenderer *gameRenderer
}

func (gl *gameLoop) start() {
	frameDuration := time.Second / constants.FPS
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-gl.closeSignal:
			if !ok {
				return
			}
		case <-ticker.C:

			gl.game.Update(gl.gameContext)
			gl.gameRenderer.render(gl.gameContext.width, gl.gameContext.height)
			gl.gameContext.inputState.setPrevState()

		}
	}
}
