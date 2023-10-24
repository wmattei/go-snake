package main

import (
	"time"
)

type position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type gameState struct {
	matrix        [][]int
	mousePosition position
	timeStamp     time.Time
}

func (g *gameState) isEqual(other *gameState) bool {
	if g.mousePosition.X == other.mousePosition.X && g.mousePosition.Y == other.mousePosition.Y {
		return true
	}
	return false
}

func (g *gameState) setAt(pos position, value int) {
	g.matrix[pos.Y][pos.X] = value
}

func (g *gameState) GetMatrix() [][]int {
	return g.matrix
}

func newGameState(height, width int) *gameState {
	matrix := make([][]int, height)
	for i := range matrix {
		matrix[i] = make([]int, width)
	}
	gs := &gameState{
		matrix:        matrix,
		mousePosition: position{50, 50},
	}
	return gs
}

func (gs *gameState) update(command *position) bool {
	gs.timeStamp = time.Now()
	if command == nil {
		return true
	}

	gs.setAt(gs.mousePosition, 0)
	gs.setAt(*command, 1)
	gs.mousePosition = *command

	return true
}
