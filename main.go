package main

import (
	"riverraid/game"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
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
