package gameutil

type GameMetadata struct {
	WindowWidth  int
	WindowHeight int
	GameName     string
}

func NewGameMetadata(windowWidth, windowHeight int, gameName string) *GameMetadata {
	return &GameMetadata{
		WindowWidth:  windowWidth,
		WindowHeight: windowHeight,
		GameName:     gameName,
	}
}
