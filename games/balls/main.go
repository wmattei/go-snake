package main

import (
	"github.com/wmattei/go-snake/lib/gamerunner"
)

type BallsGame struct{}

func (*BallsGame) UpdateGameState(gs *interface{}, command interface{}, dt int64) {
	state := (*gs).(*gameState)
	state.update(dt)

}

func (*BallsGame) RenderFrame(gs *interface{}, window *gamerunner.Window) []byte {
	return renderFrame((*gs).(*gameState), window.Width, window.Height)
}

func (*BallsGame) GetMetadata() *gamerunner.GameMetadata {
	return &gamerunner.GameMetadata{
		GameName: "Balls",
	}
}

func main() {
	gameRunner := gamerunner.NewGameRunner(&BallsGame{}, nil)

	gameRunner.OnPlayerJoined(func(player *gamerunner.Player) {
		initialGameState := newGameState(player.Window.Width, player.Window.Height)
		gameRunner.StartEngine(initialGameState)
	})

	gameRunner.OpenLobby()
}
