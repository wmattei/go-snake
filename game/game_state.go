package game

import "math/rand"

type Direction string

type Position struct {
	X int
	Y int
}

func (p *Position) Equal(other Position) bool {
	return p.X == other.X && p.Y == other.Y
}

const (
	UP    Direction = "UP"
	DOWN  Direction = "DOWN"
	LEFT  Direction = "LEFT"
	RIGHT Direction = "RIGHT"
)

type GameState struct {
	snakeDirection Direction
	snakeSize      int
	snake          []Position
	matrix         [][]int
	rows           int
	cols           int
	foodPosition   Position
}

func NewGameState(rows, cols int) *GameState {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
	}
	headPos := Position{X: cols / 2, Y: rows / 2}
	snake := []Position{headPos}

	state := &GameState{
		snakeDirection: RIGHT,
		snakeSize:      3,
		matrix:         matrix,
		rows:           rows,
		cols:           cols,
	}

	state.generateFood()

	for i := 1; i < state.snakeSize; i++ {
		bodyPos := Position{Y: headPos.Y, X: headPos.X - i}
		snake = append(snake, bodyPos)
		state.setAt(bodyPos, 2)
	}
	state.snake = snake
	state.setAt(headPos, 1)

	return state
}

func (g *GameState) GetMatrix() [][]int {
	return g.matrix
}

func (g *GameState) setAt(pos Position, value int) {
	g.matrix[pos.Y][pos.X] = value
}

func (g *GameState) at(pos Position) int {
	return g.matrix[pos.Y][pos.X]
}

func (g *GameState) LastSnakePart() Position {
	return g.snake[len(g.snake)-1]
}

func (g *GameState) IsHeadingUp() bool {
	return g.snakeDirection == UP
}

func (g *GameState) IsHeadingDown() bool {
	return g.snakeDirection == DOWN
}

func (g *GameState) IsHeadingLeft() bool {
	return g.snakeDirection == LEFT
}

func (g *GameState) IsHeadingRight() bool {
	return g.snakeDirection == RIGHT
}

func (g *GameState) generateFood() {
	randomX := rand.Intn(g.cols)
	randomY := rand.Intn(g.rows)
	position := Position{X: randomX, Y: randomY}
	if g.at(position) != 0 {
		g.generateFood()
		return
	}

	g.foodPosition = Position{X: randomX, Y: randomY}
	g.setAt(g.foodPosition, 3)
}

func (g *GameState) validateNewDirection(newDirection Direction) bool {
	if g.IsHeadingUp() && newDirection == DOWN {
		return false
	}
	if g.IsHeadingDown() && newDirection == UP {
		return false
	}
	if g.IsHeadingLeft() && newDirection == RIGHT {
		return false
	}
	if g.IsHeadingRight() && newDirection == LEFT {
		return false
	}
	return true
}

func (g *GameState) updateGameState(command *string) bool {
	if command != nil && g.validateNewDirection(Direction(*command)) {
		g.snakeDirection = Direction(*command)
	}

	collided := g.move()
	return !collided
}

func (g *GameState) move() bool {
	headPos := g.snake[0]
	g.setAt(headPos, 2)
	switch g.snakeDirection {
	case UP:
		headPos.Y = (headPos.Y - 1 + g.rows) % g.rows
	case DOWN:
		headPos.Y = (headPos.Y + 1) % g.rows
	case LEFT:
		headPos.X = (headPos.X - 1 + g.cols) % g.cols
	case RIGHT:
		headPos.X = (headPos.X + 1) % g.cols
	default:
		return true
	}

	if g.at(headPos) == 2 {
		return true
	}

	if headPos.Equal(g.foodPosition) {
		g.snakeSize++
		g.generateFood()
	}

	g.setAt(headPos, 1)
	g.setAt(g.LastSnakePart(), 0)
	g.snake = append([]Position{headPos}, g.snake[:g.snakeSize-1]...)

	return false
}

func (g *GameState) handleCommand(command *string) bool {
	return g.updateGameState(command)
}
