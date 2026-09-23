package sprites

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Bridge represents a river crossing structure that serves as a section checkpoint.
type Bridge struct {
	BaseSprite
	LeftBankX        float32
	RightBankX       float32
	SectionIndex     int
	Destroyed        bool
	CollapseProgress float32
	ScoreValue       int
	VehicleX         float32
	VehicleDir       float32
	FireCooldown     float32
}

// NewBridge creates a new river bridge spanning between left and right banks.
func NewBridge(posY float32, leftX float32, rightX float32, sectionIndex int) *Bridge {
	width := rightX - leftX + 80.0 // Overlap well into embankments
	centerX := (leftX + rightX) / 2
	return &Bridge{
		BaseSprite: BaseSprite{
			Position:  rl.Vector2{X: centerX, Y: posY},
			Velocity:  rl.Vector2{X: 0, Y: 0},
			Size:      rl.Vector2{X: width, Y: 32},
			Active:    true,
			Health:    1,
			MaxHealth: 1,
		},
		LeftBankX:    leftX,
		RightBankX:   rightX,
		SectionIndex: sectionIndex,
		Destroyed:    false,
		ScoreValue:   500,
		VehicleX:     leftX + 20,
		VehicleDir:   1.0,
	}
}

func (b *Bridge) GetType() SpriteType {
	return TypeBridge
}

func (b *Bridge) Update(dt float32) {
	if !b.Active {
		return
	}
	b.Age += dt

	if b.Destroyed {
		if b.CollapseProgress < 1.0 {
			b.CollapseProgress += dt * 1.5
			if b.CollapseProgress > 1.0 {
				b.CollapseProgress = 1.0
			}
		}
		return
	}

	// Move vehicle across bridge
	b.VehicleX += b.VehicleDir * (45.0 * dt)
	if b.VehicleX > b.RightBankX-20 {
		b.VehicleX = b.RightBankX - 20
		b.VehicleDir = -1.0
	} else if b.VehicleX < b.LeftBankX+20 {
		b.VehicleX = b.LeftBankX + 20
		b.VehicleDir = 1.0
	}
}

// Destroy marks the bridge as collapsed.
func (b *Bridge) Destroy() {
	b.Destroyed = true
}

func (b *Bridge) Draw(tex rl.Texture2D) {
	if !b.Active {
		return
	}

	if !b.Destroyed {
		sourceRec := rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}
		destRec := rl.Rectangle{X: b.Position.X, Y: b.Position.Y, Width: b.Size.X, Height: b.Size.Y}
		origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}

		rl.DrawTexturePro(tex, sourceRec, destRec, origin, 0, rl.White)

		// Simple vehicle box
		rl.DrawRectangle(int32(b.VehicleX-9), int32(b.Position.Y-4), 18, 9, rl.Color{R: 75, G: 90, B: 65, A: 255})
	} else {
		// Keep vector collapse for drama or just hide it
		rl.DrawRectangle(int32(b.Position.X-b.Size.X/2), int32(b.Position.Y-2), int32(b.Size.X), 4, rl.DarkGray)
	}
}
