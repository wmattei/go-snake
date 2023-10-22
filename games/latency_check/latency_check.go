package latencycheck

import (
	"github.com/wmattei/go-snake/shared/debugutil"
	"github.com/wmattei/go-snake/shared/encodingutil"
)

type LatencyCheckInit struct {
	WindowWidth    int
	WindowHeight   int
	CommandChannel <-chan interface{}          // Read only channel
	CanvasChannel  chan<- *encodingutil.Canvas // Write only channel
	CloseSignal    chan bool                   // 2 way channel, the game can be closed from the outside (e.g) when the WebRTC connection ends, or from the inside (e.g) when the game is over
	Debugger       *debugutil.Debugger
}

type any map[string]interface{}

func StartLatencyCheck(options *LatencyCheckInit) {
	// We create a buffered channel that allows up to half a second of unprocessed frames
	gameStateCh := make(chan gameState)
	posCommandCh := make(chan position)

	// Convert interface channel into position channel
	go func() {
		for {
			// I don't want to json.Unmarshal again. So I am doing this crazy shit here lol. Not sure if it's better
			command := <-options.CommandChannel
			x := command.(map[string]interface{})["position"].(map[string]interface{})["x"].(float64)
			y := command.(map[string]interface{})["position"].(map[string]interface{})["y"].(float64)
			posCommandCh <- position{X: int(x), Y: int(y)}
		}
	}()

	gameLoopOptions := &gameLoopInit{
		CommandChannel:   posCommandCh,
		CloseSignal:      options.CloseSignal,
		GameStateChannel: gameStateCh,
		Width:            options.WindowWidth,
		Height:           options.WindowHeight,
	}
	gameLoop := newGameLoop(gameLoopOptions)

	go gameLoop.start()
	go startFrameRenderer(gameStateCh, options.CanvasChannel, options.WindowWidth, options.WindowHeight, options.Debugger)

	// Wait for close signal
	<-options.CloseSignal
	close(gameStateCh)

}
