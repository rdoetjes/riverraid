package sprites

import (
	"math"

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

func (f *FuelDepot) Draw(tex rl.Texture2D) {
	if !f.Active {
		return
	}

	center := f.Position
	waterBob := float32(math.Sin(float64(f.Age*3.0))) * 1.5

	sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
	destRec := rl.Rectangle{X: center.X, Y: center.Y + waterBob, Width: f.Size.X, Height: f.Size.Y}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

	rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)
}
