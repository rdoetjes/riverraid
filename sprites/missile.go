package sprites

import (
	"math"

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

func (m *Missile) Draw(tex rl.Texture2D) {
	if !m.Active {
		return
	}

	sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
	destRec := rl.Rectangle{X: m.Position.X, Y: m.Position.Y, Width: m.Size.X * 1.5, Height: m.Size.Y * 1.5}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rl.DrawTexturePro(tex, sourceRec, destRec, origin, m.CurrentAngle*rl.Rad2deg+90, rl.White)
}
