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
	Font         rl.Font
}

// NewMenu creates a new menu system.
func NewMenu(screenWidth, screenHeight float32, font rl.Font) *Menu {
	return &Menu{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		Font:         font,
	}
}

func (m *Menu) Update(dt float32) {
	m.Age += dt
}

// DrawTitle renders the main menu title screen with alternating high scores and instructions.
func (m *Menu) DrawTitle(highScore int, formattedScores []string, logoTex rl.Texture2D) {
	// Dark semi-transparent atmospheric backdrop
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 8, G: 16, B: 24, A: 210})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight * 0.28

	if logoTex.ID > 0 {
		// Draw Logo Texture - significantly larger
		logoH := float32(400.0)
		logoW := (logoH / float32(logoTex.Height)) * float32(logoTex.Width)

		// If width exceeds 80% screen width, cap it
		maxWidth := m.ScreenWidth * 0.8
		if logoW > maxWidth {
			logoW = maxWidth
			logoH = (logoW / float32(logoTex.Width)) * float32(logoTex.Height)
		}

		destRec := rl.Rectangle{
			X:      cx,
			Y:      cy - 80,
			Width:  logoW,
			Height: logoH,
		}
		origin := rl.Vector2{X: logoW / 2, Y: logoH / 2}

		// Pulse the logo slightly
		pulse := 1.0 + float32(math.Sin(float64(m.Age*2.5)))*0.04
		destRec.Width *= pulse
		destRec.Height *= pulse
		origin.X *= pulse
		origin.Y *= pulse

		// Draw with White tint to preserve original PNG colors and alpha transparency
		rl.DrawTexturePro(logoTex, rl.Rectangle{X: 0, Y: 0, Width: float32(logoTex.Width), Height: float32(logoTex.Height)}, destRec, origin, 0, rl.White)
	} else {
		// Fallback to Vector Title: "RIVER RAID"
		titleGlow := float32(math.Sin(float64(m.Age*3.0)))*0.2 + 0.8
		titleCol := rl.Color{R: uint8(40 * titleGlow), G: uint8(220 * titleGlow), B: 255, A: 255}

		titleText := "RIVER RAID"
		titleFont := float32(52)
		tSize := rl.MeasureTextEx(m.Font, titleText, titleFont, 1)

		// Shadow/Glow text
		rl.DrawTextEx(m.Font, titleText, rl.Vector2{X: m.ScreenWidth/2 - tSize.X/2 + 2, Y: cy - 28 + 2}, titleFont, 1, rl.Color{R: 0, G: 60, B: 100, A: 255})
		rl.DrawTextEx(m.Font, titleText, rl.Vector2{X: m.ScreenWidth/2 - tSize.X/2, Y: cy - 28}, titleFont, 1, titleCol)
	}

	subText := "21ST CENTURY TACTICAL STRIKE"
	subFont := float32(14)
	subSize := rl.MeasureTextEx(m.Font, subText, subFont, 1)
	rl.DrawTextEx(m.Font, subText, rl.Vector2{X: m.ScreenWidth/2 - subSize.X/2, Y: cy + 32}, subFont, 1, rl.Color{R: 240, G: 200, B: 50, A: 240})

	// Cycle logic: 5s High Scores, 5s Instructions = 10s period
	period := float32(10.0)
	phase := float32(math.Mod(float64(m.Age), float64(period)))
	showHighScores := phase < 5.0

	// Draw the content
	if showHighScores {
		m.DrawHighScores(cx, cy+70, formattedScores)
	} else {
		// Mission / Instructions Box
		boxW := float32(450)
		boxH := float32(220)
		boxRec := rl.Rectangle{X: cx - boxW/2, Y: cy + 70, Width: boxW, Height: boxH}
		DrawBeveledRect(boxRec, 10.0, rl.Color{R: 12, G: 24, B: 36, A: 230}, rl.Color{R: 40, G: 140, B: 200, A: 200}, 1.5)

		// Instructions
		startY := boxRec.Y + 20
		rl.DrawTextEx(m.Font, "CONTROLS & TACTICAL PROTOCOL:", rl.Vector2{X: boxRec.X + 24, Y: startY}, 13, 1, rl.Color{R: 0, G: 230, B: 255, A: 240})

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
			rl.DrawTextEx(m.Font, line, rl.Vector2{X: boxRec.X + 24, Y: startY + 24 + float32(i*18)}, 12, 1, c)
		}

		// High Score (Static footer)
		hiStr := fmt.Sprintf("TOP PILOT RECORD: %06d", highScore)
		hiSize := rl.MeasureTextEx(m.Font, hiStr, 15, 1)
		rl.DrawTextEx(m.Font, hiStr, rl.Vector2{X: m.ScreenWidth/2 - hiSize.X/2, Y: boxRec.Y + boxH + 20}, 15, 1, rl.Color{R: 240, G: 200, B: 50, A: 255})
	}

	// Apply a fade-out/fade-in overlay to the content area at transition points
	// Transition happens at phase 0.0 and 5.0
	fadeW := float32(460)
	fadeH := float32(350)
	fadeRec := rl.Rectangle{X: cx - fadeW/2, Y: cy + 60, Width: fadeW, Height: fadeH}

	// Calculate alpha for a black overlay:
	// Max opacity at 0.0, 5.0, 10.0; fully transparent in the middle (2.5, 7.5)
	// We'll use a sharper curve for the fade to make it feel like a transition
	transitionWidth := float32(0.8) // seconds for the full fade out/in
	distToTransition := float32(math.Min(float64(phase), math.Min(float64(math.Abs(float64(phase-5.0))), float64(math.Abs(float64(phase-10.0))))))

	if distToTransition < transitionWidth {
		alpha := uint8(255 * (1.0 - distToTransition/transitionWidth))
		rl.DrawRectangleRec(fadeRec, rl.Color{R: 8, G: 16, B: 24, A: alpha})
	}

	// Blinking Prompt (Always visible)
	if math.Sin(float64(m.Age*5.0)) > -0.2 {
		prompt := ">> PRESS SPACE OR ENTER TO LAUNCH <<"
		pSize := rl.MeasureTextEx(m.Font, prompt, 18, 1)
		rl.DrawTextEx(m.Font, prompt, rl.Vector2{X: m.ScreenWidth/2 - pSize.X/2, Y: m.ScreenHeight - 50}, 18, 1, rl.Color{R: 40, G: 255, B: 140, A: 255})
	}
}

