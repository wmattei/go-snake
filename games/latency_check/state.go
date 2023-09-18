package latencycheck

type gameState struct {
	matrix        [][]int
	mousePosition position
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
		matrix: matrix,
	}
}

func (gs *gameState) handleCommand(command position) bool {
	gs.setAt(command, 1)
	gs.mousePosition = command
	return true
}
