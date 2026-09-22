package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// FuelDepot represents an offshore fuel station floating on the river.
type FuelDepot struct {
	BaseSprite
	FuelAmount float32
	ScoreValue int
	LightTimer float32
}

// NewFuelDepot creates a new fuel station entity.
func NewFuelDepot(pos rl.Vector2) *FuelDepot {
	return &FuelDepot{
		BaseSprite: BaseSprite{
			Position:  pos,
			Velocity:  rl.Vector2{X: 0, Y: 0},
			Size:      rl.Vector2{X: 28, Y: 48},
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		FuelAmount: 35.0, // Refueling rate
		ScoreValue: 80,
	}
}

func (f *FuelDepot) GetType() SpriteType {
	return TypeFuelDepot
}

func (f *FuelDepot) Update(dt float32) {
	if !f.Active {
		return
	}
	f.Age += dt
	f.LightTimer += dt
}

func (f *FuelDepot) Draw() {
	if !f.Active {
		return
	}

	center := f.Position
	halfW := f.Size.X / 2
	halfH := f.Size.Y / 2

	// 1. Water Ripple / Anchored Buoyancy Rings
	waterBob := float32(math.Sin(float64(f.Age*3.0))) * 1.5
	ringAlpha := uint8(60 + math.Sin(float64(f.Age*4.0))*25.0)
	rl.DrawCircleLines(int32(center.X), int32(center.Y+waterBob), halfH*1.1, rl.Color{R: 180, G: 230, B: 255, A: ringAlpha})

	// 2. Pontoon Platform Base (Industrial safety yellow & dark steel)
	bounds := rl.Rectangle{
		X:      center.X - halfW,
		Y:      center.Y - halfH + waterBob,
		Width:  f.Size.X,
		Height: f.Size.Y,
	}
	ui.DrawBeveledRect(bounds, 4.0, rl.Color{R: 45, G: 50, B: 58, A: 255}, rl.Color{R: 240, G: 190, B: 30, A: 255}, 1.5)

	// Diagonal hazard stripes on upper and lower borders
	rl.DrawRectangle(int32(bounds.X+2), int32(bounds.Y+2), int32(bounds.Width-4), 3, rl.Color{R: 240, G: 190, B: 30, A: 255})
	rl.DrawRectangle(int32(bounds.X+2), int32(bounds.Y+bounds.Height-5), int32(bounds.Width-4), 3, rl.Color{R: 240, G: 190, B: 30, A: 255})

	// 3. Cylindrical Fuel Storage Tanks (Twin white/metallic tanks with gauge level)
	tankW := float32(8.0)
	tankH := bounds.Height - 16.0
	tankY := bounds.Y + 8.0

	// Left tank
	leftTankRec := rl.Rectangle{X: bounds.X + 4, Y: tankY, Width: tankW, Height: tankH}
	rl.DrawRectangleRec(leftTankRec, rl.Color{R: 225, G: 230, B: 240, A: 255})
	rl.DrawRectangleLinesEx(leftTankRec, 1.0, rl.Color{R: 120, G: 130, B: 145, A: 255})
	// Fuel level fill indicator
	rl.DrawRectangle(int32(leftTankRec.X+1), int32(leftTankRec.Y+tankH*0.3), int32(tankW-2), int32(tankH*0.7-1), rl.Color{R: 40, G: 200, B: 80, A: 180})

	// Right tank
	rightTankRec := rl.Rectangle{X: bounds.X + bounds.Width - 4 - tankW, Y: tankY, Width: tankW, Height: tankH}
	rl.DrawRectangleRec(rightTankRec, rl.Color{R: 225, G: 230, B: 240, A: 255})
	rl.DrawRectangleLinesEx(rightTankRec, 1.0, rl.Color{R: 120, G: 130, B: 145, A: 255})
	rl.DrawRectangle(int32(rightTankRec.X+1), int32(rightTankRec.Y+tankH*0.3), int32(tankW-2), int32(tankH*0.7-1), rl.Color{R: 40, G: 200, B: 80, A: 180})

	// Connecting transfer pipes
	pipeY := center.Y + waterBob
	rl.DrawLineEx(rl.Vector2{X: bounds.X + 4 + tankW, Y: pipeY}, rl.Vector2{X: rightTankRec.X, Y: pipeY}, 2.0, rl.Color{R: 240, G: 190, B: 30, A: 255})

	// 4. Central Sign: "FUEL" text with illuminated green/cyan glow
	fuelBox := rl.Rectangle{
		X:      center.X - 9,
		Y:      center.Y - 10 + waterBob,
		Width:  18,
		Height: 20,
	}
	rl.DrawRectangleRec(fuelBox, rl.Color{R: 20, G: 35, B: 30, A: 240})
	rl.DrawRectangleLinesEx(fuelBox, 1.0, rl.Color{R: 40, G: 220, B: 120, A: 255})

	// Vector "F", "U", "E", "L" or glowing icon
	fontSize := int32(10)
	textWidth := rl.MeasureText("FUEL", fontSize)
	rl.DrawText("FUEL", int32(center.X)-textWidth/2, int32(center.Y-5+waterBob), fontSize, rl.Color{R: 60, G: 255, B: 140, A: 255})

	// 5. Flashing Beacon Lights (Top and Bottom)
	beaconGlow := float32(math.Sin(float64(f.LightTimer * 8.0)))
	if beaconGlow > 0 {
		topBeacon := rl.Vector2{X: center.X, Y: bounds.Y + 3}
		botBeacon := rl.Vector2{X: center.X, Y: bounds.Y + bounds.Height - 3}
		rl.DrawCircleV(topBeacon, 2.0, rl.Color{R: 255, G: 60, B: 40, A: 255})
		rl.DrawCircleV(botBeacon, 2.0, rl.Color{R: 255, G: 60, B: 40, A: 255})
		ui.DrawGlowCircle(topBeacon, 6.0, rl.Color{R: 255, G: 60, B: 40, A: 160}, 2)
		ui.DrawGlowCircle(botBeacon, 6.0, rl.Color{R: 255, G: 60, B: 40, A: 160}, 2)
	}
}
