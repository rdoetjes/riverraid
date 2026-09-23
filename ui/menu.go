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

// DrawTitle renders the main menu title screen with alternating high scores and instructions.
func (m *Menu) DrawTitle(highScore int, formattedScores []string, logoTex rl.Texture2D) {
	// Dark semi-transparent atmospheric backdrop
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 8, G: 16, B: 24, A: 210})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight * 0.28

	if logoTex.ID > 0 {
		// Draw Logo Texture - significantly larger
		logoH := float32(200.0)
		logoW := (logoH / float32(logoTex.Height)) * float32(logoTex.Width)

		// If width exceeds 80% screen width, cap it
		maxWidth := m.ScreenWidth * 0.8
		if logoW > maxWidth {
			logoW = maxWidth
			logoH = (logoW / float32(logoTex.Width)) * float32(logoTex.Height)
		}

		destRec := rl.Rectangle{
			X:      cx,
			Y:      cy - 40,
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
		titleFont := int32(52)
		tW := rl.MeasureText(titleText, titleFont)

		// Shadow/Glow text
		rl.DrawText(titleText, int32(cx)-tW/2+2, int32(cy)-28+2, titleFont, rl.Color{R: 0, G: 60, B: 100, A: 255})
		rl.DrawText(titleText, int32(cx)-tW/2, int32(cy)-28, titleFont, titleCol)
	}

	subText := "21ST CENTURY TACTICAL STRIKE"
	subFont := int32(14)
	subW := rl.MeasureText(subText, subFont)
	rl.DrawText(subText, int32(cx)-subW/2, int32(cy)+32, subFont, rl.Color{R: 240, G: 200, B: 50, A: 240})

	// Cycle logic: 5s High Scores, 5s Instructions = 10s period
	period := float32(10.0)
	phase := float32(math.Mod(float64(m.Age), float64(period)))
	showHighScores := phase < 5.0

	// Draw the content
	if showHighScores {
		m.DrawHighScores(cx, cy+70, formattedScores)
	} else {
		// Mission / Instructions Box
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

		// High Score (Static footer)
		hiStr := fmt.Sprintf("TOP PILOT RECORD: %06d", highScore)
		hiW := rl.MeasureText(hiStr, 15)
		rl.DrawText(hiStr, int32(cx)-hiW/2, int32(boxRec.Y+boxH+20), 15, rl.Color{R: 240, G: 200, B: 50, A: 255})
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
		pW := rl.MeasureText(prompt, 18)
		rl.DrawText(prompt, int32(cx)-pW/2, int32(m.ScreenHeight-50), 18, rl.Color{R: 40, G: 255, B: 140, A: 255})
	}
}

// DrawGameOver renders the post-mission debrief screen.
func (m *Menu) DrawGameOver(score, highScore, section int, reason string, formattedScores []string) {
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 15, G: 8, B: 10, A: 220})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight * 0.15

	// "MISSION FAILED"
	titleText := "MISSION FAILED"
	titleFont := int32(38)
	tW := rl.MeasureText(titleText, titleFont)
	rl.DrawText(titleText, int32(cx)-tW/2, int32(cy)-40, titleFont, rl.Color{R: 255, G: 50, B: 50, A: 255})

	// Debrief Reason
	rW := rl.MeasureText(reason, 14)
	rl.DrawText(reason, int32(cx)-rW/2, int32(cy)+2, 14, rl.Color{R: 240, G: 180, B: 160, A: 240})

	// High Score Table
	m.DrawHighScores(cx, cy+30, formattedScores)

	// Restart Prompt
	if math.Sin(float64(m.Age*5.0)) > -0.2 {
		restartPrompt := "PRESS SPACE OR ENTER TO CONTINUE"
		rstW := rl.MeasureText(restartPrompt, 18)
		rl.DrawText(restartPrompt, int32(cx)-rstW/2, int32(m.ScreenHeight-50), 18, rl.Color{R: 255, G: 240, B: 100, A: 255})
	}
}

func (m *Menu) DrawHighScores(cx, cy float32, formattedScores []string) {
	tableW := float32(360)
	tableH := float32(310)
	tableRec := rl.Rectangle{X: cx - tableW/2, Y: cy, Width: tableW, Height: tableH}
	DrawBeveledRect(tableRec, 8.0, rl.Color{R: 25, G: 18, B: 22, A: 240}, rl.Color{R: 40, G: 140, B: 200, A: 200}, 1.5)

	header := "TOP 10 ACE PILOTS"
	hW := rl.MeasureText(header, 18)
	rl.DrawText(header, int32(cx)-hW/2, int32(cy+15), 18, rl.Color{R: 0, G: 220, B: 255, A: 255})

	for i, entry := range formattedScores {
		color := rl.Color{R: 255, G: 255, B: 255, A: 255}
		if i == 0 {
			color = rl.Color{R: 255, G: 215, B: 0, A: 255} // Gold for #1
		}
		rl.DrawText(entry, int32(tableRec.X+30), int32(cy+50+float32(i*24)), 16, color)
	}
}

func (m *Menu) DrawNameEntry(score int, buffer string) {
	rl.DrawRectangle(0, 0, int32(m.ScreenWidth), int32(m.ScreenHeight), rl.Color{R: 10, G: 15, B: 30, A: 230})

	cx := m.ScreenWidth / 2
	cy := m.ScreenHeight / 2

	title := "NEW HIGH SCORE!"
	tW := rl.MeasureText(title, 32)
	rl.DrawText(title, int32(cx)-tW/2, int32(cy-100), 32, rl.Color{R: 40, G: 255, B: 140, A: 255})

	scoreStr := fmt.Sprintf("SCORE: %06d", score)
	sW := rl.MeasureText(scoreStr, 22)
	rl.DrawText(scoreStr, int32(cx)-sW/2, int32(cy-50), 22, rl.White)

	prompt := "ENTER YOUR INITIALS:"
	pW := rl.MeasureText(prompt, 18)
	rl.DrawText(prompt, int32(cx)-pW/2, int32(cy), 18, rl.Color{R: 0, G: 220, B: 255, A: 255})

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
			cW := rl.MeasureText(char, 32)
			rl.DrawText(char, int32(rect.X+boxSize/2)-cW/2, int32(rect.Y+10), 32, rl.White)
		} else if i == len(buffer) {
			// Blinking cursor
			if int(m.Age*4)%2 == 0 {
				rl.DrawRectangle(int32(rect.X+10), int32(rect.Y+boxSize-10), int32(boxSize-20), 4, rl.Color{R: 255, G: 220, B: 60, A: 255})
			}
		}
	}

	if len(buffer) == 3 {
		confirm := "PRESS ENTER TO COMMIT RECORD"
		cW := rl.MeasureText(confirm, 16)
		rl.DrawText(confirm, int32(cx)-cW/2, int32(cy+120), 16, rl.Color{R: 255, G: 220, B: 60, A: 255})
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
