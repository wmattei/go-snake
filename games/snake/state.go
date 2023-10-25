package main

import "math/rand"

type direction string

type position struct {
	X int
	Y int
}

func (p *position) equal(other position) bool {
	return p.X == other.X && p.Y == other.Y
}

const (
	up    direction = "UP"
	down  direction = "DOWN"
	left  direction = "LEFT"
	right direction = "RIGHT"
)

type gameState struct {
	snakeDirection direction
	snakeSize      int
	snake          []position
	matrix         [][]int
	rows           int
	cols           int
	foodPosition   position
}

func newGameState(rows, cols int) *gameState {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
	}
	headPos := position{X: cols / 2, Y: rows / 2}
	snake := []position{headPos}

	state := &gameState{
		snakeDirection: right,
		snakeSize:      3,
		matrix:         matrix,
		rows:           rows,
		cols:           cols,
	}

	state.generateFood()

	for i := 1; i < state.snakeSize; i++ {
		bodyPos := position{Y: headPos.Y, X: headPos.X - i}
		snake = append(snake, bodyPos)
		state.setAt(bodyPos, 2)
	}
	state.snake = snake
	state.setAt(headPos, 1)

	return state
}

func (g *gameState) GetMatrix() [][]int {
	return g.matrix
}

func (g *gameState) setAt(pos position, value int) {
	g.matrix[pos.Y][pos.X] = value
}

func (g *gameState) at(pos position) int {
	return g.matrix[pos.Y][pos.X]
}

func (g *gameState) LastSnakePart() position {
	return g.snake[len(g.snake)-1]
}

func (g *gameState) IsHeadingUp() bool {
	return g.snakeDirection == up
}

func (g *gameState) IsHeadingDown() bool {
	return g.snakeDirection == down
}

func (g *gameState) IsHeadingLeft() bool {
	return g.snakeDirection == left
}

func (g *gameState) IsHeadingRight() bool {
	return g.snakeDirection == right
}

func (g *gameState) generateFood() {
	randomX := rand.Intn(g.cols)
	randomY := rand.Intn(g.rows)
	pos := position{X: randomX, Y: randomY}
	if g.at(pos) != 0 {
		g.generateFood()
		return
	}

	g.foodPosition = position{X: randomX, Y: randomY}
	g.setAt(g.foodPosition, 3)
}

func (g *gameState) validateNewDirection(newDirection direction) bool {
	if g.IsHeadingUp() && newDirection == down {
		return false
	}
	if g.IsHeadingDown() && newDirection == up {
		return false
	}
	if g.IsHeadingLeft() && newDirection == right {
		return false
	}
	if g.IsHeadingRight() && newDirection == left {
		return false
	}
	return true
}

func (g *gameState) updateGameState(command *string) bool {
	if command != nil && g.validateNewDirection(direction(*command)) {
		g.snakeDirection = direction(*command)
	}

	collided := g.move()
	return !collided
}

func (g *gameState) move() bool {
	headPos := g.snake[0]
	g.setAt(headPos, 2)
	switch g.snakeDirection {
	case up:
		headPos.Y = (headPos.Y - 1 + g.rows) % g.rows
	case down:
		headPos.Y = (headPos.Y + 1) % g.rows
	case left:
		headPos.X = (headPos.X - 1 + g.cols) % g.cols
	case right:
		headPos.X = (headPos.X + 1) % g.cols
	default:
		return true
	}

	if g.at(headPos) == 2 {
		return true
	}

	if headPos.equal(g.foodPosition) {
		g.snakeSize++
		g.generateFood()
	}

	g.setAt(headPos, 1)
	g.setAt(g.LastSnakePart(), 0)
	g.snake = append([]position{headPos}, g.snake[:g.snakeSize-1]...)

	return false
}

func (g *gameState) handleCommand(command *string) bool {
	return g.updateGameState(command)
}
