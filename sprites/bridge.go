package sprites

import (
	"math"

	"riverraid/ui"

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

func (b *Bridge) Draw() {
	if !b.Active {
		return
	}

	y := b.Position.Y
	h := b.Size.Y
	left := b.LeftBankX - 40
	right := b.RightBankX + 40
	spanW := right - left

	if !b.Destroyed {
		// 1. Water Shadow & Support Pillars in Water
		rl.DrawRectangle(int32(left), int32(y+h/2), int32(spanW), 10, rl.Color{R: 5, G: 15, B: 25, A: 80})

		// Concrete pillars
		pillarCount := int((b.RightBankX - b.LeftBankX) / 80.0)
		if pillarCount < 2 {
			pillarCount = 2
		}
		for i := 1; i <= pillarCount; i++ {
			px := b.LeftBankX + float32(i)*(b.RightBankX-b.LeftBankX)/float32(pillarCount+1)
			pillarRec := rl.Rectangle{X: px - 6, Y: y - h/2 - 4, Width: 12, Height: h + 12}
			rl.DrawRectangleRec(pillarRec, rl.Color{R: 85, G: 90, B: 95, A: 255})
			rl.DrawRectangleLinesEx(pillarRec, 1.0, rl.Color{R: 45, G: 50, B: 55, A: 255})
		}

		// 2. Asphalt Road Deck
		roadRec := rl.Rectangle{
			X:      left,
			Y:      y - h/2,
			Width:  spanW,
			Height: h,
		}
		rl.DrawRectangleRec(roadRec, rl.Color{R: 45, G: 48, B: 52, A: 255})

		// Steel Guardrails (Top and Bottom)
		railCol := rl.Color{R: 190, G: 200, B: 215, A: 255}
		rl.DrawLineEx(rl.Vector2{X: left, Y: y - h/2}, rl.Vector2{X: right, Y: y - h/2}, 2.5, railCol)
		rl.DrawLineEx(rl.Vector2{X: left, Y: y + h/2}, rl.Vector2{X: right, Y: y + h/2}, 2.5, railCol)

		// Yellow Dashed Center Line
		dashLen := float32(14.0)
		gapLen := float32(10.0)
		for dx := left; dx < right; dx += dashLen + gapLen {
			endX := dx + dashLen
			if endX > right {
				endX = right
			}
			rl.DrawLineEx(rl.Vector2{X: dx, Y: y}, rl.Vector2{X: endX, Y: y}, 2.0, rl.Color{R: 240, G: 200, B: 40, A: 240})
		}

		// Steel Truss Diagonal Girders
		trussCol := rl.Color{R: 140, G: 155, B: 175, A: 160}
		for tx := left; tx < right; tx += 26.0 {
			rl.DrawLineEx(rl.Vector2{X: tx, Y: y - h/2}, rl.Vector2{X: tx + 13, Y: y + h/2}, 1.2, trussCol)
			rl.DrawLineEx(rl.Vector2{X: tx + 13, Y: y - h/2}, rl.Vector2{X: tx + 26, Y: y + h/2}, 1.2, trussCol)
		}

		// 3. Military Convoy Truck crossing
		truckW := float32(18.0)
		truckH := float32(9.0)
		truckPos := rl.Vector2{X: b.VehicleX, Y: y - 5}
		if b.VehicleDir < 0 {
			truckPos.Y = y + 5
		}
		truckRec := rl.Rectangle{X: truckPos.X - truckW/2, Y: truckPos.Y - truckH/2, Width: truckW, Height: truckH}
		rl.DrawRectangleRec(truckRec, rl.Color{R: 75, G: 90, B: 65, A: 255})
		rl.DrawRectangleLinesEx(truckRec, 1.0, rl.Color{R: 35, G: 45, B: 30, A: 255})
		// Truck cabin
		cabinX := truckRec.X + truckW - 5
		if b.VehicleDir < 0 {
			cabinX = truckRec.X
		}
		rl.DrawRectangle(int32(cabinX), int32(truckRec.Y+1), 5, int32(truckH-2), rl.Color{R: 160, G: 210, B: 240, A: 220})

	} else {
		// Destroyed Bridge: Collapsed Broken Spans with Gap in the Middle
		gapWidth := float32(80.0 + b.CollapseProgress*40.0)
		midX := (b.LeftBankX + b.RightBankX) / 2
		leftSpanEnd := midX - gapWidth/2
		rightSpanStart := midX + gapWidth/2

		// Left broken section angling into water
		leftSpanW := leftSpanEnd - left
		if leftSpanW > 0 {
			p1 := rl.Vector2{X: left, Y: y - h/2}
			p2 := rl.Vector2{X: leftSpanEnd, Y: y - h/2 + b.CollapseProgress*12}
			p3 := rl.Vector2{X: leftSpanEnd, Y: y + h/2 + b.CollapseProgress*18}
			p4 := rl.Vector2{X: left, Y: y + h/2}
			ui.DrawConvexPolygonFilled([]rl.Vector2{p1, p2, p3, p4}, rl.Color{R: 50, G: 52, B: 55, A: 240})
			rl.DrawLineEx(p1, p2, 2.0, rl.Color{R: 160, G: 170, B: 180, A: 255})
			rl.DrawLineEx(p4, p3, 2.0, rl.Color{R: 160, G: 170, B: 180, A: 255})
		}

		// Right broken section angling into water
		rightSpanW := right - rightSpanStart
		if rightSpanW > 0 {
			p1 := rl.Vector2{X: rightSpanStart, Y: y - h/2 + b.CollapseProgress*12}
			p2 := rl.Vector2{X: right, Y: y - h/2}
			p3 := rl.Vector2{X: right, Y: y + h/2}
			p4 := rl.Vector2{X: rightSpanStart, Y: y + h/2 + b.CollapseProgress*18}
			ui.DrawConvexPolygonFilled([]rl.Vector2{p1, p2, p3, p4}, rl.Color{R: 50, G: 52, B: 55, A: 240})
			rl.DrawLineEx(p1, p2, 2.0, rl.Color{R: 160, G: 170, B: 180, A: 255})
			rl.DrawLineEx(p4, p3, 2.0, rl.Color{R: 160, G: 170, B: 180, A: 255})
		}

		// Smoldering embers on broken edges
		flicker := float32(math.Sin(float64(b.Age*20.0)))*0.5 + 0.5
		rl.DrawCircleV(rl.Vector2{X: leftSpanEnd, Y: y}, 4.0, rl.Color{R: 255, G: uint8(100 + flicker*100), B: 20, A: 240})
		rl.DrawCircleV(rl.Vector2{X: rightSpanStart, Y: y}, 4.0, rl.Color{R: 255, G: uint8(100 + flicker*100), B: 20, A: 240})
	}
}
