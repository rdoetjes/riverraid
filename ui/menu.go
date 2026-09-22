package ui

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Menu handles rendering of the start screen, game over screen, and pause menu.
type Menu struct {
	ScreenWidth  float32
	ScreenHeight float32
	Age          float32
}

// NewMenu creates a new menu system.
func NewMenu(screenWidth, screenHeight float32) *Menu {
	return &Menu{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

func (m *Menu) Update(dt float32) {
	m.Age += dt
}

// DrawTitle renders the main menu title screen.
func (m *Menu) DrawTitle(highScore int) {
	// Dark semi-transparent atmospheric backdrop
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 8, G: 16, B: 24, A: 210})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight * 0.32

	// Vector Title: "RIVER RAID"
	titleGlow := float32(math.Sin(float64(m.Age*3.0)))*0.2 + 0.8
	titleCol := rl.Color{R: uint8(40 * titleGlow), G: uint8(220 * titleGlow), B: 255, A: 255}

	titleText := "RIVER RAID"
	titleFont := int32(52)
	tW := rl.MeasureText(titleText, titleFont)

	// Shadow/Glow text
	rl.DrawText(titleText, int32(cx)-tW/2+2, int32(cy)-28+2, titleFont, rl.Color{R: 0, G: 60, B: 100, A: 255})
	rl.DrawText(titleText, int32(cx)-tW/2, int32(cy)-28, titleFont, titleCol)

	subText := "21ST CENTURY TACTICAL STRIKE"
	subFont := int32(14)
	subW := rl.MeasureText(subText, subFont)
	rl.DrawText(subText, int32(cx)-subW/2, int32(cy)+32, subFont, rl.Color{R: 240, G: 200, B: 50, A: 240})

	// Mission Box
	boxW := float32(440)
	boxH := float32(230)
	boxRec := rl.Rectangle{X: cx - boxW/2, Y: cy + 70, Width: boxW, Height: boxH}
	DrawBeveledRect(boxRec, 10.0, rl.Color{R: 12, G: 24, B: 36, A: 230}, rl.Color{R: 40, G: 140, B: 200, A: 200}, 1.5)

	// Instructions
	startY := int32(boxRec.Y + 20)
	rl.DrawText("CONTROLS & TACTICAL PROTOCOL:", int32(boxRec.X+24), startY, 13, rl.Color{R: 0, G: 230, B: 255, A: 240})

	lines := []string{
		"[ W / UP ]     - Accelerate / Burner",
		"[ S / DOWN ]   - Decelerate / Cruise",
		"[ A / D ]      - Bank / Steer River",
		"[ SPACE / J ]  - Fire Autocannon",
		"[ P ]          - Tactical Pause",
		"[ M ]          - Toggle Sound FX",
		"",
		"* Fly over FUEL depots to replenish tanks!",
		"* Destroy BRIDGES to clear river sectors!",
	}

	for i, line := range lines {
		c := rl.Color{R: 210, G: 225, B: 240, A: 230}
		if len(line) > 0 && line[0] == '*' {
			c = rl.Color{R: 255, G: 215, B: 80, A: 240}
		}
		rl.DrawText(line, int32(boxRec.X+24), startY+24+int32(i*18), 12, c)
	}

	// High Score
	hiStr := fmt.Sprintf("TOP PILOT RECORD: %06d", highScore)
	hiW := rl.MeasureText(hiStr, 15)
	rl.DrawText(hiStr, int32(cx)-hiW/2, int32(boxRec.Y+boxH+20), 15, rl.Color{R: 240, G: 200, B: 50, A: 255})

	// Blinking Prompt
	if math.Sin(float64(m.Age*5.0)) > -0.2 {
		prompt := ">> PRESS SPACE OR ENTER TO LAUNCH <<"
		pW := rl.MeasureText(prompt, 18)
		rl.DrawText(prompt, int32(cx)-pW/2, int32(boxRec.Y+boxH+54), 18, rl.Color{R: 40, G: 255, B: 140, A: 255})
	}
}

// DrawGameOver renders the post-mission debrief screen.
func (m *Menu) DrawGameOver(score, highScore, section int, reason string) {
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 15, G: 8, B: 10, A: 220})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight * 0.35

	// "MISSION FAILED"
	titleText := "MISSION FAILED"
	titleFont := int32(44)
	tW := rl.MeasureText(titleText, titleFont)
	rl.DrawText(titleText, int32(cx)-tW/2, int32(cy)-40, titleFont, rl.Color{R: 255, G: 50, B: 50, A: 255})

	// Debrief Reason
	rW := rl.MeasureText(reason, 16)
	rl.DrawText(reason, int32(cx)-rW/2, int32(cy)+12, 16, rl.Color{R: 240, G: 180, B: 160, A: 240})

	// Score debrief card
	cardW := float32(340)
	cardH := float32(160)
	cardRec := rl.Rectangle{X: cx - cardW/2, Y: cy + 45, Width: cardW, Height: cardH}
	DrawBeveledRect(cardRec, 8.0, rl.Color{R: 25, G: 18, B: 22, A: 240}, rl.Color{R: 220, G: 60, B: 60, A: 200}, 1.5)

	sY := int32(cardRec.Y + 22)
	scoreLine := fmt.Sprintf("FINAL SCORE:        %06d", score)
	rl.DrawText(scoreLine, int32(cardRec.X+24), sY, 15, rl.Color{R: 255, G: 255, B: 255, A: 255})

	hiLine := fmt.Sprintf("RECORD HIGH:        %06d", highScore)
	rl.DrawText(hiLine, int32(cardRec.X+24), sY+32, 15, rl.Color{R: 240, G: 200, B: 50, A: 255})

	secLine := fmt.Sprintf("SECTORS REACHED:    ZONE %02d", section)
	rl.DrawText(secLine, int32(cardRec.X+24), sY+64, 15, rl.Color{R: 0, G: 220, B: 255, A: 255})

	// Restart Prompt
	if math.Sin(float64(m.Age*5.0)) > -0.2 {
		restartPrompt := "PRESS SPACE OR ENTER TO RETRY"
		rstW := rl.MeasureText(restartPrompt, 18)
		rl.DrawText(restartPrompt, int32(cx)-rstW/2, int32(cardRec.Y+cardH+35), 18, rl.Color{R: 255, G: 240, B: 100, A: 255})
	}
}

// DrawPause renders the pause overlay.
func (m *Menu) DrawPause() {
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 5, G: 12, B: 20, A: 160})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight / 2

	pauseText := "TACTICAL PAUSE"
	pW := rl.MeasureText(pauseText, 36)
	rl.DrawText(pauseText, int32(cx)-pW/2, int32(cy)-30, 36, rl.Color{R: 0, G: 220, B: 255, A: 255})

	sub := "PRESS P OR ESCAPE TO RESUME"
	sW := rl.MeasureText(sub, 16)
	rl.DrawText(sub, int32(cx)-sW/2, int32(cy)+20, 16, rl.Color{R: 220, G: 240, B: 255, A: 200})
}
