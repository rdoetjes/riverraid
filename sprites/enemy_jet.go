package sprites

import (
	"math"

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

func (ej *EnemyJet) Draw(tex rl.Texture2D) {
	if !ej.Active {
		return
	}

	sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
	if ej.Direction < 0 {
		sourceRec.Width *= -1
	}

	destRec := rl.Rectangle{X: ej.Position.X, Y: ej.Position.Y, Width: ej.Size.X * 1.5, Height: ej.Size.Y * 1.5}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)
}