// DrawGameOver renders the post-mission debrief screen.
func (m *Menu) DrawGameOver(score, highScore, section int, reason string, formattedScores []string, autoTimer float32) {
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 15, G: 8, B: 10, A: 220})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight * 0.15

	// "MISSION FAILED"
	titleText := "MISSION FAILED"
	titleFont := float32(38)
	tSize := rl.MeasureTextEx(m.Font, titleText, titleFont, 1)
	rl.DrawTextEx(m.Font, titleText, rl.Vector2{X: m.ScreenWidth/2 - tSize.X/2, Y: cy - 40}, titleFont, 1, rl.Color{R: 255, G: 50, B: 50, A: 255})

	// Debrief Reason
	rSize := rl.MeasureTextEx(m.Font, reason, 14, 1)
	rl.DrawTextEx(m.Font, reason, rl.Vector2{X: m.ScreenWidth/2 - rSize.X/2, Y: cy + 2}, 14, 1, rl.Color{R: 240, G: 180, B: 160, A: 240})

	// High Score Table
	m.DrawHighScores(cx, cy+30, formattedScores)

	// Restart Prompt
	if math.Sin(float64(m.Age*5.0)) > -0.2 {
		restartPrompt := "PRESS SPACE OR ENTER TO CONTINUE"
		rstSize := rl.MeasureTextEx(m.Font, restartPrompt, 18, 1)
		rl.DrawTextEx(m.Font, restartPrompt, rl.Vector2{X: m.ScreenWidth/2 - rstSize.X/2, Y: m.ScreenHeight - 65}, 18, 1, rl.Color{R: 255, G: 240, B: 100, A: 255})

		// Return to title countdown
		autoText := fmt.Sprintf("RETURNING TO HQ IN %d...", int(math.Ceil(float64(autoTimer))))
		autoSize := rl.MeasureTextEx(m.Font, autoText, 14, 1)
		rl.DrawTextEx(m.Font, autoText, rl.Vector2{X: m.ScreenWidth/2 - autoSize.X/2, Y: m.ScreenHeight - 35}, 14, 1, rl.Color{R: 180, G: 200, B: 220, A: 200})
	}
}

