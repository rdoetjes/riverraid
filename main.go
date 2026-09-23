package main

import (
	"os"
	"path/filepath"
	"riverraid/game"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Change working directory to the executable's directory to ensure assets are found
	exePath, err := os.Executable()
	if err == nil {
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

	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagVsyncHint)
	rl.InitWindow(game.DefaultScreenWidth, game.DefaultScreenHeight, "River Raid - 21st Century Strike")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	g := game.NewGame(game.DefaultScreenWidth, game.DefaultScreenHeight)
	defer g.Close()

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()
		// Clamp dt to avoid physics spiral on hitch or window dragging
		if dt > 0.05 {
			dt = 0.05
		}

		g.HandleInput(dt)
		g.Update(dt)
		g.Draw()
	}
}
