package sprites

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Destroyer represents an advanced naval combat ship that hunts the player.
type Destroyer struct {
	BaseSprite
	PatrolMinX   float32
	PatrolMaxX   float32
	Direction    float32
	RadarAngle   float32
	FireCooldown float32
	ScoreValue   int
}

// NewDestroyer creates a new naval hunter destroyer.
func NewDestroyer(pos rl.Vector2, minX, maxX float32, speed float32) *Destroyer {
	dir := float32(1.0)
	if pos.X > (minX+maxX)/2 {
		dir = -1.0
	}
	return &Destroyer{
		BaseSprite: BaseSprite{
			Position:  pos,
			Velocity:  rl.Vector2{X: 0, Y: speed}, // Initial vertical speed
			Size:      rl.Vector2{X: 26, Y: 54},   // Swapped: Width 26, Height 54
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		PatrolMinX:   minX,
		PatrolMaxX:   maxX,
		Direction:    dir,
		RadarAngle:   0.0,
		FireCooldown: 2.0,
		ScoreValue:   250,
	}
}

func (d *Destroyer) GetType() SpriteType {
	return TypeDestroyer
}

func (d *Destroyer) Update(dt float32) {
	// Standard update falls back to simple logic if hunting not used
	d.Age += dt
	d.RadarAngle += 8.0 * dt
	d.Position.X += d.Velocity.X * dt

	if d.FireCooldown > 0 {
		d.FireCooldown -= dt
	}
}

// UpdateWithHunterLogic moves the destroyer laterally to hunt the player and vertically to sail down the river.
func (d *Destroyer) UpdateWithHunterLogic(dt float32, playerPos rl.Vector2, leftBank, rightBank float32, hasIsland bool, islLeft, islRight float32) {
	if !d.Active {
		return
	}

	d.Age += dt
	d.RadarAngle += 8.0 * dt

	// 1. Lateral Hunting Logic
	sideSpeed := float32(math.Abs(float64(d.Velocity.X)))
	if sideSpeed < 50.0 {
		sideSpeed = 50.0
	}

	if playerPos.X < d.Position.X-5 {
		d.Velocity.X = -sideSpeed
		d.Direction = -1.0
	} else if playerPos.X > d.Position.X+5 {
		d.Velocity.X = sideSpeed
		d.Direction = 1.0
	}

	// 2. Vertical "Sailing Down" Logic
	// Sail down relative to the world at a steady pace
	d.Velocity.Y = 90.0 // Constant speed sailing "down-river" aggressively
	d.Position.Y += d.Velocity.Y * dt

	// 3. Boundary avoidance (Stay in water)
	halfW := d.Size.X / 2.0
	margin := float32(8.0)

	minX := leftBank + halfW + margin
	maxX := rightBank - halfW - margin

	if hasIsland {
		islandMid := (islLeft + islRight) / 2.0
		if d.Position.X < islandMid {
			maxX = islLeft - halfW - margin
		} else {
			minX = islRight + halfW + margin
		}
	}

	if maxX <= minX {
		maxX = minX + 2.0
	}

	d.Position.X += d.Velocity.X * dt

	// Clamp and Reverse lateral direction at shores
	if d.Position.X <= minX {
		d.Position.X = minX
		d.Velocity.X = sideSpeed
		d.Direction = 1.0
	} else if d.Position.X >= maxX {
		d.Position.X = maxX
		d.Velocity.X = -sideSpeed
		d.Direction = -1.0
	}

	if d.FireCooldown > 0 {
		d.FireCooldown -= dt
	}
}

func (d *Destroyer) Draw(tex rl.Texture2D) {
	if !d.Active {
		return
	}

	sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
	destRec := rl.Rectangle{X: d.Position.X, Y: d.Position.Y, Width: d.Size.X, Height: d.Size.Y}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)
}
