package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SAMSite represents a Surface-to-Air Missile battery located on the riverbanks.
type SAMSite struct {
	BaseSprite
	RadarAngle     float32
	FireCooldown   float32
	DetectionRange float32
	ScoreValue     int
}

// NewSAMSite creates a new static SAM battery.
func NewSAMSite(pos rl.Vector2) *SAMSite {
	return &SAMSite{
		BaseSprite: BaseSprite{
			Position:  pos,
			Velocity:  rl.Vector2{X: 0, Y: 0},
			Size:      rl.Vector2{X: 32, Y: 32},
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		RadarAngle:     0.0,
		FireCooldown:   0.0,
		DetectionRange: 500.0,
		ScoreValue:     150,
	}
}

func (s *SAMSite) GetType() SpriteType {
	return TypeSAMSite
}

func (s *SAMSite) Update(dt float32) {
	if !s.Active {
		return
	}
	s.Age += dt
	s.RadarAngle += 4.0 * dt

	if s.FireCooldown > 0 {
		s.FireCooldown -= dt
	}
}

func (s *SAMSite) Draw() {
	if !s.Active {
		return
	}

	center := s.Position
	scl := float32(1.0)

	// 1. Concrete Octagonal Base
	baseCol := rl.Color{R: 70, G: 75, B: 80, A: 255}
	borderCol := rl.Color{R: 40, G: 45, B: 50, A: 255}

	basePts := []rl.Vector2{
		{X: center.X - 10*scl, Y: center.Y - 16*scl},
		{X: center.X + 10*scl, Y: center.Y - 16*scl},
		{X: center.X + 16*scl, Y: center.Y - 10*scl},
		{X: center.X + 16*scl, Y: center.Y + 10*scl},
		{X: center.X + 10*scl, Y: center.Y + 16*scl},
		{X: center.X - 10*scl, Y: center.Y + 16*scl},
		{X: center.X - 16*scl, Y: center.Y + 10*scl},
		{X: center.X - 16*scl, Y: center.Y - 10*scl},
	}
	ui.DrawConvexPolygonFilled(basePts, baseCol)
	ui.DrawThickPolygonOutline(basePts, 1.5, borderCol)

	// 2. Turret Platform
	rl.DrawCircleV(center, 9*scl, rl.Color{R: 50, G: 55, B: 60, A: 255})
	rl.DrawCircleLines(int32(center.X), int32(center.Y), 9*scl, borderCol)

	// 3. Radar Dish (Rotating)
	dishLen := float32(12.0)
	p1 := rl.Vector2{
		X: center.X - float32(math.Cos(float64(s.RadarAngle)))*dishLen,
		Y: center.Y - float32(math.Sin(float64(s.RadarAngle)))*dishLen,
	}
	p2 := rl.Vector2{
		X: center.X + float32(math.Cos(float64(s.RadarAngle)))*dishLen,
		Y: center.Y + float32(math.Sin(float64(s.RadarAngle)))*dishLen,
	}
	rl.DrawLineEx(p1, p2, 3.0, rl.Color{R: 200, G: 210, B: 220, A: 255})

	// Blinking status light
	if math.Sin(float64(s.Age*10.0)) > 0 {
		rl.DrawCircleV(center, 2.5, rl.Color{R: 255, G: 50, B: 50, A: 255})
		ui.DrawGlowCircle(center, 6.0, rl.Color{R: 255, G: 50, B: 50, A: 150}, 2)
	}

	// 4. Missile Rails (Static indicators)
	railCol := rl.Color{R: 30, G: 35, B: 40, A: 255}
	rl.DrawRectangle(int32(center.X-14), int32(center.Y-4), 6, 8, railCol)
	rl.DrawRectangle(int32(center.X+8), int32(center.Y-4), 6, 8, railCol)
}
