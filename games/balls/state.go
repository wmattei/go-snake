package main

import (
	"github.com/wmattei/go-snake/games/balls/ball"
	"github.com/wmattei/go-snake/lib/artemisia"
	"golang.org/x/exp/rand"
)

var colors = []*artemisia.Color{
	&artemisia.Color{255, 0, 0},
	&artemisia.Color{0, 255, 0},
	&artemisia.Color{0, 0, 255},
	&artemisia.Color{255, 255, 0},
	&artemisia.Color{255, 0, 255},
	&artemisia.Color{0, 255, 255},
	&artemisia.Color{255, 255, 255},
}

type gameState struct {
	balls  []*ball.Ball
	height int
}

func (gs *gameState) update(dt int64) {
	aliveBalls := []*ball.Ball{}
	for _, ball := range gs.balls {
		if !ball.IsDead {
			ball.Update(dt)
			aliveBalls = append(aliveBalls, ball)
		}
	}
	gs.balls = aliveBalls
}

func (gs *gameState) newBall(x, y float64) {
	radius := rand.Intn(50) + 10
	colorIdx := rand.Intn(len(colors))

	gs.balls = append(gs.balls, ball.NewBall(int(x), int(y), float64(radius), gs.height-radius, colors[colorIdx]))
}

func newGameState(width, height int) *gameState {
	return &gameState{
		height: height,
	}
}
