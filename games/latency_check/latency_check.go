package latencycheck

type LatencyCheckInit struct {
	WindowWidth    int
	WindowHeight   int
	CommandChannel <-chan interface{} // Read only channel
	FrameChannel   chan<- []byte      // Write only channel
	CloseSignal    chan bool          // 2 way channel, the game can be closed from the outside (e.g) when the WebRTC connection ends, or from the inside (e.g) when the game is over
}

type any map[string]interface{}

func StartLatencyCheck(options *LatencyCheckInit) {
	gameStateCh := make(chan gameState, 1)
	posCommandCh := make(chan position, 1)

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
		Width:            options.WindowWidth, Height: options.WindowHeight,
	}
	gameLoop := newGameLoop(gameLoopOptions)

	go gameLoop.start()
	go startFrameRenderer(gameStateCh, options.FrameChannel, options.WindowWidth, options.WindowHeight)

	// Wait for close signal
	<-options.CloseSignal
	close(gameStateCh)

}
