package game

import (
	"image"
	"image/color"

	"github.com/wmattei/go-snake/constants"
)

func drawRectangle(img *image.RGBA, min, max image.Point, col color.RGBA) {
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			img.Set(x, y, col)
		}
	}
}

func RenderFrame(gameState GameState, frameChannel chan []byte) {
	// fmt.Println("Game logic update. Id: ", gameState.id, "Direction:", gameState.snakeDirection)
	img := image.NewRGBA(image.Rect(0, 0, constants.FRAME_WIDTH, constants.FRAME_HEIGHT))
	// Loop through the game matrix and render each element as a colored box
	for y := 0; y < len(gameState.matrix); y++ {
		for x := 0; x < len(gameState.matrix[0]); x++ {
			rectMin := image.Point{X: x * constants.CHUNK_SIZE, Y: y * constants.CHUNK_SIZE}
			rectMax := image.Point{X: rectMin.X + constants.CHUNK_SIZE, Y: rectMin.Y + constants.CHUNK_SIZE}
			switch gameState.matrix[y][x] {
			case 1: // Snake head (Red)
				drawRectangle(img, rectMin, rectMax, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			case 2: // Snake body (White)
				drawRectangle(img, rectMin, rectMax, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			case 3: // Food (Green)
				drawRectangle(img, rectMin, rectMax, color.RGBA{R: 0, G: 255, B: 0, A: 255})
			}
		}
	}

	// Convert RGBA image to raw RGB pixel data
	rawRGBData := make([]byte, 3*constants.FRAME_WIDTH*constants.FRAME_HEIGHT)
	idx := 0
	for y := 0; y < constants.FRAME_HEIGHT; y++ {
		for x := 0; x < constants.FRAME_WIDTH; x++ {
			pixel := img.RGBAAt(x, y)
			rawRGBData[idx] = pixel.R
			rawRGBData[idx+1] = pixel.G
			rawRGBData[idx+2] = pixel.B
			idx += 3
		}
	}

	frameChannel <- rawRGBData
}
