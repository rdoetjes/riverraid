package sprites

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SubMissile represents a slower, less accurate heat-seeking missile fired by submarines.
type SubMissile struct {
	BaseSprite
	Target       *PlayerJet
	LifeTimer    float32
	TurnSpeed    float32 // Radians per second (lower than SAM)
	CurrentAngle float32
	Speed        float32 // Lower than SAM
}

// NewSubMissile creates a new slower, less accurate missile.
func NewSubMissile(pos rl.Vector2, target *PlayerJet) *SubMissile {
	dx := target.Position.X - pos.X
	dy := target.Position.Y - pos.Y
	angle := float32(math.Atan2(float64(dy), float64(dx)))

	return &SubMissile{
		BaseSprite: BaseSprite{
			Position: pos,
			Size:     rl.Vector2{X: 8, Y: 12}, // Slightly smaller
			Active:   true,
		},
		Target:       target,
		LifeTimer:    3.5, // Lives slightly longer because it's slower
		TurnSpeed:    1.2, // Standard is 2.2
		CurrentAngle: angle,
		Speed:        200.0, // Standard is 320.0
	}
}

func (m *SubMissile) GetType() SpriteType {
	return TypeSubMissile
}

func (m *SubMissile) SetPosition(pos rl.Vector2) {
	m.Position = pos
}

func (m *SubMissile) SetActive(active bool) {
	m.Active = active
}

func (m *SubMissile) Update(dt float32) {
	if !m.Active {
		return
	}

	m.Age += dt
	m.LifeTimer -= dt

	if m.LifeTimer <= 0 {
		m.Active = false
		return
	}

	if m.Target != nil && m.Target.Active {
		tx := m.Target.Position.X - m.Position.X
		ty := m.Target.Position.Y - m.Position.Y
		targetAngle := float32(math.Atan2(float64(ty), float64(tx)))

		angleDiff := targetAngle - m.CurrentAngle
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

	m.Velocity.X = float32(math.Cos(float64(m.CurrentAngle))) * m.Speed
	m.Velocity.Y = float32(math.Sin(float64(m.CurrentAngle))) * m.Speed

	m.Position.X += m.Velocity.X * dt
	m.Position.Y += m.Velocity.Y * dt
}

func (m *SubMissile) Draw(tex rl.Texture2D) {
	if !m.Active {
		return
	}

	sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
	destRec := rl.Rectangle{X: m.Position.X, Y: m.Position.Y, Width: m.Size.X * 1.5, Height: m.Size.Y * 1.5}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	// Tint it slightly different to distinguish (yellowish/orange)
	rl.DrawTexturePro(tex, sourceRec, destRec, origin, m.CurrentAngle*rl.Rad2deg+90, rl.Color{R: 255, G: 200, B: 50, A: 255})
}
