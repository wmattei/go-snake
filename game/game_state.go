package game

type Direction string

type Position struct {
	X int
	Y int
}

const (
	UP    Direction = "UP"
	DOWN  Direction = "DOWN"
	LEFT  Direction = "LEFT"
	RIGHT Direction = "RIGHT"
)

type GameState struct {
	id             int64
	snakeDirection Direction
	snakeSize      int
	snakePosition  Position
	matrix         [][]int
}

func (g *GameState) At(pos Position) int {
	return g.matrix[pos.Y][pos.X]
}

func (g *GameState) SetAt(pos Position, value int) {
	g.matrix[pos.Y][pos.X] = value
}

func (g *GameState) IncrementId() {
	g.id++
}

func updateGameState(state *GameState, command *string) bool {
	state.IncrementId()
	if command != nil {
		state.snakeDirection = Direction(*command)
	}

	newPos := calculateNextPosition(state.snakePosition, state.snakeDirection, len(state.matrix), len(state.matrix[0]))
	state.SetAt(newPos, 1)
	state.SetAt(state.snakePosition, 0)

	state.snakePosition = newPos

	return true

	// switch state.matrix[newPos.Y][newPos.X] {
	// case 0: // Empty space
	// 	state.matrix[newPos.Y][newPos.X] = 1
	// 	moveSnakeBody(state, newPos) // Move the snake body including the new part
	// 	state.snakePosition = newPos
	// case 3: // Food
	// 	state.matrix[newPos.Y][newPos.X] = 1
	// 	state.snakeSize++
	// 	generateFood(state)
	// default:
	// 	return false // Snake collided with itself, game over
	// }

	// return true
}

func NewGameState(rows, cols int) *GameState {
	matrix := make([][]int, cols)
	for i := range matrix {
		matrix[i] = make([]int, rows)
	}
	snakePos := Position{X: cols / 2, Y: rows / 2}
	matrix[snakePos.Y][snakePos.X] = 1
	return &GameState{
		snakeDirection: RIGHT,
		snakeSize:      1,
		snakePosition:  snakePos,
		matrix:         matrix,
	}
}

func calculateNextPosition(currentPos Position, direction Direction, rows, cols int) Position {
	switch direction {
	case UP:
		return Position{X: currentPos.X, Y: (currentPos.Y - 1 + rows) % rows}
	case DOWN:
		return Position{X: currentPos.X, Y: (currentPos.Y + 1) % rows}
	case LEFT:
		return Position{X: (currentPos.X - 1 + cols) % cols, Y: currentPos.Y}
	case RIGHT:
		return Position{X: (currentPos.X + 1) % cols, Y: currentPos.Y}
	default:
		return currentPos
	}
}
