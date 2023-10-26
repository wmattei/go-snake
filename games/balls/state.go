package main

import "github.com/wmattei/go-snake/games/balls/ball"

type gameState struct {
	balls []*ball.Ball
}

func (gs *gameState) update(dt int64) {
	for _, ball := range gs.balls {
		ball.Update(dt)
	}
}

func newGameState(width, height int) *gameState {
	radius := 50
	ball1 := ball.NewBall(500, radius, radius, height-radius)
	// ball2 := ball.NewBall(500, 100, 0)
	return &gameState{
		balls: []*ball.Ball{ball1},
	}
}
