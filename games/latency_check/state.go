package latencycheck

import (
	"time"
)

type gameState struct {
	matrix        [][]int
	mousePosition position
	timeStamp     time.Time
}

func (g *gameState) setAt(pos position, value int) {
	g.matrix[pos.Y][pos.X] = value
}

func (g *gameState) GetMatrix() [][]int {
	return g.matrix
}

type position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func newGameState(rows, cols int) *gameState {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
	}
	return &gameState{
		matrix:    matrix,
		timeStamp: time.Now(),
	}
}

func (gs *gameState) handleCommand(command *position) bool {
	gs.timeStamp = time.Now()
	if command == nil {
		return true
	}
	gs.setAt(gs.mousePosition, 0)
	gs.setAt(*command, 1)
	gs.mousePosition = *command

	return true
}
