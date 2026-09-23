package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// PlayerJet represents the user-controlled vector fighter aircraft.
type PlayerJet struct {
	BaseSprite
	Fuel            float32
	MaxFuel         float32
	FuelDrainRate   float32
	BankAngle       float32 // -1.0 (hard left) to +1.0 (hard right)
	TargetBank      float32
	SpeedMultiplier float32 // 0.65 to 1.5
	TargetSpeedMul  float32
	ShootTimer      float32
	ShootInterval   float32
	InvincibleTimer float32
	Altitude        float32
	Score           int
	Lives           int
	Refueling       bool
}

// NewPlayerJet creates a new player jet instance.
func NewPlayerJet(startPos rl.Vector2) *PlayerJet {
	return &PlayerJet{
		BaseSprite: BaseSprite{
			Position:  startPos,
			Velocity:  rl.Vector2{X: 0, Y: 0},
			Size:      rl.Vector2{X: 34, Y: 46},
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		Fuel:            100.0,
		MaxFuel:         100.0,
		FuelDrainRate:   3.2,
		BankAngle:       0.0,
		TargetBank:      0.0,
		SpeedMultiplier: 1.0,
		TargetSpeedMul:  1.0,
		ShootInterval:   0.14,
		Altitude:        18.0,
		Score:           0,
		Lives:           3,
	}
}

func (p *PlayerJet) GetType() SpriteType {
	return TypePlayer
}

// ResetForRespawn resets flight position, fuel, and temporary invincibility after dying.
func (p *PlayerJet) ResetForRespawn(pos rl.Vector2) {
	p.Position = pos
	p.Velocity = rl.Vector2{X: 0, Y: 0}
	p.Fuel = 100.0
	p.Active = true
	p.BankAngle = 0.0
	p.TargetBank = 0.0
	p.SpeedMultiplier = 1.0
	p.TargetSpeedMul = 1.0
	p.InvincibleTimer = 3.0 // 3 seconds grace
	p.Refueling = false
}

func (p *PlayerJet) Update(dt float32) {
	if !p.Active {
		return
	}

	p.Age += dt

	// Smoothly interpolate banking roll
	p.BankAngle += (p.TargetBank - p.BankAngle) * (14.0 * dt)
	if p.BankAngle < -1.0 {
		p.BankAngle = -1.0
	} else if p.BankAngle > 1.0 {
		p.BankAngle = 1.0
	}

	// Smoothly interpolate speed multiplier
	p.SpeedMultiplier += (p.TargetSpeedMul - p.SpeedMultiplier) * (10.0 * dt)

	// Apply lateral velocity
	p.Position.X += p.Velocity.X * dt

	// Update shoot cooldown
	if p.ShootTimer > 0 {
		p.ShootTimer -= dt
	}

	// Invincibility countdown
	if p.InvincibleTimer > 0 {
		p.InvincibleTimer -= dt
	}

	// Drain fuel based on speed
	drain := p.FuelDrainRate * p.SpeedMultiplier * dt
	p.Fuel -= drain
	if p.Fuel < 0 {
		p.Fuel = 0
	}
}

func (p *PlayerJet) CanShoot() bool {
	return p.ShootTimer <= 0 && p.Active && p.Fuel > 0
}

func (p *PlayerJet) RecordShot() {
	p.ShootTimer = p.ShootInterval
}

func (p *PlayerJet) AddFuel(amount float32) {
	p.Fuel += amount
	if p.Fuel > p.MaxFuel {
		p.Fuel = p.MaxFuel
	}
	p.Refueling = true
}

func (p *PlayerJet) Draw(tex rl.Texture2D) {
	if !p.Active {
		return
	}

	// Invincibility flashing
	if p.InvincibleTimer > 0 {
		if int(p.Age*16)%2 == 0 {
			return
		}
	}

	// Map BankAngle (-1.0 to 1.0) to frame index (0 to 4)
	frame := int(math.Round(float64(p.BankAngle*2.0 + 2.0)))
	if frame < 0 {
		frame = 0
	}
	if frame > 4 {
		frame = 4
	}

	frameW := float32(tex.Width) / 5.0
	sourceRec := rl.Rectangle{X: float32(frame) * frameW, Y: 0, Width: frameW, Height: float32(tex.Height)}

	// Mirror the sprite based on banking direction
	if p.BankAngle < -0.1 {
		sourceRec.Width *= -1
	} else if p.BankAngle > 0.1 {
		// Normal orientation
	}

	destRec := rl.Rectangle{X: p.Position.X, Y: p.Position.Y, Width: p.Size.X * 1.5, Height: p.Size.Y * 1.5}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)

	// Refueling aura glow
	if p.Refueling {
		ui.DrawGlowCircle(p.Position, 28.0, rl.Color{R: 80, G: 255, B: 120, A: 160}, 3)
	}
}
