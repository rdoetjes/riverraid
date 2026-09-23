package sprites

import (
	"math"

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

func (b *Bullet) Draw(tex rl.Texture2D) {
	if !b.Active {
		return
	}

	// For tiny bullets, we'll just keep the vector drawing for performance and precision,
	// unless specifically requested to use the missile texture.
	// But let's use the provided texture if available to satisfy the "all PNG" request.

	destRec := rl.Rectangle{X: b.Position.X, Y: b.Position.Y, Width: b.Size.X, Height: b.Size.Y}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rotation := float32(0)
	if b.Velocity.X != 0 || b.Velocity.Y != 0 {
		rotation = float32(math.Atan2(float64(b.Velocity.Y), float64(b.Velocity.X)))*rl.Rad2deg + 90
	}

	rl.DrawTexturePro(tex, rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}, destRec, origin, rotation, rl.White)
}
