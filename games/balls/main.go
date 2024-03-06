package main

import (
	"math/rand"

	"github.com/wmattei/go-snake/games/balls/ball"
	"github.com/wmattei/go-snake/lib/artemisia"
	"github.com/wmattei/go-snake/lib/gamerunner"
)

type BallsGame struct {
	dt    int64
	balls []*ball.Ball
}

var colors = []*artemisia.Color{
	{255, 0, 0},
	{255, 255, 0},
	{0, 255, 0},
	{0, 255, 255},
	{0, 0, 255},
	{255, 0, 255},
}

func (game *BallsGame) Update(ctx *gamerunner.GameContext) {
	game.dt++
	_, height := ctx.GetScreenBounds()
	if ctx.IsMouseButtonJustPressed(0) {
		radius := rand.Intn(50) + 10
		colorIdx := rand.Intn(len(colors))

		mouseX, mouseY := ctx.GetMousePosition()
		game.balls = append(game.balls, ball.NewBall(mouseX, mouseY, float64(radius), height-radius, colors[colorIdx]))
	}
	for _, ball := range game.balls {
		ball.Ground = height - int(ball.Radius)
		ball.Update(game.dt)
	}
}

func (game *BallsGame) RenderFrame(frame *artemisia.Frame) {
	for _, ball := range game.balls {
		frame.DrawCircle(ball.Position.X, ball.Position.Y, int(ball.Radius), ball.Color)
	}
}

func main() {
	game := &BallsGame{}
	gameRunner := gamerunner.NewGameRunner(game)

	gameRunner.RunAfterResize()
}
