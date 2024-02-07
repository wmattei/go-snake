package gamerunner

import (
	"encoding/json"
)

type inputState struct {
	width               int
	pressedKeys         map[string]bool
	pressedMouseButtons map[int]bool
	height              int
	mousePosition       [2]int

	prevState *inputState
}

func NewInputState() *inputState {
	return &inputState{
		pressedKeys:         make(map[string]bool),
		pressedMouseButtons: make(map[int]bool),
	}
}

func (is *inputState) setPrevState() {
	prevState := &inputState{
		width:  is.width,
		height: is.height,

		pressedMouseButtons: make(map[int]bool),
	}
	for button, pressed := range is.pressedMouseButtons {
		prevState.pressedMouseButtons[button] = pressed
	}
	is.prevState = prevState
}

func (is *inputState) IsKeyPressed(key string) bool {
	return is.pressedKeys[key]
}

func (is *inputState) IsMouseButtonPressed(button int) bool {
	return is.pressedMouseButtons[button]
}

func (is *inputState) IsMouseButtonJustPressed(button int) bool {
	return is.pressedMouseButtons[button] && !is.prevState.pressedMouseButtons[button]
}

func (is *inputState) GetMousePosition() (int, int) {
	return is.mousePosition[0], is.mousePosition[1]
}

func (is *inputState) handleCommand(command *Command) {
	switch command.Type {
	case Resize:
		data := ResizeData{}
		json.Unmarshal(command.Data, &data)

		is.height = data.Height
		is.width = data.Width
	case MouseClick:
		data := MouseClickData{}
		json.Unmarshal(command.Data, &data)

		is.pressedMouseButtons[data.Button] = true
	case MouseRelease:
		data := MouseReleaseData{}
		json.Unmarshal(command.Data, &data)

		is.pressedMouseButtons[data.Button] = false
	case MouseMove:
		data := MouseMoveData{}
		json.Unmarshal(command.Data, &data)

		is.mousePosition = [2]int{data.X, data.Y}
	}

}
