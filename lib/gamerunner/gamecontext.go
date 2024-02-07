package gamerunner

type GameContext struct {
	*inputState
}

func NewGameContext() *GameContext {
	return &GameContext{
		inputState: NewInputState(),
	}
}

func (gc *GameContext) GetScreenBounds() (int, int) {
	return gc.width, gc.height
}
