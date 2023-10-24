package main

import (
	"github.com/wmattei/go-snake/shared/gamerunner"
)

type LatencyCheckGame struct{}

func (lcg *LatencyCheckGame) UpdateGameState(gs *interface{}, command interface{}, dt int64) {
	state, _ := (*gs).(*gameState)
	if command == nil {
		state.update(nil)
		return
	}
	x := command.(map[string]interface{})["position"].(map[string]interface{})["x"].(float64)
	y := command.(map[string]interface{})["position"].(map[string]interface{})["y"].(float64)
	state.update(&position{int(x), int(y)})
}

func (lcg *LatencyCheckGame) RenderFrame(gs *interface{}, window *gamerunner.Window) []byte {
	state, _ := (*gs).(*gameState)
	return renderFrame(state, window.Width, window.Height)
}

func (lcg *LatencyCheckGame) GetMetadata() *gamerunner.GameMetadata {
	return &gamerunner.GameMetadata{
		GameName: "Latency Check",
	}
}

func main() {
	game := &LatencyCheckGame{}
	gameRunner := gamerunner.NewGameRunner(game, nil)

	gameRunner.OnPlayerJoined(func(player *gamerunner.Player) {
		initialGameState := newGameState(player.Window.Height, player.Window.Width)
		gameRunner.StartEngine(initialGameState)
	})

	gameRunner.OpenLobby()
}
