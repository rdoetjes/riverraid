package main

import (
	"math"
	"os"
	"path/filepath"
	"riverraid/game"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// ... (working directory logic)
	// Change working directory to the executable's directory to ensure assets are found
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)

		// If we are inside a macOS app bundle (Contents/MacOS), move up to Resources
		if filepath.Base(exeDir) == "MacOS" {
			parent := filepath.Dir(exeDir)
			if filepath.Base(parent) == "Contents" {
				os.Chdir(filepath.Join(parent, "Resources"))
			} else {
				os.Chdir(exeDir)
			}
		} else {
			os.Chdir(exeDir)
		}
	}

	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagVsyncHint | rl.FlagWindowResizable | rl.FlagWindowHighdpi)
	rl.InitWindow(game.DefaultScreenWidth, game.DefaultScreenHeight, "River Raid - 21st Century Strike")
	defer rl.CloseWindow()

	rl.ToggleFullscreen()
	rl.HideCursor()
	rl.SetTargetFPS(60)

	g := game.NewGame(game.DefaultScreenWidth, game.DefaultScreenHeight)
	defer g.Close()

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()
		if dt > 0.05 {
			dt = 0.05
		}

		g.HandleInput(dt)
		g.Update(dt)

		// Final Screen Presentation with Bezel logic
		rl.BeginDrawing()
		rl.ClearBackground(rl.Color{R: 10, G: 10, B: 15, A: 255}) // Outer bezel background

		g.Draw() // Renders into g.RenderTex

		// Calculate scaling to fit screen while maintaining aspect ratio
		sw := float32(rl.GetScreenWidth())
		sh := float32(rl.GetScreenHeight())
		scale := float32(math.Min(float64(sw/game.DefaultScreenWidth), float64(sh/game.DefaultScreenHeight)))

		renderW := game.DefaultScreenWidth * scale
		renderH := game.DefaultScreenHeight * scale
		offsetX := (sw - renderW) / 2
		offsetY := (sh - renderH) / 2

		// 1. Draw stylized bezel (glow/frame) around the game area
		bezelMargin := float32(4.0)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      offsetX - bezelMargin,
			Y:      offsetY - bezelMargin,
			Width:  renderW + bezelMargin*2,
			Height: renderH + bezelMargin*2,
		}, 2.0, rl.Color{R: 40, G: 50, B: 80, A: 255})

		// 2. Draw the actual game texture
		sourceRec := rl.Rectangle{X: 0, Y: 0, Width: game.DefaultScreenWidth, Height: -game.DefaultScreenHeight}
		destRec := rl.Rectangle{X: offsetX, Y: offsetY, Width: renderW, Height: renderH}
		rl.DrawTexturePro(g.RenderTex.Texture, sourceRec, destRec, rl.Vector2{X: 0, Y: 0}, 0, rl.White)

		// 3. Optional: Scanline/Vignette overlay for arcade feel
		// rl.DrawRectangleGradientV(0, 0, int32(sw), int32(sh), rl.Color{0,0,0,50}, rl.Color{0,0,0,0})

		rl.EndDrawing()
	}
}
