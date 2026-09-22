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

func (p *PlayerJet) Draw() {
	if !p.Active {
		return
	}

	// Invincibility flashing
	if p.InvincibleTimer > 0 {
		if int(p.Age*16)%2 == 0 {
			return
		}
	}

	center := p.Position
	bank := p.BankAngle // -1.0 to 1.0
	widthFactor := float32(1.0 - math.Abs(float64(bank))*0.28)

	// 1. Shadow beneath the jet (offset diagonally)
	shadowOffset := rl.Vector2{X: -14.0 + bank*6.0, Y: p.Altitude * 1.6}
	shadowPts := []rl.Vector2{
		{X: center.X, Y: center.Y - 22},
		{X: center.X + 16*widthFactor, Y: center.Y + 8},
		{X: center.X + 8*widthFactor, Y: center.Y + 18},
		{X: center.X - 8*widthFactor, Y: center.Y + 18},
		{X: center.X - 16*widthFactor, Y: center.Y + 8},
	}
	ui.DrawDropShadow(shadowPts, shadowOffset, 75)

	// 2. Engine Exhaust Flames / Afterburner
	flameLength := float32(10.0 + p.SpeedMultiplier*12.0 + float32(math.Sin(float64(p.Age*45.0)))*3.0)
	leftNozzle := rl.Vector2{X: center.X - 5.0*widthFactor, Y: center.Y + 19.0}
	rightNozzle := rl.Vector2{X: center.X + 5.0*widthFactor, Y: center.Y + 19.0}

	// Outer fiery plume
	rl.DrawTriangle(
		rl.Vector2{X: leftNozzle.X - 3.0, Y: leftNozzle.Y},
		rl.Vector2{X: leftNozzle.X + 3.0, Y: leftNozzle.Y},
		rl.Vector2{X: leftNozzle.X, Y: leftNozzle.Y + flameLength},
		rl.Color{R: 255, G: 140, B: 30, A: 220},
	)
	rl.DrawTriangle(
		rl.Vector2{X: rightNozzle.X - 3.0, Y: rightNozzle.Y},
		rl.Vector2{X: rightNozzle.X + 3.0, Y: rightNozzle.Y},
		rl.Vector2{X: rightNozzle.X, Y: rightNozzle.Y + flameLength},
		rl.Color{R: 255, G: 140, B: 30, A: 220},
	)

	// Cyan high-energy inner core
	rl.DrawTriangle(
		rl.Vector2{X: leftNozzle.X - 1.5, Y: leftNozzle.Y},
		rl.Vector2{X: leftNozzle.X + 1.5, Y: leftNozzle.Y},
		rl.Vector2{X: leftNozzle.X, Y: leftNozzle.Y + flameLength*0.6},
		rl.Color{R: 120, G: 240, B: 255, A: 255},
	)
	rl.DrawTriangle(
		rl.Vector2{X: rightNozzle.X - 1.5, Y: rightNozzle.Y},
		rl.Vector2{X: rightNozzle.X + 1.5, Y: rightNozzle.Y},
		rl.Vector2{X: rightNozzle.X, Y: rightNozzle.Y + flameLength*0.6},
		rl.Color{R: 120, G: 240, B: 255, A: 255},
	)

	// 3. Main Jet Airframe (21st Century Stealth Fighter: F-22 raptor style)
	// Body Color shades for 3D lighting
	mainBodyCol := rl.Color{R: 195, G: 205, B: 220, A: 255}
	leftWingCol := rl.Color{R: 165, G: 175, B: 195, A: 255}
	rightWingCol := rl.Color{R: 215, G: 225, B: 235, A: 255}
	fuselageCol := rl.Color{R: 145, G: 155, B: 175, A: 255}
	borderCol := rl.Color{R: 50, G: 65, B: 85, A: 255}

	// Adjust lighting based on bank
	if bank > 0 {
		// Rolling right: left side shaded, right side bright
		leftWingCol.R -= uint8(bank * 35)
		leftWingCol.G -= uint8(bank * 35)
		leftWingCol.B -= uint8(bank * 35)
	} else if bank < 0 {
		bMag := -bank
		rightWingCol.R -= uint8(bMag * 35)
		rightWingCol.G -= uint8(bMag * 35)
		rightWingCol.B -= uint8(bMag * 35)
	}

	nosePt := rl.Vector2{X: center.X, Y: center.Y - 24}
	leftWingTip := rl.Vector2{X: center.X - 18.0*widthFactor - bank*4.0, Y: center.Y + 8}
	rightWingTip := rl.Vector2{X: center.X + 18.0*widthFactor - bank*4.0, Y: center.Y + 8}
	leftTailWing := rl.Vector2{X: center.X - 12.0*widthFactor, Y: center.Y + 19}
	rightTailWing := rl.Vector2{X: center.X + 12.0*widthFactor, Y: center.Y + 19}
	tailCenter := rl.Vector2{X: center.X, Y: center.Y + 17}

	// Left wing polygon
	ui.DrawConvexPolygonFilled([]rl.Vector2{nosePt, leftWingTip, leftTailWing, tailCenter}, leftWingCol)
	// Right wing polygon
	ui.DrawConvexPolygonFilled([]rl.Vector2{nosePt, rightWingTip, rightTailWing, tailCenter}, rightWingCol)

	// Central stealth spine / fuselage
	spineLeft := rl.Vector2{X: center.X - 4.5*widthFactor, Y: center.Y + 18}
	spineRight := rl.Vector2{X: center.X + 4.5*widthFactor, Y: center.Y + 18}
	ui.DrawConvexPolygonFilled([]rl.Vector2{nosePt, spineRight, spineLeft}, mainBodyCol)

	// Twin vertical stabilizers (rudders)
	rudderL1 := rl.Vector2{X: center.X - 7.0*widthFactor, Y: center.Y + 10}
	rudderL2 := rl.Vector2{X: center.X - 9.0*widthFactor, Y: center.Y + 21}
	rudderR1 := rl.Vector2{X: center.X + 7.0*widthFactor, Y: center.Y + 10}
	rudderR2 := rl.Vector2{X: center.X + 9.0*widthFactor, Y: center.Y + 21}
	rl.DrawLineEx(rudderL1, rudderL2, 2.5, fuselageCol)
	rl.DrawLineEx(rudderR1, rudderR2, 2.5, fuselageCol)

	// Cockpit glass canopy (glowing tinted vector bubble)
	cockpitTop := rl.Vector2{X: center.X, Y: center.Y - 14}
	cockpitBottom := rl.Vector2{X: center.X, Y: center.Y - 2}
	cockpitLeft := rl.Vector2{X: center.X - 3.0*widthFactor, Y: center.Y - 7}
	cockpitRight := rl.Vector2{X: center.X + 3.0*widthFactor, Y: center.Y - 7}

	ui.DrawConvexPolygonFilled([]rl.Vector2{cockpitTop, cockpitRight, cockpitBottom, cockpitLeft}, rl.Color{R: 20, G: 120, B: 180, A: 230})
	// Glass specular reflection glint
	rl.DrawLineEx(
		rl.Vector2{X: cockpitTop.X + 0.5, Y: cockpitTop.Y + 2},
		rl.Vector2{X: cockpitLeft.X + 1.2, Y: cockpitLeft.Y + 2},
		1.5,
		rl.Color{R: 200, G: 245, B: 255, A: 220},
	)

	// Stealth airframe panel outlines
	ui.DrawThickPolygonOutline([]rl.Vector2{nosePt, rightWingTip, rightTailWing, spineRight, leftNozzle, tailCenter, rightNozzle, spineLeft, leftTailWing, leftWingTip}, 1.2, borderCol)

	// Refueling aura glow
	if p.Refueling {
		ui.DrawGlowCircle(p.Position, 28.0, rl.Color{R: 80, G: 255, B: 120, A: 160}, 3)
	}
}
