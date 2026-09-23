package sprites

import (
	"math"

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

func (s *SAMSite) Draw(tex rl.Texture2D) {
	if !s.Active {
		return
	}

	// Map RadarAngle (0..2Pi) to frame index (0..7)
	frame := int(s.RadarAngle/(math.Pi*2.0)*8.0) % 8
	if frame < 0 {
		frame += 8
	}

	frameW := float32(tex.Width) / 8.0
	sourceRec := rl.Rectangle{X: float32(frame) * frameW, Y: 0, Width: frameW, Height: float32(tex.Height)}
	destRec := rl.Rectangle{X: s.Position.X, Y: s.Position.Y, Width: s.Size.X * 1.5, Height: s.Size.Y * 1.5}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)
}
