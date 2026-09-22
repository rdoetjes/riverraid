package sprites

import (
	"math"

	"riverraid/ui"

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
			Velocity:  rl.Vector2{X: speed * dir, Y: 0},
			Size:      rl.Vector2{X: 54, Y: 26},
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

func (d *Destroyer) Draw() {
	if !d.Active {
		return
	}

	center := d.Position
	dir := d.Direction
	halfW := d.Size.X / 2
	halfH := d.Size.Y / 2

	// 1. Water Wake
	sternX := center.X - dir*halfW
	rl.DrawTriangle(
		rl.Vector2{X: sternX, Y: center.Y - 5},
		rl.Vector2{X: sternX, Y: center.Y + 5},
		rl.Vector2{X: sternX - dir*30, Y: center.Y},
		rl.Color{R: 200, G: 235, B: 255, A: 80},
	)

	// 2. Hull (Dark Charcoal Neutral Grey - "Threatening")
	hullCol := rl.Color{R: 70, G: 72, B: 75, A: 255}
	trimCol := rl.Color{R: 35, G: 36, B: 38, A: 255}

	hullPts := []rl.Vector2{
		{X: center.X + dir*halfW, Y: center.Y},
		{X: center.X + dir*(halfW-10), Y: center.Y - halfH},
		{X: center.X - dir*(halfW-5), Y: center.Y - halfH},
		{X: center.X - dir*halfW, Y: center.Y - (halfH - 4)},
		{X: center.X - dir*halfW, Y: center.Y + (halfH - 4)},
		{X: center.X - dir*(halfW-5), Y: center.Y + halfH},
		{X: center.X + dir*(halfW-10), Y: center.Y + halfH},
	}
	ui.DrawConvexPolygonFilled(hullPts, hullCol)
	ui.DrawThickPolygonOutline(hullPts, 1.5, trimCol)

	// Twin Railgun Turrets
	ui.DrawGlowCircle(rl.Vector2{X: center.X + dir*14, Y: center.Y}, 5.0, rl.Color{R: 40, G: 45, B: 50, A: 200}, 2)
	rl.DrawCircleV(rl.Vector2{X: center.X + dir*14, Y: center.Y}, 3.5, trimCol)
	rl.DrawCircleV(rl.Vector2{X: center.X - dir*10, Y: center.Y}, 3.5, trimCol)

	// Command Bridge
	rl.DrawRectangle(int32(center.X-6), int32(center.Y-6), 12, 12, rl.Color{R: 45, G: 50, B: 60, A: 255})
	rl.DrawRectangleLines(int32(center.X-6), int32(center.Y-6), 12, 12, trimCol)

	// Radar
	radarX := center.X - dir*2
	rl.DrawLineEx(
		rl.Vector2{X: radarX - float32(math.Cos(float64(d.RadarAngle)))*6, Y: center.Y - float32(math.Sin(float64(d.RadarAngle)))*6},
		rl.Vector2{X: radarX + float32(math.Cos(float64(d.RadarAngle)))*6, Y: center.Y + float32(math.Sin(float64(d.RadarAngle)))*6},
		2.0,
		rl.Color{R: 255, G: 200, B: 50, A: 240},
	)
}
