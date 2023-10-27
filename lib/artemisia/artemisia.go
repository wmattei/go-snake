package artemisia

const bytesPerPixel = 3

type Color [3]byte

type Point struct {
	X int
	Y int
}

type Canvas struct {
	Width     int
	Height    int
	rawBytes  []byte
	setPixels [][]bool // A 2D slice to track which pixels are set
}

func NewCanvas(width, height int) *Canvas {
	canvas := &Canvas{
		Width:    width,
		Height:   height,
		rawBytes: make([]byte, width*height*bytesPerPixel),
		// setPixels: make([][]bool, height),
	}
	// for i := range canvas.setPixels {
	// 	canvas.setPixels[i] = make([]bool, width)
	// }
	return canvas
}

func (c *Canvas) inBounds(x, y int) bool {
	return x >= 0 && x < c.Width && y >= 0 && y < c.Height
}

func (c *Canvas) SetPixel(x, y int, color *Color) {
	if c.inBounds(x, y) { // Check boundaries and if pixel is not set already
		index := (y*c.Width + x) * bytesPerPixel
		c.rawBytes[index] = color[0]
		c.rawBytes[index+1] = color[1]
		c.rawBytes[index+2] = color[2]
		// c.setPixels[y][x] = true
	}
}

func (c *Canvas) DrawCircle(cx, cy, radius int, col *Color) {

	x := radius
	y := 0

	if radius > 0 {
		// Draw horizontal lines for the initial point
		for i := -x; i <= x; i++ {
			c.SetPixel(cx+i, cy-y, col)
			c.SetPixel(cx+i, cy+y, col)
		}
	}

	// Initial decision parameter
	p := 1 - radius
	for x > y {
		y++

		// Mid-point is inside or on the perimeter of the circle
		if p <= 0 {
			p = p + 2*y + 1
		} else { // Mid-point is outside the perimeter of the circle
			x--
			p = p + 2*y - 2*x + 1
		}

		// If the radius is zero, only a single point will be printed at center
		if x < y {
			break
		}

		// Draw horizontal lines for the generated point and its reflection
		for i := -x; i <= x; i++ {
			c.SetPixel(cx+i, cy-y, col)
			c.SetPixel(cx+i, cy+y, col)
		}
		if x != y {
			for i := -y; i <= y; i++ {
				c.SetPixel(cx+i, cy-x, col)
				c.SetPixel(cx+i, cy+x, col)
			}
		}
	}
}

func (c *Canvas) GetBytes() []byte {
	return c.rawBytes
}
