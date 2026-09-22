package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// EnemyJet represents a high-speed interceptor fighter that streaks across the screen.
type EnemyJet struct {
	BaseSprite
	PatrolMinX float32
	PatrolMaxX float32
	Direction  float32 // -1 or 1
	Altitude   float32
	ScoreValue int
}

// NewEnemyJet creates a fast enemy interceptor jet.
func NewEnemyJet(pos rl.Vector2, minX, maxX float32, speed float32) *EnemyJet {
	dir := float32(1.0)
	if pos.X > (minX+maxX)/2 {
		dir = -1.0
	}
	return &EnemyJet{
		BaseSprite: BaseSprite{
			Position:  pos,
			Velocity:  rl.Vector2{X: speed * dir, Y: 0},
			Size:      rl.Vector2{X: 36, Y: 26},
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		PatrolMinX: minX,
		PatrolMaxX: maxX,
		Direction:  dir,
		Altitude:   22.0,
		ScoreValue: 100,
	}
}

func (ej *EnemyJet) GetType() SpriteType {
	return TypeEnemyJet
}

func (ej *EnemyJet) Update(dt float32) {
	if !ej.Active {
		return
	}

	ej.Age += dt
	ej.Position.X += ej.Velocity.X * dt

	// Bounce or turn around at river boundaries
	if ej.Position.X <= ej.PatrolMinX {
		ej.Position.X = ej.PatrolMinX
		ej.Velocity.X = float32(math.Abs(float64(ej.Velocity.X)))
		ej.Direction = 1.0
	} else if ej.Position.X >= ej.PatrolMaxX {
		ej.Position.X = ej.PatrolMaxX
		ej.Velocity.X = -float32(math.Abs(float64(ej.Velocity.X)))
		ej.Direction = -1.0
	}
}

func (ej *EnemyJet) Draw() {
	if !ej.Active {
		return
	}

	center := ej.Position
	dir := ej.Direction // 1.0 (flying right), -1.0 (flying left)

	// 1. Drop Shadow
	shadowOffset := rl.Vector2{X: -16, Y: ej.Altitude * 1.5}
	shadowPts := []rl.Vector2{
		{X: center.X + dir*16, Y: center.Y},
		{X: center.X - dir*14, Y: center.Y - 11},
		{X: center.X - dir*8, Y: center.Y},
		{X: center.X - dir*14, Y: center.Y + 11},
	}
	ui.DrawDropShadow(shadowPts, shadowOffset, 65)

	// 2. Engine Exhaust Flame behind jet
	flameLen := float32(12.0 + math.Sin(float64(ej.Age*50.0))*4.0)
	tailX := center.X - dir*12
	rl.DrawTriangle(
		rl.Vector2{X: tailX, Y: center.Y - 3},
		rl.Vector2{X: tailX, Y: center.Y + 3},
		rl.Vector2{X: tailX - dir*flameLen, Y: center.Y},
		rl.Color{R: 255, G: 120, B: 30, A: 240},
	)
	rl.DrawTriangle(
		rl.Vector2{X: tailX, Y: center.Y - 1.5},
		rl.Vector2{X: tailX, Y: center.Y + 1.5},
		rl.Vector2{X: tailX - dir*flameLen*0.6, Y: center.Y},
		rl.Color{R: 255, G: 240, B: 100, A: 255},
	)

	// 3. Airframe: Swept Delta Interceptor (Crimson / Stealth Charcoal)
	mainCol := rl.Color{R: 190, G: 45, B: 45, A: 255}
	darkCol := rl.Color{R: 130, G: 25, B: 25, A: 255}
	trimCol := rl.Color{R: 50, G: 20, B: 20, A: 255}

	nosePt := rl.Vector2{X: center.X + dir*17, Y: center.Y}
	leftWingPt := rl.Vector2{X: center.X - dir*12, Y: center.Y - 12}
	rightWingPt := rl.Vector2{X: center.X - dir*12, Y: center.Y + 12}
	tailInnerPt := rl.Vector2{X: center.X - dir*9, Y: center.Y}

	// Upper wing half
	ui.DrawConvexPolygonFilled([]rl.Vector2{nosePt, leftWingPt, tailInnerPt}, mainCol)
	// Lower wing half
	ui.DrawConvexPolygonFilled([]rl.Vector2{nosePt, tailInnerPt, rightWingPt}, darkCol)

	// Cockpit canopy
	canopyPts := []rl.Vector2{
		{X: center.X + dir*10, Y: center.Y},
		{X: center.X + dir*2, Y: center.Y - 2.5},
		{X: center.X - dir*4, Y: center.Y},
		{X: center.X + dir*2, Y: center.Y + 2.5},
	}
	ui.DrawConvexPolygonFilled(canopyPts, rl.Color{R: 30, G: 40, B: 55, A: 240})
	rl.DrawLineEx(canopyPts[0], canopyPts[1], 1.2, rl.Color{R: 180, G: 220, B: 255, A: 200})

	// Outline
	ui.DrawThickPolygonOutline([]rl.Vector2{nosePt, leftWingPt, tailInnerPt, rightWingPt}, 1.2, trimCol)
}
