package main

import (
	"image"

	"github.com/wmattei/go-snake/constants"
)

var (
	snakeHeadColor = [3]byte{255, 0, 0}
	snakeBodyColor = [3]byte{255, 255, 255}
	foodColor      = [3]byte{0, 255, 0}
)

const bytesPerPixel = 3 // RGB: 3 bytes per pixel

func drawRectangle(rawRGBData []byte, width int, min, max image.Point, col [3]byte) {
	for y := min.Y; y < max.Y; y++ {
		for x := min.X; x < max.X; x++ {
			idx := (y*width + x) * bytesPerPixel
			rawRGBData[idx] = col[0]
			rawRGBData[idx+1] = col[1]
			rawRGBData[idx+2] = col[2]
		}
	}
}

func renderFrame(gameState *gameState, width, height int) []byte {
	// startedAt := time.Now()
	// defer logutil.LogTimeElapsed(startedAt, "renderFrame")

	rawRGBData := make([]byte, bytesPerPixel*width*height)
	matrix := gameState.GetMatrix()

	for y := 0; y < len(matrix); y++ {
		for x := 0; x < len(matrix[0]); x++ {
			rectMin := image.Point{X: x * constants.CHUNK_SIZE, Y: y * constants.CHUNK_SIZE}
			rectMax := image.Point{X: rectMin.X + constants.CHUNK_SIZE, Y: rectMin.Y + constants.CHUNK_SIZE}
			switch matrix[y][x] {
			case 1:
				drawRectangle(rawRGBData, width, rectMin, rectMax, snakeHeadColor)
			case 2:
				drawRectangle(rawRGBData, width, rectMin, rectMax, snakeBodyColor)
			case 3:
				drawRectangle(rawRGBData, width, rectMin, rectMax, foodColor)
			}
		}
	}

	return rawRGBData
}
