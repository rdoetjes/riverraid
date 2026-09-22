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

func (d *Destroyer) Draw() {
	if !d.Active {
		return
	}

	center := d.Position
	halfW := d.Size.X / 2
	halfH := d.Size.Y / 2

	// Since the destroyer sails "down" (towards positive Y), the bow (nose) should point down.

	// 1. Water Wake / Bow Wave
	// We'll create a V-shaped wake at the bow (bottom) pointing down, and a smaller foam trail at the stern (top).
	wakeColor := rl.Color{R: 210, G: 240, B: 255, A: 100}

	// Bow Wave (V-shape at the front pointing down)
	bowY := center.Y + halfH
	rl.DrawTriangle(
		rl.Vector2{X: center.X - 12, Y: bowY - 5},
		rl.Vector2{X: center.X + 12, Y: bowY - 5},
		rl.Vector2{X: center.X, Y: bowY + 12}, // Pointing down, less tall
		rl.Color{R: 220, G: 245, B: 255, A: 60},
	)

	// Stern Wake (Small foam trail at the back pointing up)
	sternY := center.Y - halfH
	rl.DrawTriangle(
		rl.Vector2{X: center.X - 6, Y: sternY},
		rl.Vector2{X: center.X + 6, Y: sternY},
		rl.Vector2{X: center.X, Y: sternY - 10}, // Short trail
		wakeColor,
	)

	// 2. Hull (Pointed down towards the bottom of the screen - Naval Grey - FIXED CONSTANT)
	// Vertical Hull Polygon (Sharp Bow Pointing Down, Tapered Stern)
	hullPts := []rl.Vector2{
		{X: center.X, Y: center.Y + halfH},                 // Sharp Bow Point (Bottom)
		{X: center.X - halfW*0.6, Y: center.Y + halfH - 8}, // Bow curve left
		{X: center.X - halfW, Y: center.Y + halfH*0.2},     // Midship Left
		{X: center.X - halfW, Y: center.Y - halfH*0.6},     // Stern start left
		{X: center.X - halfW*0.4, Y: center.Y - halfH},     // Stern end left
		{X: center.X + halfW*0.4, Y: center.Y - halfH},     // Stern end right
		{X: center.X + halfW, Y: center.Y - halfH*0.6},     // Stern start right
		{X: center.X + halfW, Y: center.Y + halfH*0.2},     // Midship Right
		{X: center.X + halfW*0.6, Y: center.Y + halfH - 8}, // Bow curve right
	}
	ui.DrawConvexPolygonFilled(hullPts, rl.Color{R: 110, G: 115, B: 120, A: 255})
	ui.DrawThickPolygonOutline(hullPts, 1.5, rl.Color{R: 50, G: 52, B: 55, A: 255})

	// Deck Detail (Central strip)
	rl.DrawRectangleRec(rl.Rectangle{X: center.X - 3, Y: center.Y - halfH + 8, Width: 6, Height: halfH * 1.5}, rl.Color{R: 80, G: 85, B: 90, A: 255})

	// Twin Railgun Turrets (Aligned vertically along the deck)
	// Forward turret (near bow/bottom)
	rl.DrawCircleV(rl.Vector2{X: center.X, Y: center.Y + 12}, 4.0, rl.Color{R: 50, G: 52, B: 55, A: 255})
	gunPos1 := rl.Vector2{X: center.X, Y: center.Y + 12}
	rl.DrawLineEx(gunPos1, rl.Vector2{X: center.X, Y: gunPos1.Y + 6}, 2.0, rl.Color{R: 50, G: 52, B: 55, A: 255})

	// Aft turret (near stern/top)
	rl.DrawCircleV(rl.Vector2{X: center.X, Y: center.Y - 14}, 4.0, rl.Color{R: 50, G: 52, B: 55, A: 255})
	gunPos2 := rl.Vector2{X: center.X, Y: center.Y - 14}
	rl.DrawLineEx(gunPos2, rl.Vector2{X: center.X, Y: gunPos2.Y - 6}, 2.0, rl.Color{R: 50, G: 52, B: 55, A: 255})

	// Command Bridge (Central island)
	bridgeRec := rl.Rectangle{X: center.X - 6, Y: center.Y - 2, Width: 12, Height: 10}
	rl.DrawRectangleRec(bridgeRec, rl.Color{R: 45, G: 50, B: 60, A: 255})
	rl.DrawRectangleLinesEx(bridgeRec, 1.0, rl.Color{R: 50, G: 52, B: 55, A: 255})

	// Radar (On top of bridge)
	radarY := center.Y + 2
	rl.DrawLineEx(
		rl.Vector2{X: center.X - float32(math.Cos(float64(d.RadarAngle)))*6, Y: radarY - float32(math.Sin(float64(d.RadarAngle)))*6},
		rl.Vector2{X: center.X + float32(math.Cos(float64(d.RadarAngle)))*6, Y: radarY + float32(math.Sin(float64(d.RadarAngle)))*6},
		2.0,
		rl.Color{R: 255, G: 200, B: 50, A: 240},
	)
}
