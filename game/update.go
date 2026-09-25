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
	g.updateSystemStatus(dt)

	switch g.State {
	case StateDying:
		g.updateDyingState(dt)
	case StateGameOver:
		g.updateGameOverState(dt)
	case StatePlaying:
		g.updateGameplay(dt)
	}
}

// updateSystemStatus manages global timers, shaders, and UI state.
func (g *Game) updateSystemStatus(dt float32) {
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
}

// updateDyingState handles the delay and transition after player destruction.
func (g *Game) updateDyingState(dt float32) {
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
}

// updateGameOverState manages the countdown to return to the title screen.
func (g *Game) updateGameOverState(dt float32) {
	g.GameOverTimer -= dt
	if g.GameOverTimer <= 0 {
		g.State = StateTitle
		g.Menu.Age = 0 // Reset menu animation
	}
}

// updateGameplay handles the core active flight and combat logic.
func (g *Game) updateGameplay(dt float32) {
	g.updateWorldScrolling(dt)
	g.updatePlayerState(dt)
	g.updateAudioAlerts(dt)
	g.updateWorldGeneration(dt)
	g.updateEnemies(dt)
	g.updateProjectiles(dt)
	g.Particles.Update(dt)
	g.checkCollisions(dt)
	g.updateScoreAndMilestones(dt)
}

// updateWorldScrolling advances the river and moves the player forward.
func (g *Game) updateWorldScrolling(dt float32) {
	scrollDelta := g.ScrollSpeed * g.Player.SpeedMultiplier * dt
	g.CameraY -= scrollDelta
	g.Player.Position.Y -= scrollDelta
}

// updatePlayerState handles fuel, input response, and screen boundary clamping.
func (g *Game) updatePlayerState(dt float32) {
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
	}
}

// updateAudioAlerts triggers tactical warning sounds for fuel and incoming missiles.
func (g *Game) updateAudioAlerts(dt float32) {
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
}

// updateWorldGeneration handles procedural spawning and automated entity logic.
func (g *Game) updateWorldGeneration(dt float32) {
	g.World.GenerateAhead(g.CameraY - g.ScreenHeight*1.8)
	g.World.CleanupBehind(g.CameraY + g.ScreenHeight*1.2)

	// Update active decorations
	for _, deco := range g.World.Decorations {
		deco.Update(dt)
	}

	// Update active bridges
	for _, bridge := range g.World.Bridges {
		bridge.Update(dt)
		g.updateBridgeCombat(bridge, dt)
	}
}

// updateBridgeCombat handles bridge vehicle AI firing.
func (g *Game) updateBridgeCombat(bridge *sprites.Bridge, dt float32) {
	if g.World.CurrentSection < 3 || bridge.Destroyed || !bridge.Active {
		return
	}

	if bridge.FireCooldown > 0 {
		bridge.FireCooldown -= dt
		return
	}

	// Check if player is within range and in front of the bridge
	vehiclePos := rl.Vector2{X: bridge.VehicleX, Y: bridge.Position.Y}
	dist := rl.Vector2Distance(vehiclePos, g.Player.Position)

	// Fire if player is within 400px and vertically close
	if dist < 400.0 && g.Player.Position.Y > bridge.Position.Y-350 {
		dx := g.Player.Position.X - vehiclePos.X
		dy := g.Player.Position.Y - vehiclePos.Y
		angle := math.Atan2(float64(dy), float64(dx))

		bulletSpeed := float32(280.0)
		bulletVel := rl.Vector2{
			X: float32(math.Cos(angle)) * bulletSpeed,
			Y: float32(math.Sin(angle)) * bulletSpeed,
		}

		g.Bullets = append(g.Bullets, sprites.NewBullet(vehiclePos, bulletVel, false))
		bridge.FireCooldown = 2.5 + float32(math.Max(0, 1.5-float64(g.World.CurrentSection)*0.1))
		g.Audio.Play(audio.SoundShoot)
	}
}

// updateProjectiles manages movement and culling of bullets and missiles.
func (g *Game) updateProjectiles(dt float32) {
	// Update and cull Bullets
	aliveBullets := 0
	for i := 0; i < len(g.Bullets); i++ {
		b := g.Bullets[i]
		if !b.IsActive() {
			continue
		}
		b.Update(dt)

		g.Bullets[aliveBullets] = b
		aliveBullets++
	}
	g.Bullets = g.Bullets[:aliveBullets]

	// Update and cull Missiles
	g.updateMissiles(dt)
}

