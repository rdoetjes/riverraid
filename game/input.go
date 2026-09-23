package game

import (
	"math"
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
	}
}

// handleFlightControls manages jet steering, acceleration, and autocannon firing.
func (g *Game) handleFlightControls(dt float32) {
	if !g.Player.IsActive() {
		return
	}

	// 1. Lateral Steering / Banking
	steerInput := float32(0.0)
	if rl.IsKeyDown(rl.KeyLeft) || rl.IsKeyDown(rl.KeyA) {
		steerInput -= 1.0
	}
	if rl.IsKeyDown(rl.KeyRight) || rl.IsKeyDown(rl.KeyD) {
		steerInput += 1.0
	}

	g.Player.TargetBank = steerInput
	g.Player.Velocity.X = steerInput * PlayerLateralSpeed

	// 2. Throttle & Speed Control (Up = Afterburner, Down = Airbrake)
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

	// 3. Autocannon / Laser Firing
	firePressed := rl.IsKeyDown(rl.KeySpace) || rl.IsKeyDown(rl.KeyJ) ||
		rl.IsMouseButtonDown(rl.MouseLeftButton) || rl.IsKeyDown(rl.KeyLeftControl)

	if firePressed && g.Player.CanShoot() {
		// Limit player to maximum 3 bullets on screen
		playerBulletCount := 0
		for _, b := range g.Bullets {
			if b.IsPlayerBullet && b.Active {
				playerBulletCount++
			}
		}

		if playerBulletCount < 3 {
			// Spawn bullet traveling in the direction the nose is pointing (based on BankAngle)
			bulletSpeed := float32(650.0)

			// Calculate horizontal component from banking angle (0.0 center, -1.0 left, 1.0 right)
			// We'll give it a max diagonal tilt of about 15-20 degrees
			horizontalFactor := g.Player.BankAngle * 0.35

			bulletVel := rl.Vector2{
				X: horizontalFactor * bulletSpeed,
				Y: -bulletSpeed,
			}

			// Normalize velocity to keep consistent speed regardless of direction
			mag := float32(math.Sqrt(float64(bulletVel.X*bulletVel.X + bulletVel.Y*bulletVel.Y)))
			bulletVel.X = (bulletVel.X / mag) * bulletSpeed
			bulletVel.Y = (bulletVel.Y / mag) * bulletSpeed

			bulletPos := rl.Vector2{
				X: g.Player.Position.X + horizontalFactor*15.0, // Offset spawn point slightly based on tilt
				Y: g.Player.Position.Y - 24,
			}

			bullet := sprites.NewBullet(bulletPos, bulletVel, true)
			g.Bullets = append(g.Bullets, bullet)

			g.Player.RecordShot()
			g.Audio.Play(audio.SoundShoot)

			// Muzzle flash particle
			g.Particles.AddContrail(bulletPos, rl.Vector2{X: 0, Y: -50}, false)
		}
	}
}
