package sprites

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// SpriteType categorizes entities in the River Raid ecosystem.
type SpriteType int

const (
	TypePlayer SpriteType = iota
	TypeHelicopter
	TypeShip
	TypeEnemyJet
	TypeFuelDepot
	TypeBridge
	TypeSAMSite
	TypeMissile
	TypeBullet
	TypeParticle
	TypeDecoration
)

// Sprite is the base interface that every game entity implements.
type Sprite interface {
	GetType() SpriteType
	GetPosition() rl.Vector2
	SetPosition(pos rl.Vector2)
	GetBounds() rl.Rectangle
	IsActive() bool
	SetActive(active bool)
	Update(dt float32)
	Draw()
}

// BaseSprite provides standard fields for position, velocity, alive state, and bounding box.
type BaseSprite struct {
	Position  rl.Vector2
	Velocity  rl.Vector2
	Size      rl.Vector2
	Active    bool
	Health    int
	MaxHealth int
	Age       float32
}

// GetPosition returns the center position of the sprite.
func (b *BaseSprite) GetPosition() rl.Vector2 {
	return b.Position
}

// SetPosition sets the center position of the sprite.
func (b *BaseSprite) SetPosition(pos rl.Vector2) {
	b.Position = pos
}

// IsActive returns whether this sprite should be updated and rendered.
func (b *BaseSprite) IsActive() bool {
	return b.Active
}

// SetActive modifies the active state.
func (b *BaseSprite) SetActive(active bool) {
	b.Active = active
}

// GetBounds returns the AABB collision rectangle.
func (b *BaseSprite) GetBounds() rl.Rectangle {
	return rl.Rectangle{
		X:      b.Position.X - b.Size.X/2,
		Y:      b.Position.Y - b.Size.Y/2,
		Width:  b.Size.X,
		Height: b.Size.Y,
	}
}

// CheckCollisionAABB tests if two sprites collide.
func CheckCollisionAABB(a, b Sprite) bool {
	if !a.IsActive() || !b.IsActive() {
		return false
	}
	return rl.CheckCollisionRecs(a.GetBounds(), b.GetBounds())
}

// CheckCollisionCircle tests circular collision between two points.
func CheckCollisionCircle(centerA rl.Vector2, radiusA float32, centerB rl.Vector2, radiusB float32) bool {
	return rl.CheckCollisionCircles(centerA, radiusA, centerB, radiusB)
}
