package sprites

import (
	"math"

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
			Size:      rl.Vector2{X: 65, Y: 35},
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

func (s *Ship) Draw(tex rl.Texture2D) {
	if !s.Active {
		return
	}

	sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
	if s.Direction < 0 {
		sourceRec.Width *= -1 // Flip texture if sailing left
	}

	destRec := rl.Rectangle{X: s.Position.X, Y: s.Position.Y, Width: s.Size.X, Height: s.Size.Y}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)
}
