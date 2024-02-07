package gamerunner

import "encoding/json"

type CommandType int

const (
	// Lobby
	Join CommandType = iota
	Quit
	Pause
	Resume
	Reset

	// Mouse
	MouseMove // 5
	MouseClick
	MouseRelease

	// Window
	Resize // 8
)

type CommandData interface{}
type MouseMoveData struct {
	X int
	Y int
}

type MouseClickData struct {
	Button int
}

type MouseReleaseData struct {
	Button int
}

type ResizeData struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Command struct {
	Type CommandType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// func (c *Command) UnmarshalJSON(data []byte) error {
// 	var tmp struct {
// 		Type CommandType            `json:"type"`
// 		Data map[string]interface{} `json:"data"`
// 	}
// 	if err := json.Unmarshal(data, &tmp); err != nil {
// 		return err
// 	}

// 	c.Type = tmp.Type

// 	switch c.Type {
// 	case Resize:
// 		var resizeData ResizeData
// 		dataBytes, err := json.Marshal(tmp.Data)
// 		if err != nil {
// 			return err
// 		}
// 		if err := json.Unmarshal(dataBytes, &resizeData); err != nil {
// 			return err
// 		}
// 		c.Data = resizeData
// 	default:
// 		return errors.New("unsupported command type")
// 	}
// 	return nil
// }
