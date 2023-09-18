package latencycheck

import (
	"image"
	"image/color"
)

var (
	RED = color.RGBA{R: 255, G: 0, B: 0, A: 255}
)

const bytesPerPixel = 3 // RGB: 3 bytes per pixel

func drawRectangle(img *image.RGBA, min, max image.Point, col color.RGBA) {
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			img.Set(x, y, col)
		}
	}
}

func convertRGBAtoRGB(img *image.RGBA) []byte {
	width, height := img.Rect.Dx(), img.Rect.Dy()
	rawRGBData := make([]byte, bytesPerPixel*width*height)

	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.RGBAAt(x, y)
			rawRGBData[idx] = pixel.R
			rawRGBData[idx+1] = pixel.G
			rawRGBData[idx+2] = pixel.B
			idx += bytesPerPixel
		}
	}

	return rawRGBData
}

func hasStateChanged(prev, curr *gameState) bool {
	if prev.mousePosition.X == curr.mousePosition.X && prev.mousePosition.Y == curr.mousePosition.Y {
		return false
	}
	return true
}

func startFrameRenderer(gameStateCh chan gameState, pixelCh chan<- []byte, width, height int) {
	var lastRenderedState *gameState
	for {
		gameState := <-gameStateCh
		if lastRenderedState != nil && !hasStateChanged(lastRenderedState, &gameState) {
			continue
		}
		lastRenderedState = &gameState

		img := image.NewRGBA(image.Rect(0, 0, width, height))
		matrix := gameState.GetMatrix()

		for y := 0; y < len(matrix); y++ {
			for x := 0; x < len(matrix[0]); x++ {
				rectMin := image.Point{X: x, Y: y}
				rectMax := image.Point{X: rectMin.X + 10, Y: rectMin.Y + 10}
				if matrix[y][x] == 1 {
					drawRectangle(img, rectMin, rectMax, RED)
				}
			}
		}

		rawRGBData := convertRGBAtoRGB(img)
		pixelCh <- rawRGBData
	}
}