func (m *Menu) DrawHighScores(cx, cy float32, formattedScores []string) {
	tableW := float32(360)
	tableH := float32(310)
	tableRec := rl.Rectangle{X: cx - tableW/2, Y: cy, Width: tableW, Height: tableH}
	DrawBeveledRect(tableRec, 8.0, rl.Color{R: 25, G: 18, B: 22, A: 240}, rl.Color{R: 40, G: 140, B: 200, A: 200}, 1.5)

	header := "TOP 10 ACE PILOTS"
	hSize := rl.MeasureTextEx(m.Font, header, 18, 1)
	rl.DrawTextEx(m.Font, header, rl.Vector2{X: m.ScreenWidth/2 - hSize.X/2, Y: cy + 15}, 18, 1, rl.Color{R: 0, G: 220, B: 255, A: 255})

	for i, entry := range formattedScores {
		color := rl.Color{R: 255, G: 255, B: 255, A: 255}
		if i == 0 {
			color = rl.Color{R: 255, G: 215, B: 0, A: 255} // Gold for #1
		}
		rl.DrawTextEx(m.Font, entry, rl.Vector2{X: tableRec.X + 30, Y: cy + 50 + float32(i*24)}, 16, 1, color)
	}
}

func (m *Menu) DrawNameEntry(score int, buffer string) {
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 10, G: 15, B: 30, A: 230})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight / 2

	title := "NEW HIGH SCORE!"
	tSize := rl.MeasureTextEx(m.Font, title, 32, 1)
	rl.DrawTextEx(m.Font, title, rl.Vector2{X: m.ScreenWidth/2 - tSize.X/2, Y: cy - 100}, 32, 1, rl.Color{R: 40, G: 255, B: 140, A: 255})

	scoreStr := fmt.Sprintf("SCORE: %06d", score)
	sSize := rl.MeasureTextEx(m.Font, scoreStr, 22, 1)
	rl.DrawTextEx(m.Font, scoreStr, rl.Vector2{X: m.ScreenWidth/2 - sSize.X/2, Y: cy - 50}, 22, 1, rl.White)

	prompt := "ENTER YOUR INITIALS:"
	pSize := rl.MeasureTextEx(m.Font, prompt, 18, 1)
	rl.DrawTextEx(m.Font, prompt, rl.Vector2{X: m.ScreenWidth/2 - pSize.X/2, Y: cy}, 18, 1, rl.Color{R: 0, G: 220, B: 255, A: 255})

	// 3-letter boxes
	boxSize := float32(50)
	gap := float32(15)
	startX := cx - (boxSize*3+gap*2)/2

	for i := 0; i < 3; i++ {
		rect := rl.Rectangle{X: startX + float32(i)*(boxSize+gap), Y: cy + 40, Width: boxSize, Height: boxSize}
		rl.DrawRectangleRec(rect, rl.Color{R: 20, G: 40, B: 60, A: 255})
		rl.DrawRectangleLinesEx(rect, 2, rl.Color{R: 0, G: 220, B: 255, A: 255})

		if i < len(buffer) {
			char := string(buffer[i])
			cSize := rl.MeasureTextEx(m.Font, char, 32, 1)
			rl.DrawTextEx(m.Font, char, rl.Vector2{X: rect.X + boxSize/2 - cSize.X/2, Y: rect.Y + 10}, 32, 1, rl.White)
		} else if i == len(buffer) {
			// Blinking cursor
			if int(m.Age*4)%2 == 0 {
				rl.DrawRectangle(int32(rect.X+10), int32(rect.Y+boxSize-10), int32(boxSize-20), 4, rl.Color{R: 255, G: 220, B: 60, A: 255})
			}
		}
	}

	if len(buffer) == 3 {
		confirm := "PRESS ENTER TO COMMIT RECORD"
		cSize := rl.MeasureTextEx(m.Font, confirm, 16, 1)
		rl.DrawTextEx(m.Font, confirm, rl.Vector2{X: m.ScreenWidth/2 - cSize.X/2, Y: cy + 120}, 16, 1, rl.Color{R: 255, G: 220, B: 60, A: 255})
	}
}

// DrawPause renders the pause overlay.
func (m *Menu) DrawPause() {
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 5, G: 12, B: 20, A: 160})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight / 2

	pauseText := "TACTICAL PAUSE"
	pSize := rl.MeasureTextEx(m.Font, pauseText, 36, 1)
	rl.DrawTextEx(m.Font, pauseText, rl.Vector2{X: cx - pSize.X/2, Y: cy - 30}, 36, 1, rl.Color{R: 0, G: 220, B: 255, A: 255})

	sub := "PRESS P OR ESCAPE TO RESUME"
	sSize := rl.MeasureTextEx(m.Font, sub, 16, 1)
	rl.DrawTextEx(m.Font, sub, rl.Vector2{X: cx - sSize.X/2, Y: cy + 20}, 16, 1, rl.Color{R: 220, G: 240, B: 255, A: 200})
}
