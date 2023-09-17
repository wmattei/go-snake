package snake

func StartSnakeGame(commandChannel chan string, frameChannel chan []byte, closeSignal chan bool) {
	gameStateCh := make(chan *GameState, 1)
	gameLoop := newGameLoop(&gameLoopInit{CommandChannel: commandChannel, CloseSignal: closeSignal, GameStateChannel: gameStateCh})

	go gameLoop.start()
	go startFrameRenderer(gameStateCh, frameChannel)

	// Wait for close signal
	<-closeSignal
	close(gameStateCh)
}
