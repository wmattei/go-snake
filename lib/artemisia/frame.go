package artemisia

type Frame struct {
	Width  int
	Height int

	image []byte
}

func NewFrame(width, height int) *Frame {
	return &Frame{
		Width:  width,
		Height: height,
		image:  make([]byte, width*height*bytesPerPixel),
	}
}

func (f *Frame) inBounds(x, y int) bool {
	return x >= 0 && x < f.Width && y >= 0 && y < f.Height
}

func (f *Frame) SetPixel(x, y int, color *Color) {
	if f.inBounds(x, y) {
		index := (y*f.Width + x) * bytesPerPixel
		f.image[index] = color[0]
		f.image[index+1] = color[1]
		f.image[index+2] = color[2]
	}
}

func (f *Frame) Bytes() []byte {
	return f.image
}

func (c *Frame) DrawCircle(cx, cy, radius int, col *Color) {

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
