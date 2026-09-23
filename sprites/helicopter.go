package sprites

import (
	"math"

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
			Size:      rl.Vector2{X: 38, Y: 38},
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

func (h *Helicopter) Draw(tex rl.Texture2D) {
	if !h.Active {
		return
	}

	frame := int(h.Age*12.0) % 4
	frameW := float32(tex.Width) / 4.0
	sourceRec := rl.Rectangle{X: float32(frame) * frameW, Y: 0, Width: frameW, Height: float32(tex.Height)}
	destRec := rl.Rectangle{X: h.Position.X, Y: h.Position.Y, Width: h.Size.X * 1.5, Height: h.Size.Y * 1.5}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	// Draw Shadow (offset and darkened)
	shadowOffset := rl.Vector2{X: 10, Y: 10}
	shadowRec := destRec
	shadowRec.X += shadowOffset.X
	shadowRec.Y += shadowOffset.Y
	rl.DrawTexturePro(tex, sourceRec, shadowRec, origin, 0, rl.Color{R: 0, G: 0, B: 0, A: 110})

	// Draw Sprite
	rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)
}
