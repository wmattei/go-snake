package ball

import (
	"time"

	"github.com/wmattei/go-snake/lib/artemisia"
)

type Position struct {
	X int
	Y int
}

type Ball struct {
	Position   Position
	Velocity   Velocity
	Ground     int
	Elasticity float64
	Radius     float64
	Color      *artemisia.Color
	StoppedAt  *time.Time
	IsDead     bool
}

type Velocity struct {
	X float64
	Y float64
}

func NewBall(x, y int, radius float64, ground int, color *artemisia.Color) *Ball {
	return &Ball{
		Radius: radius,
		Position: Position{
			X: x,
			Y: y,
		},
		Ground:     ground,
		Elasticity: 0.8,
		Color:      color,
	}
}

const gravity = 9.8

func (b *Ball) Update(dt int64) {
	b.Velocity.Y += float64(gravity)
	b.Position.X += int(b.Velocity.X)
	b.Position.Y += int(b.Velocity.Y)
	if b.Position.Y >= b.Ground {
		b.Position.Y = b.Ground
		b.Velocity.Y = -(float64(b.Velocity.Y) * b.Elasticity)
	}
	if b.Position.Y == b.Ground && int(b.Velocity.Y) == 0 && b.StoppedAt == nil {
		now := time.Now()
		b.StoppedAt = &now
	}

	if b.StoppedAt != nil && time.Since(*b.StoppedAt) > 3*time.Second {
		b.IsDead = true
	}
}
