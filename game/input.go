package game

import (
	"math"
	"strings"

	"riverraid/audio"
	"riverraid/sprites"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// HandleInput processes all keyboard, mouse, and gamepad inputs based on current game state.
func (g *Game) HandleInput(dt float32) {
	// Global toggle audio mute
	if rl.IsKeyPressed(rl.KeyM) {
		g.Audio.ToggleMute()
	}

	switch g.State {
	case StateTitle:
		if rl.IsKeyPressed(rl.KeySpace) || rl.IsKeyPressed(rl.KeyEnter) || rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			g.StartNewGame()
			g.Audio.Play(audio.SoundShoot)
		}

	case StatePlaying:
		g.handleFlightControls(dt)

		if rl.IsKeyPressed(rl.KeyP) || rl.IsKeyPressed(rl.KeyEscape) {
			g.State = StatePaused
		}

	case StatePaused:
		if rl.IsKeyPressed(rl.KeyP) || rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeySpace) {
			g.State = StatePlaying
		}

	case StateGameOver:
		if rl.IsKeyPressed(rl.KeySpace) || rl.IsKeyPressed(rl.KeyEnter) || rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			g.StartNewGame()
			g.Audio.Play(audio.SoundShoot)
		}

	case StateEnteringName:
		// Handle text input for initials (3 letters)
		key := rl.GetCharPressed()
		for key > 0 {
			if len(g.EnterNameBuffer) < 3 {
				// Only allow uppercase A-Z
				if key >= 65 && key <= 90 {
					g.EnterNameBuffer += string(key)
				} else if key >= 97 && key <= 122 {
					g.EnterNameBuffer += strings.ToUpper(string(key))
				}
			}
			key = rl.GetCharPressed()
		}

		if rl.IsKeyPressed(rl.KeyBackspace) && len(g.EnterNameBuffer) > 0 {
			g.EnterNameBuffer = g.EnterNameBuffer[:len(g.EnterNameBuffer)-1]
		}

		if rl.IsKeyPressed(rl.KeyEnter) && len(g.EnterNameBuffer) == 3 {
			g.AddHighScore(g.EnterNameBuffer, g.Player.Score)
			g.ScoreSubmitted = true
			g.State = StateGameOver
			g.GameOverTimer = 10.0
			g.Audio.Play(audio.SoundShoot)
		}
	}
}

// handleFlightControls orchestrates player movement and combat inputs.
func (g *Game) handleFlightControls(dt float32) {
	if !g.Player.IsActive() {
		return
	}

	g.handleLateralControls(dt)
	g.handleThrottleControls(dt)
	g.handleCombatControls(dt)
}

// handleLateralControls manages banking and side-to-side movement.
func (g *Game) handleLateralControls(dt float32) {
	steerInput := float32(0.0)
	if rl.IsKeyDown(rl.KeyLeft) || rl.IsKeyDown(rl.KeyA) {
		steerInput -= 1.0
	}
	if rl.IsKeyDown(rl.KeyRight) || rl.IsKeyDown(rl.KeyD) {
		steerInput += 1.0
	}

	g.Player.TargetBank = steerInput
	g.Player.Velocity.X = steerInput * PlayerLateralSpeed
}

// handleThrottleControls manages afterburners, airbrakes, and exhaust effects.
func (g *Game) handleThrottleControls(dt float32) {
	if rl.IsKeyDown(rl.KeyUp) || rl.IsKeyDown(rl.KeyW) {
		g.Player.TargetSpeedMul = 1.45
		// Emit exhaust particles
		if rl.GetRandomValue(0, 10) < 4 {
			flamePos := rl.Vector2{X: g.Player.Position.X, Y: g.Player.Position.Y + 20}
			g.Particles.AddContrail(flamePos, rl.Vector2{X: 0, Y: 40}, true)
		}
	} else if rl.IsKeyDown(rl.KeyDown) || rl.IsKeyDown(rl.KeyS) {
		g.Player.TargetSpeedMul = 0.65
	} else {
		g.Player.TargetSpeedMul = 1.0
	}
}

// handleCombatControls manages autocannon firing logic, including diagonal shot calculations.
func (g *Game) handleCombatControls(dt float32) {
	firePressed := rl.IsKeyDown(rl.KeySpace) || rl.IsKeyDown(rl.KeyJ) ||
		rl.IsMouseButtonDown(rl.MouseLeftButton) || rl.IsKeyDown(rl.KeyLeftControl)

	if !firePressed || !g.Player.CanShoot() {
		return
	}

	// Limit player to maximum 3 bullets on screen
	playerBulletCount := 0
	for _, b := range g.Bullets {
		if b.IsPlayerBullet && b.Active {
			playerBulletCount++
		}
	}

	if playerBulletCount < 3 {
		// Calculate trajectory based on banking (diagonal firing)
		bulletSpeed := float32(650.0)
		horizontalFactor := g.Player.BankAngle * 0.35

		bulletVel := rl.Vector2{
			X: horizontalFactor * bulletSpeed,
			Y: -bulletSpeed,
		}

		// Normalize velocity to maintain consistent kinetic energy
		mag := float32(math.Sqrt(float64(bulletVel.X*bulletVel.X + bulletVel.Y*bulletVel.Y)))
		bulletVel.X = (bulletVel.X / mag) * bulletSpeed
		bulletVel.Y = (bulletVel.Y / mag) * bulletSpeed

		// Align spawn point with the tilted nose
		bulletPos := rl.Vector2{
			X: g.Player.Position.X + horizontalFactor*15.0,
			Y: g.Player.Position.Y - 24,
		}

		bullet := sprites.NewBullet(bulletPos, bulletVel, true)
		g.Bullets = append(g.Bullets, bullet)

		g.Player.RecordShot()
		g.Audio.Play(audio.SoundShoot)

		// Muzzle flash / Heat distortion particle
		g.Particles.AddContrail(bulletPos, rl.Vector2{X: 0, Y: -50}, false)
	}
}
