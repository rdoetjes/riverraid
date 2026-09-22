package game

import (
	"fmt"
	"math"
	"math/rand"

	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Draw renders the procedural river world, entities, particle effects, and tactical HUD.
func (g *Game) Draw() {
	rl.BeginDrawing()

	// Clear background to deep ocean blue
	rl.ClearBackground(rl.Color{R: 15, G: 65, B: 115, A: 255})

	// Calculate Camera Shake Offset
	var shakeX, shakeY float32
	if g.ScreenShake > 0.05 {
		shakeX = (rand.Float32() - 0.5) * g.ScreenShake * 2.0
		shakeY = (rand.Float32() - 0.5) * g.ScreenShake * 2.0
	}

	// 1. Render River Mesh & Embankments
	g.drawRiverAndTerrain(shakeX, shakeY)

	// 2. Render Terrain Scenery Decorations
	g.drawDecorations(shakeX, shakeY)

	// 3. Render Bridges (Underneath floating craft & aircraft)
	g.drawBridges(shakeX, shakeY)

	// 4. Render Surface & Aerial Entities (Ships, Depots, Choppers, Enemy Jets)
	g.drawEnemies(shakeX, shakeY) // Draw enemies (including destroyer)

	// 5. Render Player Jet
	g.drawPlayer(shakeX, shakeY)

	// 6. Render Bullets & Missiles
	g.drawBullets(shakeX, shakeY)
	g.drawMissiles(shakeX, shakeY)

	// 7. Render Particle FX
	g.drawParticles(shakeX, shakeY)

	// 8. Render HUD & Menus (Screen Space)
	playerSection := g.World.GetSectionAt(g.Player.Position.Y)

	switch g.State {
	case StatePlaying:
		g.HUD.Draw(
			g.Player.Fuel,
			g.Player.MaxFuel,
			g.Player.Score,
			g.HighScore,
			g.Player.Lives,
			playerSection,
			g.Player.SpeedMultiplier,
		)

	case StateDying:
		g.HUD.Draw(
			0,
			g.Player.MaxFuel,
			g.Player.Score,
			g.HighScore,
			g.Player.Lives,
			playerSection,
			1.0,
		)
		g.drawDeathOverlay()

	case StatePaused:
		g.HUD.Draw(
			g.Player.Fuel,
			g.Player.MaxFuel,
			g.Player.Score,
			g.HighScore,
			g.Player.Lives,
			playerSection,
			g.Player.SpeedMultiplier,
		)
		g.Menu.DrawPause()

	case StateTitle:
		g.Menu.DrawTitle(g.HighScore)

	case StateGameOver:
		g.Menu.DrawGameOver(g.Player.Score, g.HighScore, playerSection, g.GameOverReason)
	}

	rl.EndDrawing()
}

// drawRiverAndTerrain draws water ripples, shoreline gradients, embankments, and islands.
func (g *Game) drawRiverAndTerrain(sx, sy float32) {
	// Water Wave Flow Shimmer Lines
	waveTime := float64(g.TotalPlayTime)
	for i := 0; i < 24; i++ {
		waveY := float32(math.Mod(float64(i)*40.0+waveTime*50.0, float64(g.ScreenHeight)))
		waveX := float32(math.Sin(waveTime*2.0+float64(i)*1.5))*30.0 + g.ScreenWidth/2
		waveLen := float32(40.0 + math.Sin(float64(i)*2.2)*20.0)

		rl.DrawLineEx(
			rl.Vector2{X: waveX - waveLen/2 + sx, Y: waveY + sy},
			rl.Vector2{X: waveX + waveLen/2 + sx, Y: waveY + sy},
			1.5,
			rl.Color{R: 50, G: 140, B: 200, A: 60},
		)
	}

	slices := g.World.ActiveSlices
	if len(slices) < 2 {
		return
	}

	landDark := rl.Color{R: 28, G: 68, B: 32, A: 255}
	landMid := rl.Color{R: 45, G: 98, B: 48, A: 255}
	sandCoast := rl.Color{R: 215, G: 185, B: 115, A: 255}
	waterShallow := rl.Color{R: 35, G: 110, B: 165, A: 120}

	// Iterate through river slices to construct embankment and island quad strips
	for i := 0; i < len(slices)-1; i++ {
		s0 := slices[i]
		s1 := slices[i+1]

		screenY0 := s0.WorldY - g.CameraY + sy
		screenY1 := s1.WorldY - g.CameraY + sy

		// Skip slices outside vertical viewport
		if screenY1 > g.ScreenHeight+40 || screenY0 < -40 {
			continue
		}

		// --- Left Embankment ---
		l0 := s0.LeftBankX + sx
		l1 := s1.LeftBankX + sx

		// Solid land interior
		pL1 := rl.Vector2{X: 0, Y: screenY0}
		pL2 := rl.Vector2{X: l0, Y: screenY0}
		pL3 := rl.Vector2{X: l1, Y: screenY1}
		pL4 := rl.Vector2{X: 0, Y: screenY1}
		ui.DrawConvexPolygonFilled([]rl.Vector2{pL1, pL2, pL3, pL4}, landDark)

		// Grass highlight band
		ui.DrawConvexPolygonFilled([]rl.Vector2{
			{X: l0 - 18, Y: screenY0},
			{X: l0, Y: screenY0},
			{X: l1, Y: screenY1},
			{X: l1 - 18, Y: screenY1},
		}, landMid)

		// Sandy beach border line
		rl.DrawLineEx(rl.Vector2{X: l0, Y: screenY0}, rl.Vector2{X: l1, Y: screenY1}, 3.0, sandCoast)
		// Shallow water rim
		rl.DrawLineEx(rl.Vector2{X: l0 + 3.0, Y: screenY0}, rl.Vector2{X: l1 + 3.0, Y: screenY1}, 4.0, waterShallow)

		// --- Right Embankment ---
		r0 := s0.RightBankX + sx
		r1 := s1.RightBankX + sx

		// Solid land interior
		pR1 := rl.Vector2{X: r0, Y: screenY0}
		pR2 := rl.Vector2{X: g.ScreenWidth, Y: screenY0}
		pR3 := rl.Vector2{X: g.ScreenWidth, Y: screenY1}
		pR4 := rl.Vector2{X: r1, Y: screenY1}
		ui.DrawConvexPolygonFilled([]rl.Vector2{pR1, pR2, pR3, pR4}, landDark)

		// Grass highlight band
		ui.DrawConvexPolygonFilled([]rl.Vector2{
			{X: r0, Y: screenY0},
			{X: r0 + 18, Y: screenY0},
			{X: r1 + 18, Y: screenY1},
			{X: r1, Y: screenY1},
		}, landMid)

		// Sandy beach border line
		rl.DrawLineEx(rl.Vector2{X: r0, Y: screenY0}, rl.Vector2{X: r1, Y: screenY1}, 3.0, sandCoast)
		// Shallow water rim
		rl.DrawLineEx(rl.Vector2{X: r0 - 3.0, Y: screenY0}, rl.Vector2{X: r1 - 3.0, Y: screenY1}, 4.0, waterShallow)

		// --- Central Island (if present in either slice) ---
		if s0.HasIsland || s1.HasIsland {
			// Interpolate island bounds smoothly
			il0, ir0 := s0.IslandLeftX+sx, s0.IslandRightX+sx
			if !s0.HasIsland {
				mid := (s0.RiverCenter) + sx
				il0, ir0 = mid, mid
			}
			il1, ir1 := s1.IslandLeftX+sx, s1.IslandRightX+sx
			if !s1.HasIsland {
				mid := (s1.RiverCenter) + sx
				il1, ir1 = mid, mid
			}

			if ir0 > il0 || ir1 > il1 {
				islP1 := rl.Vector2{X: il0, Y: screenY0}
				islP2 := rl.Vector2{X: ir0, Y: screenY0}
				islP3 := rl.Vector2{X: ir1, Y: screenY1}
				islP4 := rl.Vector2{X: il1, Y: screenY1}

				ui.DrawConvexPolygonFilled([]rl.Vector2{islP1, islP2, islP3, islP4}, landMid)
				// Left beach
				rl.DrawLineEx(islP1, islP4, 2.5, sandCoast)
				// Right beach
				rl.DrawLineEx(islP2, islP3, 2.5, sandCoast)
			}
		}
	}
}

// drawDecorations renders embankment and island scenery.
func (g *Game) drawDecorations(sx, sy float32) {
	for _, deco := range g.World.Decorations {
		origPos := deco.Position
		screenY := origPos.Y - g.CameraY + sy
		if screenY < -30 || screenY > g.ScreenHeight+30 {
			continue
		}
		deco.Position.X = origPos.X + sx
		deco.Position.Y = screenY
		deco.Draw()
		deco.Position = origPos // Restore world pos
	}
}

// drawBridges renders river bridges.
func (g *Game) drawBridges(sx, sy float32) {
	for _, bridge := range g.World.Bridges {
		origPos := bridge.Position
		screenY := origPos.Y - g.CameraY + sy
		if screenY < -60 || screenY > g.ScreenHeight+60 {
			continue
		}
		bridge.Position.X = origPos.X + sx
		bridge.Position.Y = screenY
		bridge.Draw()
		bridge.Position = origPos
	}
}

// drawEnemies renders ships, fuel depots, choppers, and jets.
func (g *Game) drawEnemies(sx, sy float32) {
	for _, enemy := range g.World.Enemies {
		if !enemy.IsActive() {
			continue
		}
		origPos := enemy.GetPosition()
		screenY := origPos.Y - g.CameraY + sy
		if screenY < -60 || screenY > g.ScreenHeight+60 {
			continue
		}
		enemy.SetPosition(rl.Vector2{X: origPos.X + sx, Y: screenY})
		enemy.Draw()
		enemy.SetPosition(origPos)
	}
}

// drawPlayer renders the player jet in screen space.
func (g *Game) drawPlayer(sx, sy float32) {
	if !g.Player.IsActive() {
		return
	}
	origPos := g.Player.Position
	screenY := origPos.Y - g.CameraY + sy
	g.Player.Position.X = origPos.X + sx
	g.Player.Position.Y = screenY
	g.Player.Draw()
	g.Player.Position = origPos
}

// drawBullets renders projectiles.
func (g *Game) drawBullets(sx, sy float32) {
	for _, b := range g.Bullets {
		if !b.IsActive() {
			continue
		}
		origPos := b.Position
		screenY := origPos.Y - g.CameraY + sy
		if screenY < -30 || screenY > g.ScreenHeight+30 {
			continue
		}
		b.Position.X = origPos.X + sx
		b.Position.Y = screenY
		b.Draw()
		b.Position = origPos
	}
}

// drawMissiles renders SAM missiles.
func (g *Game) drawMissiles(sx, sy float32) {
	for _, m := range g.Missiles {
		if !m.IsActive() {
			continue
		}
		origPos := m.Position
		screenY := origPos.Y - g.CameraY + sy
		if screenY < -50 || screenY > g.ScreenHeight+50 {
			continue
		}
		m.Position.X = origPos.X + sx
		m.Position.Y = screenY
		m.Draw()
		m.Position = origPos
	}
}

// drawParticles renders explosion, shockwave, fire, and smoke particles.
func (g *Game) drawParticles(sx, sy float32) {
	// Temporarily transform particle positions to screen coordinates
	for _, p := range g.Particles.GetParticles() {
		if !p.Active {
			continue
		}
		origPos := p.Position
		p.Position.X = origPos.X + sx
		p.Position.Y = origPos.Y - g.CameraY + sy
	}
	g.Particles.Draw()
	// Restore world positions
	for _, p := range g.Particles.GetParticles() {
		if !p.Active {
			continue
		}
		p.Position.X = p.Position.X - sx
		p.Position.Y = p.Position.Y + g.CameraY - sy
	}
}

// drawDeathOverlay renders tactical feedback and the 3-second respawn timer between lives.
func (g *Game) drawDeathOverlay() {
	cx := g.ScreenWidth / 2
	cy := g.ScreenHeight * 0.42

	boxW := float32(420)
	boxH := float32(140)
	boxRec := rl.Rectangle{X: cx - boxW/2, Y: cy - boxH/2, Width: boxW, Height: boxH}

	if g.Player.Lives > 0 {
		ui.DrawBeveledRect(boxRec, 8.0, rl.Color{R: 20, G: 15, B: 25, A: 225}, rl.Color{R: 255, G: 80, B: 60, A: 220}, 2.0)

		// Warning title
		title := "AIRCRAFT DESTROYED"
		tW := rl.MeasureText(title, 22)
		rl.DrawText(title, int32(cx)-tW/2, int32(boxRec.Y+18), 22, rl.Color{R: 255, G: 70, B: 60, A: 255})

		// Reason
		if len(g.GameOverReason) > 0 {
			rW := rl.MeasureText(g.GameOverReason, 13)
			rl.DrawText(g.GameOverReason, int32(cx)-rW/2, int32(boxRec.Y+48), 13, rl.Color{R: 240, G: 200, B: 180, A: 230})
		}

		// Countdown Timer
		secondsLeft := int(math.Ceil(float64(g.RespawnTimer)))
		if secondsLeft < 1 {
			secondsLeft = 1
		}
		countText := fmt.Sprintf("RE-ENGAGING NEXT JET IN %d...", secondsLeft)
		cW := rl.MeasureText(countText, 18)
		rl.DrawText(countText, int32(cx)-cW/2, int32(boxRec.Y+74), 18, rl.Color{R: 255, G: 220, B: 60, A: 255})

		// Reserves left
		resText := fmt.Sprintf("RESERVES REMAINING: %d", g.Player.Lives)
		resW := rl.MeasureText(resText, 12)
		rl.DrawText(resText, int32(cx)-resW/2, int32(boxRec.Y+106), 12, rl.Color{R: 120, G: 210, B: 255, A: 220})
	} else {
		ui.DrawBeveledRect(boxRec, 8.0, rl.Color{R: 30, G: 10, B: 15, A: 235}, rl.Color{R: 255, G: 30, B: 30, A: 240}, 2.0)

		title := "SQUADRON DEPLETED"
		tW := rl.MeasureText(title, 24)
		rl.DrawText(title, int32(cx)-tW/2, int32(boxRec.Y+24), 24, rl.Color{R: 255, G: 40, B: 40, A: 255})

		sub := "ALL 3 PLANES HAVE BEEN DESTROYED"
		sW := rl.MeasureText(sub, 14)
		rl.DrawText(sub, int32(cx)-sW/2, int32(boxRec.Y+60), 14, rl.Color{R: 255, G: 200, B: 200, A: 240})

		endText := "PREPARING MISSION DEBRIEF..."
		eW := rl.MeasureText(endText, 14)
		rl.DrawText(endText, int32(cx)-eW/2, int32(boxRec.Y+92), 14, rl.Color{R: 255, G: 220, B: 80, A: 240})
	}
}
