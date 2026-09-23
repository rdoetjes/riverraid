package ui

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// HUD renders modern 21st-century tactical flight instruments.
type HUD struct {
	ScreenWidth  float32
	ScreenHeight float32
	AlertText    string
	AlertTimer   float32
	AlertColor   rl.Color
	Age          float32
	Font         rl.Font
}

// NewHUD creates a new tactical HUD renderer.
func NewHUD(screenWidth, screenHeight float32, font rl.Font) *HUD {
	return &HUD{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		Font:         font,
	}
}

// SetAlert triggers a temporary tactical alert message on the HUD.
func (h *HUD) SetAlert(text string, duration float32, col rl.Color) {
	h.AlertText = text
	h.AlertTimer = duration
	h.AlertColor = col
}

func (h *HUD) Update(dt float32) {
	h.Age += dt
	if h.AlertTimer > 0 {
		h.AlertTimer -= dt
	}
}

// Draw renders the cockpit HUD overlay.
func (h *HUD) Draw(fuel, maxFuel float32, score, highScore, lives, section int, speedMul float32) {
	// Top status bar panel
	topBarRec := rl.Rectangle{X: 12, Y: 10, Width: h.ScreenWidth - 20, Height: 48}
	DrawBeveledRect(topBarRec, 8.0, rl.Color{R: 10, G: 20, B: 30, A: 200}, rl.Color{R: 40, G: 120, B: 180, A: 200}, 1.5)

	// 1. Score Readout
	scoreStr := fmt.Sprintf("%06d", score)
	rl.DrawTextEx(h.Font, "SCORE", rl.Vector2{X: 28, Y: 16}, 11, 1, rl.Color{R: 120, G: 200, B: 255, A: 220})
	rl.DrawTextEx(h.Font, scoreStr, rl.Vector2{X: 28, Y: 28}, 22, 1, rl.Color{R: 255, G: 255, B: 255, A: 255})

	// 2. High Score Readout (Center)
	highStr := fmt.Sprintf("HI: %06d", highScore)
	hiSize := rl.MeasureTextEx(h.Font, highStr, 14, 1)
	rl.DrawTextEx(h.Font, highStr, rl.Vector2{X: h.ScreenWidth/2 - hiSize.X/2, Y: 26}, 14, 1, rl.Color{R: 240, G: 210, B: 60, A: 240})

	// 3. Sector / Level Readout
	sectorStr := fmt.Sprintf("ZONE %02d", section)
	secSize := rl.MeasureTextEx(h.Font, sectorStr, 20, 1)
	secX := h.ScreenWidth - 32 - secSize.X
	rl.DrawTextEx(h.Font, "SECTOR", rl.Vector2{X: secX, Y: 16}, 11, 1, rl.Color{R: 120, G: 200, B: 255, A: 220})
	rl.DrawTextEx(h.Font, sectorStr, rl.Vector2{X: secX, Y: 28}, 20, 1, rl.Color{R: 0, G: 240, B: 255, A: 255})

	// Bottom Instrument Dashboard
	bottomBarRec := rl.Rectangle{X: 12, Y: h.ScreenHeight - 64, Width: h.ScreenWidth - 24, Height: 54}
	DrawBeveledRect(bottomBarRec, 8.0, rl.Color{R: 10, G: 20, B: 30, A: 220}, rl.Color{R: 40, G: 120, B: 180, A: 200}, 1.5)

	// 4. Lives Indicator (Miniature Jet Icons)
	rl.DrawTextEx(h.Font, "RESERVES", rl.Vector2{X: 28, Y: bottomBarRec.Y + 8}, 10, 1, rl.Color{R: 120, G: 200, B: 255, A: 200})
	iconStartX := float32(32)
	iconY := bottomBarRec.Y + 34
	for i := 0; i < lives-1 && i < 6; i++ {
		ix := iconStartX + float32(i)*22
		h.drawMiniJetIcon(ix, iconY)
	}

	// 5. Tactical Fuel Gauge
	fuelPct := fuel / maxFuel
	fuelGaugeRec := rl.Rectangle{
		X:      h.ScreenWidth/2 - 120,
		Y:      bottomBarRec.Y + 22,
		Width:  240,
		Height: 18,
	}

	// Dynamic Fuel Color: Green -> Amber -> Flashing Red
	fuelCol := rl.Color{R: 40, G: 230, B: 110, A: 255}
	if fuelPct <= 0.4 && fuelPct > 0.2 {
		fuelCol = rl.Color{R: 240, G: 190, B: 30, A: 255} // Amber
	} else if fuelPct <= 0.2 {
		// Flashing Red Alert
		flash := math.Sin(float64(h.Age * 14.0))
		if flash > 0 {
			fuelCol = rl.Color{R: 255, G: 40, B: 40, A: 255}
		} else {
			fuelCol = rl.Color{R: 160, G: 20, B: 20, A: 255}
		}
	}

	DrawProgressBar(fuelGaugeRec, fuelPct, fuelCol, rl.Color{R: 25, G: 30, B: 40, A: 255}, rl.Color{R: 60, G: 140, B: 200, A: 255})

	// Fuel Text & Ticks
	rl.DrawTextEx(h.Font, "E", rl.Vector2{X: fuelGaugeRec.X - 14, Y: fuelGaugeRec.Y + 2}, 14, 1, rl.Color{R: 255, G: 100, B: 100, A: 255})
	rl.DrawTextEx(h.Font, "F", rl.Vector2{X: fuelGaugeRec.X + fuelGaugeRec.Width + 6, Y: fuelGaugeRec.Y + 2}, 14, 1, rl.Color{R: 100, G: 255, B: 150, A: 255})
	fuelLabel := fmt.Sprintf("FUEL  %3.0f%%", fuelPct*100)
	flSize := rl.MeasureTextEx(h.Font, fuelLabel, 11, 1)
	rl.DrawTextEx(h.Font, fuelLabel, rl.Vector2{X: h.ScreenWidth/2 - flSize.X/2, Y: bottomBarRec.Y + 7}, 11, 1, rl.Color{R: 200, G: 230, B: 255, A: 240})

	// 6. Throttle / Speed Indicator
	throttleW := float32(70)
	throttleX := h.ScreenWidth - 28 - throttleW
	throttleY := bottomBarRec.Y + 22
	rl.DrawTextEx(h.Font, "THROTTLE", rl.Vector2{X: throttleX, Y: bottomBarRec.Y + 8}, 10, 1, rl.Color{R: 120, G: 200, B: 255, A: 200})

	throttlePct := (speedMul - 0.65) / (1.5 - 0.65)
	if throttlePct < 0 {
		throttlePct = 0
	} else if throttlePct > 1 {
		throttlePct = 1
	}
	throttleRec := rl.Rectangle{X: throttleX, Y: throttleY, Width: throttleW, Height: 18}
	DrawProgressBar(throttleRec, throttlePct, rl.Color{R: 0, G: 200, B: 255, A: 255}, rl.Color{R: 25, G: 30, B: 40, A: 255}, rl.Color{R: 60, G: 140, B: 200, A: 255})

	// 7. Tactical Alert Popup (Center Screen)
	if h.AlertTimer > 0 {
		flicker := float32(math.Sin(float64(h.Age*18.0)))*0.2 + 0.8
		alpha := uint8(float32(h.AlertColor.A) * flicker)
		col := rl.Color{R: h.AlertColor.R, G: h.AlertColor.G, B: h.AlertColor.B, A: alpha}

		alertFontSize := float32(24)
		textSize := rl.MeasureTextEx(h.Font, h.AlertText, alertFontSize, 1)
		boxRec := rl.Rectangle{
			X:      h.ScreenWidth/2 - textSize.X/2 - 20,
			Y:      h.ScreenHeight*0.38 - 18,
			Width:  textSize.X + 40,
			Height: 46,
		}
		DrawBeveledRect(boxRec, 6.0, rl.Color{R: 15, G: 20, B: 30, A: 220}, col, 2.0)
		rl.DrawTextEx(h.Font, h.AlertText, rl.Vector2{X: h.ScreenWidth/2 - textSize.X/2, Y: h.ScreenHeight*0.38 - 8}, alertFontSize, 1, col)
	}

	// 8. Low Fuel Critical Warning (blinking if fuel < 18%)
	if fuelPct < 0.18 && fuelPct > 0 {
		if math.Sin(float64(h.Age*12.0)) > 0 {
			warnStr := "CRITICAL: LOW FUEL"
			wSize := rl.MeasureTextEx(h.Font, warnStr, 20, 1)
			rl.DrawTextEx(h.Font, warnStr, rl.Vector2{X: h.ScreenWidth/2 - wSize.X/2, Y: h.ScreenHeight - 92}, 20, 1, rl.Color{R: 255, G: 50, B: 50, A: 255})
		}
	}
}

// drawMiniJetIcon renders a mini vector jet for remaining lives.
func (h *HUD) drawMiniJetIcon(cx, cy float32) {
	col := rl.Color{R: 0, G: 220, B: 255, A: 240}
	nose := rl.Vector2{X: cx, Y: cy - 9}
	lWing := rl.Vector2{X: cx - 7, Y: cy + 4}
	rWing := rl.Vector2{X: cx + 7, Y: cy + 4}
	tail := rl.Vector2{X: cx, Y: cy + 6}

	rl.DrawTriangle(nose, lWing, tail, col)
	rl.DrawTriangle(nose, tail, rWing, col)
}
