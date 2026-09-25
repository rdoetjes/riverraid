package sprites

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SubmarineState int

const (
	SubStateSubmerged SubmarineState = iota
	SubStateSurfacing
	SubStateSurfaced
	SubStateDiving
)

// Submarine represents a naval unit that cycles between submerged and surfaced states.
type Submarine struct {
	BaseSprite
	State        SubmarineState
	StateTimer   float32
	PatrolMinX   float32
	PatrolMaxX   float32
	Direction    float32
	FireCooldown float32
	ScoreValue   int
}

// NewSubmarine creates a new submarine enemy.
func NewSubmarine(pos rl.Vector2, minX, maxX float32, speed float32) *Submarine {
	dir := float32(1.0)
	if pos.X > (minX+maxX)/2 {
		dir = -1.0
	}
	return &Submarine{
		BaseSprite: BaseSprite{
			Position:  pos,
			Velocity:  rl.Vector2{X: speed * dir, Y: 0},
			Size:      rl.Vector2{X: 60, Y: 28}, // Imposing size
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		State:        SubStateSubmerged,
		StateTimer:   2.0, // Cycle time
		PatrolMinX:   minX,
		PatrolMaxX:   maxX,
		Direction:    dir,
		FireCooldown: 0.5,
		ScoreValue:   350,
	}
}

func (s *Submarine) GetType() SpriteType {
	return TypeSubmarine
}

func (s *Submarine) Update(dt float32) {
	// Standard update if bounds not provided
	s.UpdateWithRiverBounds(dt, s.PatrolMinX, s.PatrolMaxX, false, 0, 0)
}

func (s *Submarine) UpdateWithRiverBounds(dt float32, leftBank, rightBank float32, hasIsland bool, islLeft, islRight float32) {
	if !s.Active {
		return
	}
	s.Age += dt
	s.StateTimer -= dt

	// 1. State Machine: Cycle every 2 seconds roughly
	if s.StateTimer <= 0 {
		switch s.State {
		case SubStateSubmerged:
			s.State = SubStateSurfacing
			s.StateTimer = 0.6 // Transition time
		case SubStateSurfacing:
			s.State = SubStateSurfaced
			s.StateTimer = 2.0 // Surfaced duration
		case SubStateSurfaced:
			s.State = SubStateDiving
			s.StateTimer = 0.6
		case SubStateDiving:
			s.State = SubStateSubmerged
			s.StateTimer = 2.0 // Submerged duration
		}
	}

	// 2. Movement logic (slower when submerged)
	speedMult := float32(1.0)
	if s.State == SubStateSubmerged {
		speedMult = 0.5
	}
	s.Position.X += s.Velocity.X * speedMult * dt

	// 3. Boundary handling
	halfW := s.Size.X / 2.0
	margin := float32(6.0)
	minX := leftBank + halfW + margin
	maxX := rightBank - halfW - margin

	if hasIsland {
		islandMid := (islLeft + islRight) / 2.0
		if s.Position.X < islandMid {
			maxX = islLeft - halfW - margin
		} else {
			minX = islRight + halfW - margin
		}
	}

	if maxX <= minX {
		maxX = minX + 2.0
	}

	if s.Position.X <= minX {
		s.Position.X = minX
		s.Velocity.X = float32(math.Abs(float64(s.Velocity.X)))
		s.Direction = 1.0
	} else if s.Position.X >= maxX {
		s.Position.X = maxX
		s.Velocity.X = -float32(math.Abs(float64(s.Velocity.X)))
		s.Direction = -1.0
	}

	if s.FireCooldown > 0 {
		s.FireCooldown -= dt
	}
}

func (s *Submarine) Draw(tex rl.Texture2D) {
	if !s.Active {
		return
	}

	sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
	if s.Direction < 0 {
		sourceRec.Width *= -1
	}

	destRec := rl.Rectangle{X: s.Position.X, Y: s.Position.Y, Width: s.Size.X, Height: s.Size.Y}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	switch s.State {
	case SubStateSubmerged:
		rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.Color{R: 0, G: 30, B: 60, A: 60})
	case SubStateSurfacing, SubStateDiving:
		alpha := uint8(100 + 100*math.Abs(math.Sin(float64(s.Age*8))))
		rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.Color{R: 200, G: 200, B: 200, A: alpha})
	case SubStateSurfaced:
		rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)
		// Surfaced wake
		rl.DrawCircleV(rl.Vector2{X: s.Position.X - s.Direction*30, Y: s.Position.Y}, 4+float32(math.Sin(float64(s.Age*5))), rl.Color{R: 255, G: 255, B: 255, A: 120})
	}
}
