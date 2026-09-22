package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Missile represents a heat-seeking surface-to-air missile.
type Missile struct {
	BaseSprite
	Target       *PlayerJet
	LifeTimer    float32
	TurnSpeed    float32 // Radians per second
	CurrentAngle float32
	Speed        float32
	IsExhausting bool
}

// NewMissile creates a new heat-seeking missile targeting the player.
func NewMissile(pos rl.Vector2, target *PlayerJet) *Missile {
	// Calculate initial angle towards target
	dx := target.Position.X - pos.X
	dy := target.Position.Y - pos.Y
	angle := float32(math.Atan2(float64(dy), float64(dx)))

	return &Missile{
		BaseSprite: BaseSprite{
			Position: pos,
			Size:     rl.Vector2{X: 8, Y: 16},
			Active:   true,
		},
		Target:       target,
		LifeTimer:    2.5, // Follow for 2.5 seconds
		TurnSpeed:    2.2, // Turn rate
		CurrentAngle: angle,
		Speed:        320.0,
		IsExhausting: true,
	}
}

func (m *Missile) GetType() SpriteType {
	return TypeMissile
}

func (m *Missile) Update(dt float32) {
	if !m.Active {
		return
	}

	m.Age += dt
	m.LifeTimer -= dt

	if m.LifeTimer <= 0 {
		m.Active = false
		return
	}

	// Homing Logic
	if m.Target != nil && m.Target.Active {
		// Vector to target
		tx := m.Target.Position.X - m.Position.X
		ty := m.Target.Position.Y - m.Position.Y
		targetAngle := float32(math.Atan2(float64(ty), float64(tx)))

		// Smoothly rotate current angle towards target angle
		angleDiff := targetAngle - m.CurrentAngle
		// Normalize angle difference to -Pi..Pi
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}
		for angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		maxTurn := m.TurnSpeed * dt
		if math.Abs(float64(angleDiff)) < float64(maxTurn) {
			m.CurrentAngle = targetAngle
		} else {
			if angleDiff > 0 {
				m.CurrentAngle += maxTurn
			} else {
				m.CurrentAngle -= maxTurn
			}
		}
	}

	// Move in direction of current angle
	m.Velocity.X = float32(math.Cos(float64(m.CurrentAngle))) * m.Speed
	m.Velocity.Y = float32(math.Sin(float64(m.CurrentAngle))) * m.Speed

	m.Position.X += m.Velocity.X * dt
	m.Position.Y += m.Velocity.Y * dt
}

func (m *Missile) Draw() {
	if !m.Active {
		return
	}

	center := m.Position
	angle := m.CurrentAngle
	s := float32(1.0)

	// 1. Rocket Exhaust Flame
	flicker := float32(math.Sin(float64(m.Age*30.0)))*2.0 + 4.0
	exhaustOffset := rl.Vector2{
		X: -float32(math.Cos(float64(angle))) * (10 + flicker),
		Y: -float32(math.Sin(float64(angle))) * (10 + flicker),
	}
	ui.DrawGlowCircle(rl.Vector2Add(center, exhaustOffset), 6.0, rl.Color{R: 255, G: 160, B: 50, A: 180}, 2)
	rl.DrawCircleV(rl.Vector2Add(center, exhaustOffset), 2.5, rl.Color{R: 255, G: 240, B: 200, A: 255})

	// 2. Missile Body (Cylindrical Vector)
	tip := rl.Vector2{
		X: center.X + float32(math.Cos(float64(angle)))*8*s,
		Y: center.Y + float32(math.Sin(float64(angle)))*8*s,
	}
	tail := rl.Vector2{
		X: center.X - float32(math.Cos(float64(angle)))*8*s,
		Y: center.Y - float32(math.Sin(float64(angle)))*8*s,
	}

	// Draw main shaft
	rl.DrawLineEx(tail, tip, 2.5, rl.Color{R: 220, G: 225, B: 230, A: 255})

	// Draw fins
	finAngle := angle + math.Pi/2
	finLen := float32(5.0)
	f1 := rl.Vector2{
		X: tail.X + float32(math.Cos(float64(finAngle)))*finLen,
		Y: tail.Y + float32(math.Sin(float64(finAngle)))*finLen,
	}
	f2 := rl.Vector2{
		X: tail.X - float32(math.Cos(float64(finAngle)))*finLen,
		Y: tail.Y - float32(math.Sin(float64(finAngle)))*finLen,
	}
	rl.DrawLineEx(tail, f1, 1.5, rl.Color{R: 180, G: 185, B: 190, A: 255})
	rl.DrawLineEx(tail, f2, 1.5, rl.Color{R: 180, G: 185, B: 190, A: 255})

	// 3. Warhead Tip (Red)
	rl.DrawCircleV(tip, 1.8, rl.Color{R: 255, G: 40, B: 40, A: 255})
}
