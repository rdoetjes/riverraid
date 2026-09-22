package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Helicopter represents an enemy combat gunship that patrols across the river.
type Helicopter struct {
	BaseSprite
	RotorAngle float32
	RotorSpeed float32
	PatrolMinX float32
	PatrolMaxX float32
	Direction  float32 // -1 or 1
	Altitude   float32
	ScoreValue int
}

// NewHelicopter creates a new enemy helicopter.
func NewHelicopter(pos rl.Vector2, minX, maxX float32, speed float32) *Helicopter {
	dir := float32(1.0)
	if pos.X > (minX+maxX)/2 {
		dir = -1.0
	}
	return &Helicopter{
		BaseSprite: BaseSprite{
			Position:  pos,
			Velocity:  rl.Vector2{X: speed * dir, Y: 0},
			Size:      rl.Vector2{X: 30, Y: 30},
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		RotorAngle: 0.0,
		RotorSpeed: 28.0, // Radians/sec
		PatrolMinX: minX,
		PatrolMaxX: maxX,
		Direction:  dir,
		Altitude:   14.0,
		ScoreValue: 60,
	}
}

func (h *Helicopter) GetType() SpriteType {
	return TypeHelicopter
}

// Update advances rotor animation and default patrol movement.
func (h *Helicopter) Update(dt float32) {
	if !h.Active {
		return
	}

	h.Age += dt
	h.RotorAngle += h.RotorSpeed * dt
	if h.RotorAngle > math.Pi*2 {
		h.RotorAngle -= math.Pi * 2
	}

	// Lateral patrol movement
	h.Position.X += h.Velocity.X * dt
	if h.Position.X <= h.PatrolMinX {
		h.Position.X = h.PatrolMinX
		h.Velocity.X = float32(math.Abs(float64(h.Velocity.X)))
		h.Direction = 1.0
	} else if h.Position.X >= h.PatrolMaxX {
		h.Position.X = h.PatrolMaxX
		h.Velocity.X = -float32(math.Abs(float64(h.Velocity.X)))
		h.Direction = -1.0
	}
}

// UpdateWithRiverBounds updates the helicopter position ensuring it respects riverbanks and island shores.
func (h *Helicopter) UpdateWithRiverBounds(dt float32, leftBank, rightBank float32, hasIsland bool, islLeft, islRight float32) {
	if !h.Active {
		return
	}

	h.Age += dt
	h.RotorAngle += h.RotorSpeed * dt
	if h.RotorAngle > math.Pi*2 {
		h.RotorAngle -= math.Pi * 2
	}

	halfW := h.Size.X / 2.0
	margin := float32(2.0)

	minX := leftBank + halfW + margin
	maxX := rightBank - halfW - margin

	if hasIsland {
		islandMid := (islLeft + islRight) / 2.0
		if h.Position.X < islandMid {
			// Patrol in left channel
			maxX = islLeft - halfW - margin
			if maxX <= minX {
				maxX = minX + 2.0
			}
		} else {
			// Patrol in right channel
			minX = islRight + halfW + margin
			if minX >= maxX {
				minX = maxX - 2.0
			}
		}
	}

	h.PatrolMinX = minX
	h.PatrolMaxX = maxX

	// Lateral movement
	h.Position.X += h.Velocity.X * dt

	// Reverse when hitting left/right boundaries or island edges
	if h.Position.X <= minX {
		h.Position.X = minX
		h.Velocity.X = float32(math.Abs(float64(h.Velocity.X)))
		h.Direction = 1.0
	} else if h.Position.X >= maxX {
		h.Position.X = maxX
		h.Velocity.X = -float32(math.Abs(float64(h.Velocity.X)))
		h.Direction = -1.0
	}
}

func (h *Helicopter) Draw() {
	if !h.Active {
		return
	}

	center := h.Position

	// 1. Shadow
	shadowOffset := rl.Vector2{X: -10, Y: h.Altitude * 1.5}
	shadowBody := []rl.Vector2{
		{X: center.X - 5, Y: center.Y - 12},
		{X: center.X + 5, Y: center.Y - 12},
		{X: center.X + 7, Y: center.Y + 2},
		{X: center.X + 2, Y: center.Y + 16},
		{X: center.X - 2, Y: center.Y + 16},
		{X: center.X - 7, Y: center.Y + 2},
	}
	ui.DrawDropShadow(shadowBody, shadowOffset, 70)

	// Shadow for rotor disc
	rl.DrawCircle(
		int32(center.X+shadowOffset.X),
		int32(center.Y-3+shadowOffset.Y),
		18.0,
		rl.Color{R: 10, G: 20, B: 30, A: 35},
	)

	// 2. Fuselage / Body (Military Olive / Dark Gunmetal)
	bodyColor := rl.Color{R: 70, G: 85, B: 65, A: 255}
	highlightColor := rl.Color{R: 95, G: 115, B: 85, A: 255}
	darkColor := rl.Color{R: 45, G: 55, B: 40, A: 255}

	// Tail boom
	tailEnd := rl.Vector2{X: center.X, Y: center.Y + 16}
	rl.DrawLineEx(rl.Vector2{X: center.X, Y: center.Y}, tailEnd, 3.5, darkColor)

	// Tail rotor
	tailRotorY := tailEnd.Y
	tailRotorLen := float32(math.Sin(float64(h.Age*40.0))) * 6.0
	rl.DrawLineEx(
		rl.Vector2{X: tailEnd.X - tailRotorLen, Y: tailRotorY},
		rl.Vector2{X: tailEnd.X + tailRotorLen, Y: tailRotorY},
		1.8,
		rl.Color{R: 200, G: 210, B: 220, A: 200},
	)

	// Stub wings / weapon pylons
	rl.DrawLineEx(
		rl.Vector2{X: center.X - 11, Y: center.Y - 1},
		rl.Vector2{X: center.X + 11, Y: center.Y - 1},
		2.5,
		darkColor,
	)
	// Rocket pods on wingtips
	rl.DrawRectangle(int32(center.X-13), int32(center.Y-3), 3, 5, rl.Color{R: 30, G: 35, B: 30, A: 255})
	rl.DrawRectangle(int32(center.X+10), int32(center.Y-3), 3, 5, rl.Color{R: 30, G: 35, B: 30, A: 255})

	// Main fuselage polygon
	fuselagePts := []rl.Vector2{
		{X: center.X, Y: center.Y - 14}, // Nose
		{X: center.X + 6, Y: center.Y - 8},
		{X: center.X + 6, Y: center.Y + 4},
		{X: center.X + 2, Y: center.Y + 8},
		{X: center.X - 2, Y: center.Y + 8},
		{X: center.X - 6, Y: center.Y + 4},
		{X: center.X - 6, Y: center.Y - 8},
	}
	ui.DrawConvexPolygonFilled(fuselagePts, bodyColor)
	ui.DrawThickPolygonOutline(fuselagePts, 1.2, darkColor)

	// Cockpit glass (amber/gold armored tint)
	cockpitPts := []rl.Vector2{
		{X: center.X, Y: center.Y - 13},
		{X: center.X + 3.5, Y: center.Y - 7},
		{X: center.X - 3.5, Y: center.Y - 7},
	}
	ui.DrawConvexPolygonFilled(cockpitPts, rl.Color{R: 220, G: 160, B: 40, A: 220})
	rl.DrawLineEx(cockpitPts[0], cockpitPts[2], 1.2, highlightColor)

	// 3. Spinning Main Rotor Blades (4 blades with motion blur disc)
	rotorRadius := float32(20.0)
	rotorHub := rl.Vector2{X: center.X, Y: center.Y - 3}

	// Transparent rotor blur circle
	rl.DrawCircle(int32(rotorHub.X), int32(rotorHub.Y), rotorRadius, rl.Color{R: 180, G: 200, B: 210, A: 35})

	// 4 rotor blades rotating
	for i := 0; i < 4; i++ {
		angle := h.RotorAngle + float32(i)*(math.Pi/2.0)
		bladeEnd := rl.Vector2{
			X: rotorHub.X + float32(math.Cos(float64(angle)))*rotorRadius,
			Y: rotorHub.Y + float32(math.Sin(float64(angle)))*rotorRadius,
		}
		rl.DrawLineEx(rotorHub, bladeEnd, 2.0, rl.Color{R: 220, G: 230, B: 240, A: 220})
		// Blade tip highlight
		rl.DrawCircleV(bladeEnd, 1.5, rl.Color{R: 255, G: 200, B: 50, A: 240})
	}

	// Center rotor hub cap
	rl.DrawCircleV(rotorHub, 2.5, rl.Color{R: 20, G: 25, B: 20, A: 255})
}
