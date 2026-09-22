package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Ship represents a naval combat boat / patrol destroyer.
type Ship struct {
	BaseSprite
	PatrolMinX float32
	PatrolMaxX float32
	Direction  float32
	RadarAngle float32
	ScoreValue int
}

// NewShip creates a new naval combat ship.
func NewShip(pos rl.Vector2, minX, maxX float32, speed float32) *Ship {
	dir := float32(1.0)
	if pos.X > (minX+maxX)/2 {
		dir = -1.0
	}
	return &Ship{
		BaseSprite: BaseSprite{
			Position:  pos,
			Velocity:  rl.Vector2{X: speed * dir, Y: 0},
			Size:      rl.Vector2{X: 42, Y: 22},
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		PatrolMinX: minX,
		PatrolMaxX: maxX,
		Direction:  dir,
		RadarAngle: 0.0,
		ScoreValue: 30,
	}
}

func (s *Ship) GetType() SpriteType {
	return TypeShip
}

// Update advances animation and default patrol movement.
func (s *Ship) Update(dt float32) {
	if !s.Active {
		return
	}

	s.Age += dt
	s.RadarAngle += 6.0 * dt

	// Lateral patrol movement
	s.Position.X += s.Velocity.X * dt
	if s.Position.X <= s.PatrolMinX {
		s.Position.X = s.PatrolMinX
		s.Velocity.X = float32(math.Abs(float64(s.Velocity.X)))
		s.Direction = 1.0
	} else if s.Position.X >= s.PatrolMaxX {
		s.Position.X = s.PatrolMaxX
		s.Velocity.X = -float32(math.Abs(float64(s.Velocity.X)))
		s.Direction = -1.0
	}
}

// UpdateWithRiverBounds updates the ship position ensuring it strictly reverses at riverbanks and island shores.
func (s *Ship) UpdateWithRiverBounds(dt float32, leftBank, rightBank float32, hasIsland bool, islLeft, islRight float32) {
	if !s.Active {
		return
	}

	s.Age += dt
	s.RadarAngle += 6.0 * dt

	halfW := s.Size.X / 2.0
	margin := float32(4.0)

	minX := leftBank + halfW + margin
	maxX := rightBank - halfW - margin

	if hasIsland {
		islandMid := (islLeft + islRight) / 2.0
		if s.Position.X < islandMid {
			// Cruising in left channel
			maxX = islLeft - halfW - margin
			if maxX <= minX {
				maxX = minX + 2.0
			}
		} else {
			// Cruising in right channel
			minX = islRight + halfW + margin
			if minX >= maxX {
				minX = maxX - 2.0
			}
		}
	}

	s.PatrolMinX = minX
	s.PatrolMaxX = maxX

	// Lateral movement
	s.Position.X += s.Velocity.X * dt

	// Reverse when hitting left/right boundaries or island edges
	if s.Position.X <= minX {
		s.Position.X = minX
		s.Velocity.X = float32(math.Abs(float64(s.Velocity.X)))
		s.Direction = 1.0
	} else if s.Position.X >= maxX {
		s.Position.X = maxX
		s.Velocity.X = -float32(math.Abs(float64(s.Velocity.X)))
		s.Direction = -1.0
	}
}

func (s *Ship) Draw() {
	if !s.Active {
		return
	}

	center := s.Position
	dir := s.Direction // 1.0 = facing right, -1.0 = facing left
	halfW := s.Size.X / 2
	halfH := s.Size.Y / 2

	// 1. Water Wake / Hydrodynamic Foam Waves
	wakeColor := rl.Color{R: 210, G: 240, B: 255, A: 120}
	wakeLen := float32(24.0)

	// Bow wave
	bowX := center.X + dir*(halfW-4)
	rl.DrawLineEx(
		rl.Vector2{X: bowX, Y: center.Y - halfH},
		rl.Vector2{X: bowX + dir*6, Y: center.Y},
		2.0,
		wakeColor,
	)
	rl.DrawLineEx(
		rl.Vector2{X: bowX, Y: center.Y + halfH},
		rl.Vector2{X: bowX + dir*6, Y: center.Y},
		2.0,
		wakeColor,
	)

	// Stern wake trails expanding backwards
	sternX := center.X - dir*halfW
	rl.DrawTriangle(
		rl.Vector2{X: sternX, Y: center.Y - 4},
		rl.Vector2{X: sternX, Y: center.Y + 4},
		rl.Vector2{X: sternX - dir*wakeLen, Y: center.Y - 9},
		rl.Color{R: 200, G: 235, B: 255, A: 70},
	)
	rl.DrawTriangle(
		rl.Vector2{X: sternX, Y: center.Y - 4},
		rl.Vector2{X: sternX, Y: center.Y + 4},
		rl.Vector2{X: sternX - dir*wakeLen, Y: center.Y + 9},
		rl.Color{R: 200, G: 235, B: 255, A: 70},
	)

	// 2. Hull (Modern Angular Stealth Destroyer)
	hullColor := rl.Color{R: 90, G: 100, B: 115, A: 255}
	deckColor := rl.Color{R: 120, G: 130, B: 145, A: 255}
	superstructureColor := rl.Color{R: 70, G: 80, B: 95, A: 255}
	darkTrim := rl.Color{R: 45, G: 50, B: 60, A: 255}

	hullPts := []rl.Vector2{
		{X: center.X + dir*halfW, Y: center.Y},               // Bow point
		{X: center.X + dir*(halfW-8), Y: center.Y - halfH},   // Bow top
		{X: center.X - dir*(halfW-4), Y: center.Y - halfH},   // Stern top
		{X: center.X - dir*halfW, Y: center.Y - (halfH - 3)}, // Stern edge
		{X: center.X - dir*halfW, Y: center.Y + (halfH - 3)}, // Stern edge
		{X: center.X - dir*(halfW-4), Y: center.Y + halfH},   // Stern bottom
		{X: center.X + dir*(halfW-8), Y: center.Y + halfH},   // Bow bottom
	}

	ui.DrawConvexPolygonFilled(hullPts, hullColor)
	ui.DrawThickPolygonOutline(hullPts, 1.2, darkTrim)

	// Deck inlay
	deckPts := []rl.Vector2{
		{X: center.X + dir*(halfW-6), Y: center.Y},
		{X: center.X + dir*(halfW-12), Y: center.Y - (halfH - 2)},
		{X: center.X - dir*(halfW-6), Y: center.Y - (halfH - 2)},
		{X: center.X - dir*(halfW-6), Y: center.Y + (halfH - 2)},
		{X: center.X + dir*(halfW-12), Y: center.Y + (halfH - 2)},
	}
	ui.DrawConvexPolygonFilled(deckPts, deckColor)

	// 3. Superstructure (Command Bridge Island & Radar)
	bridgeRec := rl.Rectangle{
		X:      center.X - 8,
		Y:      center.Y - 5,
		Width:  16,
		Height: 10,
	}
	if dir < 0 {
		bridgeRec.X = center.X - 8
	}
	rl.DrawRectangleRec(bridgeRec, superstructureColor)
	rl.DrawRectangleLinesEx(bridgeRec, 1.0, darkTrim)

	// Forward Naval Gun Turret
	turretPos := rl.Vector2{X: center.X + dir*10, Y: center.Y}
	rl.DrawCircleV(turretPos, 3.5, darkTrim)
	gunBarrelEnd := rl.Vector2{X: turretPos.X + dir*6, Y: turretPos.Y}
	rl.DrawLineEx(turretPos, gunBarrelEnd, 2.0, darkTrim)

	// Rotating Radar Mast
	radarX := center.X - dir*2
	radarY := center.Y
	rl.DrawCircle(int32(radarX), int32(radarY), 2.5, rl.Color{R: 200, G: 215, B: 230, A: 255})
	radarLineLen := float32(4.0)
	rl.DrawLineEx(
		rl.Vector2{X: radarX - float32(math.Cos(float64(s.RadarAngle)))*radarLineLen, Y: radarY - float32(math.Sin(float64(s.RadarAngle)))*radarLineLen},
		rl.Vector2{X: radarX + float32(math.Cos(float64(s.RadarAngle)))*radarLineLen, Y: radarY + float32(math.Sin(float64(s.RadarAngle)))*radarLineLen},
		1.5,
		rl.Color{R: 255, G: 190, B: 40, A: 240},
	)

	// Stern Helipad 'H' Marking
	helipadX := center.X - dir*14
	helipadY := center.Y
	rl.DrawCircleLines(int32(helipadX), int32(helipadY), 4.0, rl.Color{R: 255, G: 255, B: 255, A: 180})
}
