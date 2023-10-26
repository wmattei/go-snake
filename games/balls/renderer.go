package main

import (
	"github.com/wmattei/go-snake/lib/artemisia"
)

func renderFrame(gs *gameState, width, height int) []byte {
	canvas := artemisia.NewCanvas(width, height)
	for _, ball := range gs.balls {
		canvas.DrawCircle(ball.Position.X, ball.Position.Y, ball.Radius)
	}

	return canvas.GetBytes()
}
