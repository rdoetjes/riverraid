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

	// Update shader time uniform
	if g.WaterShader.ID > 0 {
		rl.SetShaderValue(g.WaterShader, g.TimeLoc, []float32{g.TotalPlayTime}, rl.ShaderUniformFloat)
	}

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
		g.updateEnemies(dt)
		g.Particles.Update(dt)

		g.RespawnTimer -= dt
		if g.RespawnTimer <= 0 {
			if g.Player.Lives > 0 {
				g.RespawnPlayer()
				g.State = StatePlaying
			} else {
				if g.IsNewHighScore(g.Player.Score) {
					g.State = StateEnteringName
					g.EnterNameBuffer = ""
				} else {
					g.State = StateGameOver
					g.GameOverTimer = 10.0
				}
			}
		}
		return
	}

	if g.State == StateGameOver {
		g.GameOverTimer -= dt
		if g.GameOverTimer <= 0 {
			g.State = StateTitle
			g.Menu.Age = 0 // Reset menu animation
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

	// SAM Missile active warning beep
	if len(g.Missiles) > 0 && g.Player.IsActive() {
		// Fast beep (3.5 times per second)
		if int(g.TotalPlayTime*7)%2 == 0 && int((g.TotalPlayTime-dt)*7)%2 != 0 {
			g.Audio.Play(audio.SoundMissileWarning)
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

		// From level 3, bridge vehicles shoot at the player
		if g.World.CurrentSection >= 3 && !bridge.Destroyed && bridge.Active {
			if bridge.FireCooldown > 0 {
				bridge.FireCooldown -= dt
			} else {
				// Check if player is within range and in front of the bridge
				vehiclePos := rl.Vector2{X: bridge.VehicleX, Y: bridge.Position.Y}
				dist := rl.Vector2Distance(vehiclePos, g.Player.Position)

				// Fire if player is within 400px and vertically close (to simulate tactical firing)
				if dist < 400.0 && g.Player.Position.Y > bridge.Position.Y-350 {
					// Aim bullet towards player
					dx := g.Player.Position.X - vehiclePos.X
					dy := g.Player.Position.Y - vehiclePos.Y
					angle := math.Atan2(float64(dy), float64(dx))

					bulletSpeed := float32(280.0)
					bulletVel := rl.Vector2{
						X: float32(math.Cos(angle)) * bulletSpeed,
						Y: float32(math.Sin(angle)) * bulletSpeed,
					}

					bullet := sprites.NewBullet(vehiclePos, bulletVel, false)
					g.Bullets = append(g.Bullets, bullet)

					bridge.FireCooldown = 2.5 + float32(math.Max(0, 1.5-float64(g.World.CurrentSection)*0.1)) // Faster firing at higher levels
					g.Audio.Play(audio.SoundShoot)
				}
			}
		}
	}

	// Update active enemies with dynamic riverbank & island boundary detection
	g.updateEnemies(dt)

	// 4. Update Bullets & Missiles
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

	g.updateMissiles(dt)

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

// updateEnemies updates enemy movement and ensures watercraft/aircraft reverse when hitting riverbanks or islands.
func (g *Game) updateEnemies(dt float32) {
	for _, enemy := range g.World.Enemies {
		if !enemy.IsActive() {
			continue
		}
		pos := enemy.GetPosition()
		leftBank, rightBank, hasIsland, islLeft, islRight := g.World.GetRiverBoundsAt(pos.Y)

		switch e := enemy.(type) {
		case *sprites.Ship:
			e.UpdateWithRiverBounds(dt, leftBank, rightBank, hasIsland, islLeft, islRight)
		case *sprites.Helicopter:
			e.UpdateWithRiverBounds(dt, leftBank, rightBank, hasIsland, islLeft, islRight)
		case *sprites.Destroyer:
			e.UpdateWithHunterLogic(dt, g.Player.Position, leftBank, rightBank, hasIsland, islLeft, islRight)
			// Destroyer Firing Logic: Shoot until player passes
			if e.FireCooldown <= 0 && g.Player.Active && g.Player.InvincibleTimer <= 0 {
				dist := rl.Vector2Distance(e.Position, g.Player.Position)
				// playerY > e.Position.Y means player is "south" (behind) of the ship
				if dist < 450.0 && g.Player.Position.Y > e.Position.Y {
					// Fire bullet towards player
					dx := g.Player.Position.X - e.Position.X
					dy := g.Player.Position.Y - e.Position.Y
					angle := math.Atan2(float64(dy), float64(dx))

					bulletSpeed := float32(320.0)
					bulletVel := rl.Vector2{
						X: float32(math.Cos(angle)) * bulletSpeed,
						Y: float32(math.Sin(angle)) * bulletSpeed,
					}
					g.Bullets = append(g.Bullets, sprites.NewBullet(e.Position, bulletVel, false))
					e.FireCooldown = 2.0
					g.Audio.Play(audio.SoundShoot)
				}
			}
		case *sprites.SAMSite:
			e.Update(dt)
			// SAM Site Firing Logic
			if e.FireCooldown <= 0 && g.Player.Active && g.Player.InvincibleTimer <= 0 {
				dist := rl.Vector2Distance(e.Position, g.Player.Position)
				// Horizontal proximity check: Only fire if player is within 25% screen width of the site
				horizontalDist := math.Abs(float64(e.Position.X - g.Player.Position.X))

				if dist < e.DetectionRange && horizontalDist < float64(g.ScreenWidth*0.25) {
					// Fire a missile
					missile := sprites.NewMissile(e.Position, g.Player)
					g.Missiles = append(g.Missiles, missile)
					e.FireCooldown = 4.0           // 4 seconds between shots
					g.Audio.Play(audio.SoundShoot) // Reuse shoot sound for now
				}
			}
		default:
			enemy.Update(dt)
		}
	}
}

// updateMissiles handles homing and lifetime for SAM missiles.
func (g *Game) updateMissiles(dt float32) {
	aliveMissiles := 0
	for i := 0; i < len(g.Missiles); i++ {
		m := g.Missiles[i]
		if !m.IsActive() {
			continue
		}
		m.Update(dt)

		// Cull if too far off screen
		if m.Position.Y < g.CameraY-200 || m.Position.Y > g.CameraY+g.ScreenHeight+200 {
			m.SetActive(false)
			continue
		}

		g.Missiles[aliveMissiles] = m
		aliveMissiles++
	}
	g.Missiles = g.Missiles[:aliveMissiles]
}

// checkCollisions handles bullet impacts, refueling, enemy collisions, and terrain crash tests.
func (g *Game) checkCollisions(dt float32) {
	playerBounds := g.Player.GetBounds()
	// Tighter collision box for player fuselage
	playerHitbox := rl.Rectangle{
		X:      g.Player.Position.X - 8,
		Y:      g.Player.Position.Y - 12,
		Width:  16,
		Height: 24,
	}

	// --- A. Bullet vs Enemies & Fuel Depots ---
	for _, bullet := range g.Bullets {
		if !bullet.IsActive() {
			continue
		}
		bulletBounds := bullet.GetBounds()

		// Bullet vs Enemies: Only player bullets can destroy enemies
		if bullet.IsPlayerBullet {
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
					case *sprites.Destroyer:
						pts = e.ScoreValue
					case *sprites.FuelDepot:
						pts = e.ScoreValue
					case *sprites.SAMSite:
						pts = e.ScoreValue
					}

					g.Player.Score += pts
					g.Audio.Play(audio.SoundExplosion)
					g.Particles.AddExplosion(enemy.GetPosition(), false)
					break
				}
			}
		}

		if !bullet.IsActive() {
			continue
		}

		// Bullet vs Missiles
		for _, missile := range g.Missiles {
			if !missile.IsActive() {
				continue
			}
			if rl.CheckCollisionRecs(bulletBounds, missile.GetBounds()) {
				bullet.SetActive(false)
				missile.SetActive(false)
				g.Audio.Play(audio.SoundExplosion)
				g.Particles.AddExplosion(missile.GetPosition(), false)
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

		if !bullet.IsActive() {
			continue
		}

		// Enemy Bullet vs Player
		if !bullet.IsPlayerBullet && g.Player.Active && g.Player.InvincibleTimer <= 0 {
			if rl.CheckCollisionRecs(bulletBounds, playerHitbox) {
				bullet.SetActive(false)
				g.triggerPlayerDeath("SHOT DOWN BY ENEMY FIRE")
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

		// Vertical proximity check before detailed AABB
		enemyPos := enemy.GetPosition()
		if math.Abs(float64(enemyPos.Y-g.Player.Position.Y)) > 60 {
			continue
		}

		if rl.CheckCollisionRecs(playerHitbox, enemy.GetBounds()) {
			enemy.SetActive(false)
			g.Particles.AddExplosion(enemy.GetPosition(), false)
			g.triggerPlayerDeath("MIDAIR COLLISION WITH ENEMY UNIT")
			return
		}
	}

	// --- C2. Player vs Missiles (Crash) ---
	for _, missile := range g.Missiles {
		if !missile.IsActive() {
			continue
		}
		if rl.CheckCollisionRecs(playerHitbox, missile.GetBounds()) {
			missile.SetActive(false)
			g.Particles.AddExplosion(missile.GetPosition(), false)
			g.triggerPlayerDeath("STRUCK BY HEAT-SEEKING MISSILE")
			return
		}
	}

	// --- D. Player vs Bridges (Crash into undestroyed bridge) ---
	for _, bridge := range g.World.Bridges {
		if !bridge.IsActive() || bridge.Destroyed {
			continue
		}
		if math.Abs(float64(bridge.Position.Y-g.Player.Position.Y)) > 35 {
			continue
		}
		bridgeHitbox := rl.Rectangle{
			X:      bridge.LeftBankX,
			Y:      bridge.Position.Y - 10,
			Width:  bridge.RightBankX - bridge.LeftBankX,
			Height: 20,
		}
		if rl.CheckCollisionRecs(playerHitbox, bridgeHitbox) {
			g.triggerPlayerDeath("COLLIDED WITH RIVER CROSSING BRIDGE")
			return
		}
	}

	// --- E. Player vs River Shorelines / Islands (Terrain Crash) ---
	// Check multiple sample points on player jet (nose, left wing, right wing)
	nosePt := rl.Vector2{X: g.Player.Position.X, Y: g.Player.Position.Y - 12}
	lWingPt := rl.Vector2{X: g.Player.Position.X - 9, Y: g.Player.Position.Y + 2}
	rWingPt := rl.Vector2{X: g.Player.Position.X + 9, Y: g.Player.Position.Y + 2}

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
