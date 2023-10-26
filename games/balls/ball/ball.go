package ball

type Position struct {
	X int
	Y int
}

type Ball struct {
	Position   Position
	Velocity   Velocity
	Ground     int
	Elasticity float64
	Radius     int
}

type Velocity struct {
	X int
	Y int
}

func NewBall(x, y, radius, ground int) *Ball {
	return &Ball{
		Radius: radius,
		Position: Position{
			X: x,
			Y: y,
		},
		Ground:     ground,
		Elasticity: 0.8, // 80% of the energy is retained after a bounce
	}
}

const gravity = 1

func (b *Ball) Update(dt int64) {
	b.Velocity.Y += gravity
	b.Position.X += b.Velocity.X
	b.Position.Y += b.Velocity.Y
	if b.Position.Y >= b.Ground {
		b.Position.Y = b.Ground
		b.Velocity.Y = -int(float64(b.Velocity.Y) * b.Elasticity)
	}
}
