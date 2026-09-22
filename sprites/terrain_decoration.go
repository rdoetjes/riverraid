package sprites

import (
	"math"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DecorationType represents different landscape props.
type DecorationType int

const (
	DecoPineTree DecorationType = iota
	DecoDeciduousTree
	DecoRock
	DecoRadarStation
	DecoBunker
	DecoHouse
	DecoBuilding
	DecoBush
)

// TerrainDecoration represents an environmental vector prop on the riverbanks.
type TerrainDecoration struct {
	BaseSprite
	DecoType DecorationType
	Scale    float32
	Rotation float32
}

// NewTerrainDecoration creates an environmental decoration.
func NewTerrainDecoration(pos rl.Vector2, decoType DecorationType, scale float32) *TerrainDecoration {
	return &TerrainDecoration{
		BaseSprite: BaseSprite{
			Position: pos,
			Active:   true,
		},
		DecoType: decoType,
		Scale:    scale,
		Rotation: 0,
	}
}

func (td *TerrainDecoration) GetType() SpriteType {
	return TypeDecoration
}

func (td *TerrainDecoration) Update(dt float32) {
	td.Age += dt
}

func (td *TerrainDecoration) Draw() {
	if !td.Active {
		return
	}

	center := td.Position
	s := td.Scale
	if s <= 0 {
		s = 1.0
	}

	switch td.DecoType {
	case DecoPineTree:
		// Shadow
		shadowOffset := rl.Vector2{X: -6 * s, Y: 8 * s}
		rl.DrawCircleV(rl.Vector2Add(center, shadowOffset), 7*s, rl.Color{R: 15, G: 35, B: 15, A: 60})

		// 3-Tier Pine Tree Canopy
		tier3 := rl.Vector2{X: center.X, Y: center.Y - 14*s}
		tier2 := rl.Vector2{X: center.X, Y: center.Y - 8*s}
		tier1 := rl.Vector2{X: center.X, Y: center.Y - 2*s}

		darkPine := rl.Color{R: 28, G: 65, B: 35, A: 255}
		midPine := rl.Color{R: 42, G: 88, B: 48, A: 255}
		lightPine := rl.Color{R: 65, G: 125, B: 68, A: 255}

		// Bottom tier
		ui.DrawConvexPolygonFilled([]rl.Vector2{
			{X: tier1.X, Y: tier1.Y - 8*s},
			{X: tier1.X + 9*s, Y: tier1.Y + 4*s},
			{X: tier1.X - 9*s, Y: tier1.Y + 4*s},
		}, darkPine)

		// Mid tier
		ui.DrawConvexPolygonFilled([]rl.Vector2{
			{X: tier2.X, Y: tier2.Y - 7*s},
			{X: tier2.X + 7*s, Y: tier2.Y + 3*s},
			{X: tier2.X - 7*s, Y: tier2.Y + 3*s},
		}, midPine)

		// Top tier
		ui.DrawConvexPolygonFilled([]rl.Vector2{
			{X: tier3.X, Y: tier3.Y - 6*s},
			{X: tier3.X + 5*s, Y: tier3.Y + 3*s},
			{X: tier3.X - 5*s, Y: tier3.Y + 3*s},
		}, lightPine)

	case DecoDeciduousTree:
		// Shadow
		shadowOffset := rl.Vector2{X: -7 * s, Y: 8 * s}
		rl.DrawCircleV(rl.Vector2Add(center, shadowOffset), 9*s, rl.Color{R: 15, G: 35, B: 15, A: 60})

		// Foliage clusters
		rl.DrawCircleV(center, 9*s, rl.Color{R: 45, G: 110, B: 40, A: 255})
		rl.DrawCircleV(rl.Vector2{X: center.X - 3*s, Y: center.Y - 3*s}, 7*s, rl.Color{R: 65, G: 145, B: 55, A: 255})
		rl.DrawCircleV(rl.Vector2{X: center.X + 2*s, Y: center.Y - 4*s}, 5*s, rl.Color{R: 95, G: 180, B: 75, A: 255})

	case DecoRock:
		// Shadow
		shadowOffset := rl.Vector2{X: -4 * s, Y: 5 * s}
		rl.DrawCircleV(rl.Vector2Add(center, shadowOffset), 6*s, rl.Color{R: 15, G: 25, B: 20, A: 50})

		rockPts := []rl.Vector2{
			{X: center.X - 6*s, Y: center.Y + 3*s},
			{X: center.X - 5*s, Y: center.Y - 4*s},
			{X: center.X + 1*s, Y: center.Y - 6*s},
			{X: center.X + 6*s, Y: center.Y - 2*s},
			{X: center.X + 5*s, Y: center.Y + 4*s},
			{X: center.X - 1*s, Y: center.Y + 5*s},
		}
		ui.DrawConvexPolygonFilled(rockPts, rl.Color{R: 110, G: 115, B: 110, A: 255})
		// Highlight facet
		ui.DrawConvexPolygonFilled([]rl.Vector2{rockPts[1], rockPts[2], center}, rl.Color{R: 150, G: 155, B: 150, A: 255})
		ui.DrawThickPolygonOutline(rockPts, 1.0, rl.Color{R: 70, G: 75, B: 70, A: 255})

	case DecoRadarStation:
		// Concrete Base
		baseRec := rl.Rectangle{X: center.X - 8*s, Y: center.Y - 6*s, Width: 16 * s, Height: 12 * s}
		ui.DrawBeveledRect(baseRec, 3*s, rl.Color{R: 75, G: 80, B: 90, A: 255}, rl.Color{R: 45, G: 50, B: 58, A: 255}, 1.2)

		// Rotating Dish
		rot := td.Age * 3.0
		dishLen := float32(7.0 * s)
		dishCenter := rl.Vector2{X: center.X, Y: center.Y - 1*s}
		p1 := rl.Vector2{X: dishCenter.X - float32(math.Cos(float64(rot)))*dishLen, Y: dishCenter.Y - float32(math.Sin(float64(rot)))*dishLen}
		p2 := rl.Vector2{X: dishCenter.X + float32(math.Cos(float64(rot)))*dishLen, Y: dishCenter.Y + float32(math.Sin(float64(rot)))*dishLen}
		rl.DrawLineEx(p1, p2, 2.5*s, rl.Color{R: 220, G: 230, B: 240, A: 255})
		rl.DrawCircleV(dishCenter, 3*s, rl.Color{R: 30, G: 35, B: 45, A: 255})
		// Blinking light
		if math.Sin(float64(td.Age*6.0)) > 0 {
			rl.DrawCircleV(dishCenter, 1.8*s, rl.Color{R: 255, G: 50, B: 50, A: 255})
		}

	case DecoBunker:
		bunkerRec := rl.Rectangle{X: center.X - 10*s, Y: center.Y - 7*s, Width: 20 * s, Height: 14 * s}
		ui.DrawBeveledRect(bunkerRec, 4*s, rl.Color{R: 65, G: 70, B: 60, A: 255}, rl.Color{R: 35, G: 40, B: 30, A: 255}, 1.5)
		// Slit window
		rl.DrawRectangle(int32(center.X-5*s), int32(center.Y-2*s), int32(10*s), int32(3*s), rl.Color{R: 15, G: 20, B: 15, A: 255})

	case DecoHouse:
		// Suburban house (top-down view)
		w, h := float32(24*s), float32(20*s)
		rect := rl.Rectangle{X: center.X - w/2, Y: center.Y - h/2, Width: w, Height: h}

		// Shadow
		ui.DrawDropShadow([]rl.Vector2{
			{X: rect.X, Y: rect.Y}, {X: rect.X + w, Y: rect.Y},
			{X: rect.X + w, Y: rect.Y + h}, {X: rect.X, Y: rect.Y + h},
		}, rl.Vector2{X: -6 * s, Y: 6 * s}, 60)

		// Roof color (Terracotta red or Dark Grey)
		roofCol := rl.Color{R: 160, G: 70, B: 50, A: 255}
		ui.DrawBeveledRect(rect, 2*s, roofCol, rl.Color{R: 80, G: 30, B: 20, A: 255}, 1.2)

		// Roof ridge line (Pitched roof look from above)
		rl.DrawLineEx(rl.Vector2{X: rect.X + 4*s, Y: center.Y}, rl.Vector2{X: rect.X + w - 4*s, Y: center.Y}, 1.5, rl.Color{R: 200, G: 110, B: 90, A: 255})

		// Chimney
		rl.DrawRectangle(int32(rect.X+w-8*s), int32(rect.Y+4*s), int32(4*s), int32(4*s), rl.Color{R: 60, G: 65, B: 70, A: 255})

	case DecoBuilding:
		// Industrial/Office building
		w, h := float32(32*s), float32(38*s)
		rect := rl.Rectangle{X: center.X - w/2, Y: center.Y - h/2, Width: w, Height: h}

		// Shadow
		ui.DrawDropShadow([]rl.Vector2{
			{X: rect.X, Y: rect.Y}, {X: rect.X + w, Y: rect.Y},
			{X: rect.X + w, Y: rect.Y + h}, {X: rect.X, Y: rect.Y + h},
		}, rl.Vector2{X: -10 * s, Y: 10 * s}, 70)

		// Flat roof (Grey concrete)
		ui.DrawBeveledRect(rect, 3*s, rl.Color{R: 110, G: 115, B: 120, A: 255}, rl.Color{R: 60, G: 65, B: 70, A: 255}, 1.5)

		// Rooftop details (HVAC units, flat segments)
		hvacCol := rl.Color{R: 85, G: 90, B: 95, A: 255}
		rl.DrawRectangle(int32(center.X-10*s), int32(center.Y-12*s), int32(8*s), int32(8*s), hvacCol)
		rl.DrawRectangle(int32(center.X+2*s), int32(center.Y+4*s), int32(10*s), int32(10*s), hvacCol)

		// Solar panels or maintenance hatches
		rl.DrawRectangle(int32(center.X-12*s), int32(center.Y+6*s), int32(10*s), int32(6*s), rl.Color{R: 40, G: 60, B: 100, A: 200})

	case DecoBush:
		// Low-profile foliage cluster
		shadowOffset := rl.Vector2{X: -3 * s, Y: 3 * s}
		rl.DrawCircleV(rl.Vector2Add(center, shadowOffset), 6*s, rl.Color{R: 10, G: 30, B: 10, A: 50})

		bushCol := rl.Color{R: 40, G: 95, B: 35, A: 255}
		rl.DrawCircleV(center, 6*s, bushCol)
		rl.DrawCircleV(rl.Vector2{X: center.X - 3*s, Y: center.Y - 2*s}, 4.5*s, rl.Color{R: 55, G: 120, B: 45, A: 255})
		rl.DrawCircleV(rl.Vector2{X: center.X + 2*s, Y: center.Y - 1*s}, 4*s, rl.Color{R: 70, G: 145, B: 55, A: 255})
	}
}
