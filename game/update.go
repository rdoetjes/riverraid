package game

import (
	"fmt"
	"math"

	"riverraid/audio"
	"riverraid/sprites"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Update orchestrates physics, collision detection, procedural spawning, and game rules.
func (g *Game) Update(dt float32) {
	g.TotalPlayTime += dt
	g.HUD.Update(dt)
	g.Menu.Update(dt)

	// Screen shake dissipation
	if g.ScreenShake > 0.1 {
		g.ScreenShake *= float32(math.Pow(0.12, float64(dt)))
	} else {
		g.ScreenShake = 0
	}

	if g.State == StateDying {
		// Update world scenery and particles during death delay
		for _, deco := range g.World.Decorations {
			deco.Update(dt)
		}
		for _, bridge := range g.World.Bridges {
			bridge.Update(dt)
		}
		for _, enemy := range g.World.Enemies {
			enemy.Update(dt)
		}
		g.Particles.Update(dt)

		g.RespawnTimer -= dt
		if g.RespawnTimer <= 0 {
			if g.Player.Lives > 0 {
				g.RespawnPlayer()
				g.State = StatePlaying
			} else {
				g.State = StateGameOver
			}
		}
		return
	}

	if g.State != StatePlaying {
		return
	}

	// 1. Advance World Scrolling & Camera
	scrollDelta := g.ScrollSpeed * g.Player.SpeedMultiplier * dt
	g.CameraY -= scrollDelta
	g.Player.Position.Y -= scrollDelta

	// 2. Update Player State & Fuel
	g.Player.Refueling = false
	g.Player.Update(dt)

	// Clamp player laterally to screen visible area
	halfW := g.Player.Size.X / 2
	if g.Player.Position.X < halfW+10 {
		g.Player.Position.X = halfW + 10
	} else if g.Player.Position.X > g.ScreenWidth-halfW-10 {
		g.Player.Position.X = g.ScreenWidth - halfW - 10
	}

	// Fuel Exhaustion Check
	if g.Player.Fuel <= 0 && g.Player.IsActive() {
		g.triggerPlayerDeath("OUT OF FUEL - ENGINES FLAMED OUT")
		return
	}

	// Low fuel periodic warning beep
	if g.Player.Fuel < 20.0 && g.Player.Fuel > 0 {
		if int(g.TotalPlayTime*2.5)%2 == 0 && int((g.TotalPlayTime-dt)*2.5)%2 != 0 {
			g.Audio.Play(audio.SoundLowFuel)
		}
	}

	// 3. Update Procedural World Generation & Culling
	g.World.GenerateAhead(g.CameraY - g.ScreenHeight*1.8)
	g.World.CleanupBehind(g.CameraY + g.ScreenHeight*1.2)

	// Update active decorations
	for _, deco := range g.World.Decorations {
		deco.Update(dt)
	}

	// Update active bridges
	for _, bridge := range g.World.Bridges {
		bridge.Update(dt)
	}

	// Update active enemies
	for _, enemy := range g.World.Enemies {
		enemy.Update(dt)
	}

	// 4. Update Bullets
	aliveBullets := 0
	for i := 0; i < len(g.Bullets); i++ {
		b := g.Bullets[i]
		if !b.IsActive() {
			continue
		}
		b.Update(dt)

		// Cull bullet if scrolled off camera
		if b.Position.Y < g.CameraY-100 || b.Position.Y > g.CameraY+g.ScreenHeight+100 {
			b.SetActive(false)
			continue
		}

		g.Bullets[aliveBullets] = b
		aliveBullets++
	}
	g.Bullets = g.Bullets[:aliveBullets]

	// 5. Update Particle FX
	g.Particles.Update(dt)

	// 6. Collision Detection
	g.checkCollisions(dt)

	// 7. Extra Life Milestone Tracking
	if g.Player.Score >= g.ScoreForNextLife {
		g.Player.Lives++
		g.ScoreForNextLife += 10000
		g.Audio.Play(audio.SoundExtraLife)
		g.HUD.SetAlert("BONUS LIFE AWARDED!", 2.5, rl.Color{R: 50, G: 255, B: 150, A: 255})
	}

	// High score update
	if g.Player.Score > g.HighScore {
		g.HighScore = g.Player.Score
	}
}

// checkCollisions handles bullet impacts, refueling, enemy collisions, and terrain crash tests.
func (g *Game) checkCollisions(dt float32) {
	playerBounds := g.Player.GetBounds()
	// Tighter collision box for player fuselage
	playerHitbox := rl.Rectangle{
		X:      g.Player.Position.X - 10,
		Y:      g.Player.Position.Y - 14,
		Width:  20,
		Height: 28,
	}

	// --- A. Bullet vs Enemies & Fuel Depots ---
	for _, bullet := range g.Bullets {
		if !bullet.IsActive() {
			continue
		}
		bulletBounds := bullet.GetBounds()

		// Bullet vs Enemies
		for _, enemy := range g.World.Enemies {
			if !enemy.IsActive() {
				continue
			}
			if rl.CheckCollisionRecs(bulletBounds, enemy.GetBounds()) {
				bullet.SetActive(false)
				enemy.SetActive(false)

				pts := 0
				switch e := enemy.(type) {
				case *sprites.Helicopter:
					pts = e.ScoreValue
				case *sprites.Ship:
					pts = e.ScoreValue
				case *sprites.EnemyJet:
					pts = e.ScoreValue
				case *sprites.FuelDepot:
					pts = e.ScoreValue
				}

				g.Player.Score += pts
				g.Audio.Play(audio.SoundExplosion)
				g.Particles.AddExplosion(enemy.GetPosition(), false)
				break
			}
		}

		if !bullet.IsActive() {
			continue
		}

		// Bullet vs Bridges
		for _, bridge := range g.World.Bridges {
			if !bridge.IsActive() || bridge.Destroyed {
				continue
			}
			bridgeHitbox := rl.Rectangle{
				X:      bridge.LeftBankX,
				Y:      bridge.Position.Y - 16,
				Width:  bridge.RightBankX - bridge.LeftBankX,
				Height: 32,
			}
			if rl.CheckCollisionRecs(bulletBounds, bridgeHitbox) {
				bullet.SetActive(false)
				bridge.Destroy()
				g.Player.Score += bridge.ScoreValue

				g.Audio.Play(audio.SoundBigExplosion)
				g.AddScreenShake(9.0)

				// Massive multi-stage explosion along bridge span
				midPos := bridge.GetPosition()
				g.Particles.AddExplosion(midPos, true)
				g.Particles.AddExplosion(rl.Vector2{X: midPos.X - 40, Y: midPos.Y}, false)
				g.Particles.AddExplosion(rl.Vector2{X: midPos.X + 40, Y: midPos.Y}, false)

				alertMsg := fmt.Sprintf("SECTOR %02d SECURED - +500 PTS", bridge.SectionIndex)
				g.HUD.SetAlert(alertMsg, 3.0, rl.Color{R: 0, G: 255, B: 200, A: 255})
				break
			}
		}
	}

	if !g.Player.IsActive() {
		return
	}

	// --- B. Player vs Fuel Depots (Refueling) ---
	for _, enemy := range g.World.Enemies {
		if !enemy.IsActive() {
			continue
		}
		if fuelDepot, ok := enemy.(*sprites.FuelDepot); ok {
			// Touch depot to refuel
			if rl.CheckCollisionRecs(playerBounds, fuelDepot.GetBounds()) {
				g.Player.AddFuel(fuelDepot.FuelAmount * dt)
				g.Audio.PlayFuelRefuel(float64(g.TotalPlayTime))
			}
		}
	}

	// Skip damage if player is currently in invincibility grace period
	if g.Player.InvincibleTimer > 0 {
		return
	}

	// --- C. Player vs Enemies (Crash) ---
	for _, enemy := range g.World.Enemies {
		if !enemy.IsActive() {
			continue
		}
		// Only collide with combat enemies, not fuel depots
		if _, isFuel := enemy.(*sprites.FuelDepot); isFuel {
			continue
		}

		if rl.CheckCollisionRecs(playerHitbox, enemy.GetBounds()) {
			enemy.SetActive(false)
			g.Particles.AddExplosion(enemy.GetPosition(), false)
			g.triggerPlayerDeath("MIDAIR COLLISION WITH ENEMY UNIT")
			return
		}
	}

	// --- D. Player vs Bridges (Crash into undestroyed bridge) ---
	for _, bridge := range g.World.Bridges {
		if !bridge.IsActive() || bridge.Destroyed {
			continue
		}
		bridgeHitbox := rl.Rectangle{
			X:      bridge.LeftBankX,
			Y:      bridge.Position.Y - 12,
			Width:  bridge.RightBankX - bridge.LeftBankX,
			Height: 24,
		}
		if rl.CheckCollisionRecs(playerHitbox, bridgeHitbox) {
			g.triggerPlayerDeath("COLLIDED WITH RIVER CROSSING BRIDGE")
			return
		}
	}

	// --- E. Player vs River Shorelines / Islands (Terrain Crash) ---
	// Check multiple sample points on player jet (nose, left wing, right wing)
	nosePt := rl.Vector2{X: g.Player.Position.X, Y: g.Player.Position.Y - 14}
	lWingPt := rl.Vector2{X: g.Player.Position.X - 12, Y: g.Player.Position.Y + 4}
	rWingPt := rl.Vector2{X: g.Player.Position.X + 12, Y: g.Player.Position.Y + 4}

	if !g.World.IsPointInWater(nosePt.X, nosePt.Y) ||
		!g.World.IsPointInWater(lWingPt.X, lWingPt.Y) ||
		!g.World.IsPointInWater(rWingPt.X, rWingPt.Y) {
		g.triggerPlayerDeath("CRASHED INTO RIVER EMBANKMENT")
		return
	}
}

// triggerPlayerDeath handles jet destruction, life deduction, and respawn or game over.
func (g *Game) triggerPlayerDeath(reason string) {
	g.Player.SetActive(false)
	g.Player.Lives--

	g.Audio.Play(audio.SoundBigExplosion)
	g.AddScreenShake(14.0)
	g.Particles.AddExplosion(g.Player.GetPosition(), true)
	g.Particles.AddWaterSplash(g.Player.GetPosition())

	g.GameOverReason = reason
	g.RespawnTimer = 3.0
	g.State = StateDying
}
