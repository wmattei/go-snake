package renderer

import (
	"image"
	"image/color"

	"github.com/wmattei/go-snake/constants"
	"github.com/wmattei/go-snake/game"
)

var (
	snakeHeadColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	snakeBodyColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	foodColor      = color.RGBA{R: 0, G: 255, B: 0, A: 255}
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

func StartFrameRenderer(gameStateCh chan *game.GameState, pixelCh chan []byte) {
	for {
		gameState := <-gameStateCh
		if gameState == nil {
			break
		}

		img := image.NewRGBA(image.Rect(0, 0, constants.FRAME_WIDTH, constants.FRAME_HEIGHT))
		matrix := gameState.GetMatrix()

		for y := 0; y < len(matrix); y++ {
			for x := 0; x < len(matrix[0]); x++ {
				rectMin := image.Point{X: x * constants.CHUNK_SIZE, Y: y * constants.CHUNK_SIZE}
				rectMax := image.Point{X: rectMin.X + constants.CHUNK_SIZE, Y: rectMin.Y + constants.CHUNK_SIZE}
				switch matrix[y][x] {
				case 1:
					drawRectangle(img, rectMin, rectMax, snakeHeadColor)
				case 2:
					drawRectangle(img, rectMin, rectMax, snakeBodyColor)
				case 3:
					drawRectangle(img, rectMin, rectMax, foodColor)
				}
			}
		}

		rawRGBData := convertRGBAtoRGB(img)
		pixelCh <- rawRGBData
	}
}
