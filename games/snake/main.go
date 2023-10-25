package main

import (
	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/lib/gamerunner"
)

type SnakeGame struct {
	lastRenderedState *gameState
}

func (lcg *SnakeGame) UpdateGameState(gs *interface{}, command interface{}, dt int64) {
	state, _ := (*gs).(*gameState)
	if command == nil {
		state.updateGameState(nil)
		return
	}

	if command != nil {
		newDir := command.(map[string]interface{})["dir"].(string)
		state.updateGameState(&newDir)
		return
	}

	state.updateGameState(nil)
}

func (lcg *SnakeGame) RenderFrame(gs *interface{}, window *gamerunner.Window) []byte {
	state, _ := (*gs).(*gameState)

	return renderFrame(state, window.Width, window.Height)
}

func (lcg *SnakeGame) GetMetadata() *gamerunner.GameMetadata {
	return &gamerunner.GameMetadata{
		GameName: "Snake",
	}
}

func main() {
	game := &SnakeGame{}
	gameRunner := gamerunner.NewGameRunner(game, nil)

	gameRunner.OnPlayerJoined(func(player *gamerunner.Player) {
		initialGameState := newGameState(player.Window.Height/constants.CHUNK_SIZE, player.Window.Width/constants.CHUNK_SIZE)
		gameRunner.StartEngine(initialGameState)
	})

	gameRunner.OpenLobby()
}
