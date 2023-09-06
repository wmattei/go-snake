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

func (g *GameState) At(pos Position) int {
	return g.matrix[pos.Y][pos.X]
}

func (g *GameState) SetAt(pos Position, value int) {
	g.matrix[pos.Y][pos.X] = value
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

func (g *GameState) GenerateFood() {
	randomX := rand.Intn(g.cols)
	randomY := rand.Intn(g.rows)
	position := Position{X: randomX, Y: randomY}
	if g.At(position) != 0 {
		g.GenerateFood()
		return
	}

	g.foodPosition = Position{X: randomX, Y: randomY}
	g.SetAt(g.foodPosition, 3)
}

func (g *GameState) Move() bool {
	headPos := g.snake[0]
	g.SetAt(headPos, 2)
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

	if g.At(headPos) == 2 {
		return true
	}

	if headPos.Equal(g.foodPosition) {
		g.snakeSize++
		g.GenerateFood()
	}

	g.SetAt(headPos, 1)
	g.SetAt(g.LastSnakePart(), 0)
	g.snake = append([]Position{headPos}, g.snake[:g.snakeSize-1]...)

	return false
}

func validateNewDirection(state *GameState, newDirection Direction) bool {
	if state.IsHeadingUp() && newDirection == DOWN {
		return false
	}
	if state.IsHeadingDown() && newDirection == UP {
		return false
	}
	if state.IsHeadingLeft() && newDirection == RIGHT {
		return false
	}
	if state.IsHeadingRight() && newDirection == LEFT {
		return false
	}
	return true
}

func updateGameState(state *GameState, command *string) bool {
	if command != nil && validateNewDirection(state, Direction(*command)) {
		state.snakeDirection = Direction(*command)
	}

	collided := state.Move()

	return !collided
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

	state.GenerateFood()

	for i := 1; i < state.snakeSize; i++ {
		bodyPos := Position{Y: headPos.Y, X: headPos.X - i}
		snake = append(snake, bodyPos)
		state.SetAt(bodyPos, 2)
	}
	state.snake = snake
	state.SetAt(headPos, 1)

	return state
}
