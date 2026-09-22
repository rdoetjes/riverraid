package sprites

import (
	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Bullet represents a projectile fired by the player or enemy units.
type Bullet struct {
	BaseSprite
	IsPlayerBullet bool
	Damage         int
	TrailLength    float32
	Color          rl.Color
	CoreColor      rl.Color
}

// NewBullet creates a new bullet instance.
func NewBullet(pos rl.Vector2, vel rl.Vector2, isPlayer bool) *Bullet {
	b := &Bullet{
		BaseSprite: BaseSprite{
			Position: pos,
			Velocity: vel,
			Size:     rl.Vector2{X: 6, Y: 18},
			Active:   true,
			Health:   1,
		},
		IsPlayerBullet: isPlayer,
		Damage:         1,
		TrailLength:    16,
	}

	if isPlayer {
		b.Color = rl.Color{R: 0, G: 220, B: 255, A: 255}       // Cyan laser
		b.CoreColor = rl.Color{R: 240, G: 255, B: 255, A: 255} // White core
	} else {
		b.Color = rl.Color{R: 255, G: 60, B: 40, A: 255}       // Red plasma
		b.CoreColor = rl.Color{R: 255, G: 230, B: 200, A: 255} // Yellow core
	}

	return b
}

func (b *Bullet) GetType() SpriteType {
	return TypeBullet
}

func (b *Bullet) Update(dt float32) {
	if !b.Active {
		return
	}
	b.Position.X += b.Velocity.X * dt
	b.Position.Y += b.Velocity.Y * dt
	b.Age += dt

	// Lifetime safety cutoff
	if b.Age > 3.0 {
		b.Active = false
	}
}

func (b *Bullet) Draw() {
	if !b.Active {
		return
	}

	// Bullet vector head and tail
	tail := rl.Vector2{
		X: b.Position.X,
		Y: b.Position.Y + (b.Size.Y * 0.7),
	}
	if b.Velocity.Y > 0 {
		tail.Y = b.Position.Y - (b.Size.Y * 0.7)
	}

	// Glow trail
	ui.DrawGlowLine(tail, b.Position, 4.0, b.Color)

	// Intense bright core
	rl.DrawLineEx(tail, b.Position, 2.0, b.CoreColor)
	rl.DrawCircleV(b.Position, 3.0, b.CoreColor)
}
