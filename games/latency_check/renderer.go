package main

import "image/color"

var (
	RED = color.RGBA{R: 255, G: 0, B: 0, A: 255}
)

const (
	bytesPerPixel = 3 // RGB: 3 bytes per pixel
	boxSize       = 40
)

func drawRectangle(rawRGBData []byte, col color.RGBA, idx, width int) {
	for i := 0; i < boxSize; i++ {
		for j := 0; j < boxSize; j++ {
			index := idx + (i*width+j)*bytesPerPixel
			if index >= len(rawRGBData) {
				continue
			}
			rawRGBData[index] = RED.R
			rawRGBData[index+1] = RED.G
			rawRGBData[index+2] = RED.B
		}
	}
}
func renderFrame(gs *gameState, width, height int) []byte {
	matrix := gs.GetMatrix()

	rawRGBData := make([]byte, bytesPerPixel*width*height)
	idx := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if matrix[y][x] == 1 {
				drawRectangle(rawRGBData, RED, idx, width)
			}
			idx += bytesPerPixel
		}
	}

	return rawRGBData
}