// updateScoreAndMilestones handles life awards and high score synchronization.
func (g *Game) updateScoreAndMilestones(dt float32) {
	if g.Player.Score >= g.ScoreForNextLife {
		g.Player.Lives++
		g.ScoreForNextLife += 10000
		g.Audio.Play(audio.SoundExtraLife)
		g.HUD.SetAlert("BONUS LIFE AWARDED!", 2.5, rl.Color{R: 50, G: 255, B: 150, A: 255})
	}

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
		case *sprites.Submarine:
			e.UpdateWithRiverBounds(dt, leftBank, rightBank, hasIsland, islLeft, islRight)
			// Submarine Firing Logic: Surfaced + Player within 25% height radius
			if e.State == sprites.SubStateSurfaced && e.FireCooldown <= 0 && g.Player.Active && g.Player.InvincibleTimer <= 0 {
				distY := math.Abs(float64(e.Position.Y - g.Player.Position.Y))
				if distY < float64(g.ScreenHeight*0.25) {
					missile := sprites.NewSubMissile(e.Position, g.Player)
					g.Missiles = append(g.Missiles, missile)
					e.FireCooldown = 3.0 // Cooldown for firing
					g.Audio.Play(audio.SoundShoot)
				}
			}
		case *sprites.Destroyer:
			// Find the nearest undestroyed bridge "behind" the destroyer (the one it would hit if it sailed down-river)
			// Bridges are at -3600, -7200, etc. Destroyer sails towards more positive Y.
			limitY := float32(2000.0) // Default limit well behind the start
			for _, b := range g.World.Bridges {
				// Only undestroyed bridges act as physical barriers
				if !b.Destroyed && b.Position.Y > e.Position.Y && b.Position.Y < limitY {
					limitY = b.Position.Y
				}
			}
			e.UpdateWithHunterLogic(dt, g.Player.Position, leftBank, rightBank, hasIsland, islLeft, islRight, limitY)
			// Destroyer Firing Logic: Shoot until player passes
			if e.FireCooldown <= 0 && g.Player.Active && g.Player.InvincibleTimer <= 0 {
				dist := rl.Vector2Distance(e.Position, g.Player.Position)
				if dist < 450.0 && g.Player.Position.Y > e.Position.Y {
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
				horizontalDist := math.Abs(float64(e.Position.X - g.Player.Position.X))

				if dist < e.DetectionRange && horizontalDist < float64(g.ScreenWidth*0.25) {
					missile := sprites.NewMissile(e.Position, g.Player)
					g.Missiles = append(g.Missiles, missile)
					e.FireCooldown = 4.0
					g.Audio.Play(audio.SoundShoot)
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
		pos := m.GetPosition()
		if pos.Y < g.CameraY-200 || pos.Y > g.ScreenHeight+g.CameraY+200 {
			m.SetActive(false)
			continue
		}

		g.Missiles[aliveMissiles] = m
		aliveMissiles++
	}
	g.Missiles = g.Missiles[:aliveMissiles]
}

// checkCollisions orchestrates all intersection tests between active entities.
func (g *Game) checkCollisions(dt float32) {
	playerHitbox := rl.Rectangle{
		X:      g.Player.Position.X - 8,
		Y:      g.Player.Position.Y - 12,
		Width:  16,
		Height: 24,
	}

	g.checkBulletCollisions(playerHitbox)

	if !g.Player.IsActive() {
		return
	}

	if g.Player.InvincibleTimer > 0 {
		return
	}

	g.checkRefuelingCollisions(dt)
	g.checkEnemyCrashCollisions(playerHitbox)
	g.checkMissileCrashCollisions(playerHitbox)
	g.checkBridgeCrashCollisions(playerHitbox)
	g.checkTerrainCollisions()
}

// checkBulletCollisions tests projecticles against units, missiles, and bridges.
func (g *Game) checkBulletCollisions(playerHitbox rl.Rectangle) {
	for _, b := range g.Bullets {
		if !b.IsActive() {
			continue
		}
		bulletBounds := b.GetBounds()

		// Get bullet type if we need specific fields
		bullet, isBullet := b.(*sprites.Bullet)
		if !isBullet {
			continue
		}

		// 1. Player Bullets vs Enemies
		if bullet.IsPlayerBullet {
			g.checkPlayerBulletVsEnemies(bullet, bulletBounds)
		}

		if !bullet.IsActive() {
			continue
		}

		// 2. Bullet vs Missiles
		g.checkBulletVsMissiles(bullet, bulletBounds)

		if !bullet.IsActive() {
			continue
		}

		// 3. Bullet vs Bridges
		g.checkBulletVsBridges(bullet, bulletBounds)

		if !bullet.IsActive() {
			continue
		}

		// 4. Enemy Bullet vs Player
		if !bullet.IsPlayerBullet && g.Player.Active && g.Player.InvincibleTimer <= 0 {
			if rl.CheckCollisionRecs(bulletBounds, playerHitbox) {
				bullet.SetActive(false)
				g.triggerPlayerDeath("SHOT DOWN BY ENEMY FIRE")
			}
		}
	}
}

func (g *Game) checkPlayerBulletVsEnemies(bullet *sprites.Bullet, bulletBounds rl.Rectangle) {
	for _, enemy := range g.World.Enemies {
		if !enemy.IsActive() {
			continue
		}
		// Submarines only collide when surfaced or surfacing
		if sub, ok := enemy.(*sprites.Submarine); ok {
			if sub.State == sprites.SubStateSubmerged {
				continue
			}
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
			case *sprites.Submarine:
				pts = e.ScoreValue
			}

			g.Player.Score += pts
			g.Audio.Play(audio.SoundExplosion)
			g.Particles.AddExplosion(enemy.GetPosition(), false)
			break
		}
	}
}

func (g *Game) checkBulletVsMissiles(bullet *sprites.Bullet, bulletBounds rl.Rectangle) {
	// Only player bullets can destroy missiles
	if !bullet.IsPlayerBullet {
		return
	}

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
}

func (g *Game) checkBulletVsBridges(bullet *sprites.Bullet, bulletBounds rl.Rectangle) {
	// Only player bullets can destroy bridges
	if !bullet.IsPlayerBullet {
		return
	}

	for _, bridge := range g.World.Bridges {
		if !bridge.IsActive() || bridge.Destroyed {
			continue
		}

		// Ensure the bridge is on-screen (in sight) for hits to register
		if bridge.Position.Y < g.CameraY || bridge.Position.Y > g.CameraY+g.ScreenHeight {
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
			bridge.Health--

			if bridge.Health <= 0 {
				bridge.Destroy()
				g.Player.Score += bridge.ScoreValue

				g.Audio.Play(audio.SoundBigExplosion)
				g.AddScreenShake(9.0)

				midPos := bridge.GetPosition()
				g.Particles.AddExplosion(midPos, true)
				g.Particles.AddExplosion(rl.Vector2{X: midPos.X - 40, Y: midPos.Y}, false)
				g.Particles.AddExplosion(rl.Vector2{X: midPos.X + 40, Y: midPos.Y}, false)

				alertMsg := fmt.Sprintf("SECTOR %02d SECURED - +500 PTS", bridge.SectionIndex)
				g.HUD.SetAlert(alertMsg, 3.0, rl.Color{R: 0, G: 255, B: 200, A: 255})
			} else {
				// Flash or feedback for hit
				g.Audio.Play(audio.SoundExplosion)
				g.AddScreenShake(2.0)
				g.Particles.AddExplosion(bullet.Position, false)
			}
			break
		}
	}
}

// checkRefuelingCollisions manages player interaction with fuel depots.
func (g *Game) checkRefuelingCollisions(dt float32) {
	playerBounds := g.Player.GetBounds()
	for _, enemy := range g.World.Enemies {
		if !enemy.IsActive() {
			continue
		}
		if fuelDepot, ok := enemy.(*sprites.FuelDepot); ok {
			if rl.CheckCollisionRecs(playerBounds, fuelDepot.GetBounds()) {
				g.Player.AddFuel(fuelDepot.FuelAmount * dt)
				g.Audio.PlayFuelRefuel(float64(g.TotalPlayTime))
			}
		}
	}
}

// checkEnemyCrashCollisions tests for direct mid-air contact with enemy units.
func (g *Game) checkEnemyCrashCollisions(playerHitbox rl.Rectangle) {
	for _, enemy := range g.World.Enemies {
		if !enemy.IsActive() {
			continue
		}
		if _, isFuel := enemy.(*sprites.FuelDepot); isFuel {
			continue
		}

		enemyPos := enemy.GetPosition()
		if math.Abs(float64(enemyPos.Y-g.Player.Position.Y)) > 60 {
			continue
		}

		// Submarines only collide when surfaced or surfacing
		if sub, ok := enemy.(*sprites.Submarine); ok {
			if sub.State == sprites.SubStateSubmerged {
				continue
			}
		}

		if rl.CheckCollisionRecs(playerHitbox, enemy.GetBounds()) {
			enemy.SetActive(false)
			g.Particles.AddExplosion(enemy.GetPosition(), false)
			g.triggerPlayerDeath("MIDAIR COLLISION WITH ENEMY UNIT")
			return
		}
	}
}

// checkMissileCrashCollisions tests for player contact with incoming SAM missiles.
func (g *Game) checkMissileCrashCollisions(playerHitbox rl.Rectangle) {
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
}

// checkBridgeCrashCollisions handles impacts with undestroyed infrastructure.
func (g *Game) checkBridgeCrashCollisions(playerHitbox rl.Rectangle) {
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
}

// checkTerrainCollisions performs multi-point sampling to detect crashes into shores or islands.
func (g *Game) checkTerrainCollisions() {
	nosePt := rl.Vector2{X: g.Player.Position.X, Y: g.Player.Position.Y - 12}
	lWingPt := rl.Vector2{X: g.Player.Position.X - 9, Y: g.Player.Position.Y + 2}
	rWingPt := rl.Vector2{X: g.Player.Position.X + 9, Y: g.Player.Position.Y + 2}

	if !g.World.IsPointInWater(nosePt.X, nosePt.Y) ||
		!g.World.IsPointInWater(lWingPt.X, lWingPt.Y) ||
		!g.World.IsPointInWater(rWingPt.X, rWingPt.Y) {
		g.triggerPlayerDeath("CRASHED INTO RIVER EMBANKMENT")
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
