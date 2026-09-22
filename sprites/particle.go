package sprites

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ParticleKind specifies the visual behavior of the particle.
type ParticleKind int

const (
	ParticleFire ParticleKind = iota
	ParticleSmoke
	ParticleShockwave
	ParticleWaterSplash
	ParticleDebris
	ParticleContrail
)

// Particle represents a single visual FX element.
type Particle struct {
	BaseSprite
	Kind       ParticleKind
	MaxAge     float32
	StartSize  float32
	EndSize    float32
	Color      rl.Color
	EndColor   rl.Color
	Rotation   float32
	RotSpeed   float32
	Drag       float32
	RingRadius float32
	MaxRadius  float32
}

// ParticleSystem maintains a pool of active particles.
type ParticleSystem struct {
	particles []*Particle
}

// NewParticleSystem creates an empty particle manager.
func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		particles: make([]*Particle, 0, 512),
	}
}

// AddExplosion emits a rich multi-stage explosion at a position.
func (ps *ParticleSystem) AddExplosion(pos rl.Vector2, isBig bool) {
	sparkCount := 28
	smokeCount := 16
	debrisCount := 10
	shockwaveRadius := float32(45.0)

	if isBig {
		sparkCount = 60
		smokeCount = 35
		debrisCount = 20
		shockwaveRadius = 90.0
	}

	// Shockwave
	ps.particles = append(ps.particles, &Particle{
		BaseSprite: BaseSprite{
			Position: pos,
			Active:   true,
		},
		Kind:       ParticleShockwave,
		MaxAge:     0.45,
		RingRadius: 4.0,
		MaxRadius:  shockwaveRadius,
		Color:      rl.Color{R: 255, G: 240, B: 180, A: 240},
	})

	// Sparks / Fire
	for i := 0; i < sparkCount; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 50.0 + rand.Float64()*180.0
		if isBig {
			speed *= 1.4
		}
		vel := rl.Vector2{
			X: float32(math.Cos(angle) * speed),
			Y: float32(math.Sin(angle) * speed),
		}
		ps.particles = append(ps.particles, &Particle{
			BaseSprite: BaseSprite{
				Position: pos,
				Velocity: vel,
				Active:   true,
			},
			Kind:      ParticleFire,
			MaxAge:    0.35 + rand.Float32()*0.4,
			StartSize: 4.0 + rand.Float32()*4.0,
			EndSize:   0.5,
			Color:     rl.Color{R: 255, G: uint8(160 + rand.Intn(95)), B: 40, A: 255},
			EndColor:  rl.Color{R: 240, G: 50, B: 20, A: 0},
			Drag:      0.92,
		})
	}

	// Smoke puffs
	for i := 0; i < smokeCount; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 20.0 + rand.Float64()*60.0
		vel := rl.Vector2{
			X: float32(math.Cos(angle) * speed),
			Y: float32(math.Sin(angle) * speed),
		}
		ps.particles = append(ps.particles, &Particle{
			BaseSprite: BaseSprite{
				Position: pos,
				Velocity: vel,
				Active:   true,
			},
			Kind:      ParticleSmoke,
			MaxAge:    0.6 + rand.Float32()*0.6,
			StartSize: 8.0 + rand.Float32()*6.0,
			EndSize:   24.0 + rand.Float32()*16.0,
			Color:     rl.Color{R: 80, G: 80, B: 85, A: 190},
			EndColor:  rl.Color{R: 40, G: 40, B: 45, A: 0},
			Drag:      0.88,
		})
	}

	// Debris chunks
	for i := 0; i < debrisCount; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 60.0 + rand.Float64()*120.0
		vel := rl.Vector2{
			X: float32(math.Cos(angle) * speed),
			Y: float32(math.Sin(angle) * speed),
		}
		ps.particles = append(ps.particles, &Particle{
			BaseSprite: BaseSprite{
				Position: pos,
				Velocity: vel,
				Active:   true,
			},
			Kind:      ParticleDebris,
			MaxAge:    0.5 + rand.Float32()*0.5,
			StartSize: 4.0 + rand.Float32()*5.0,
			EndSize:   2.0,
			Color:     rl.Color{R: 200, G: 190, B: 180, A: 255},
			Rotation:  rand.Float32() * math.Pi * 2,
			RotSpeed:  (rand.Float32() - 0.5) * 15.0,
			Drag:      0.94,
		})
	}
}

// AddWaterSplash emits ripples and white foam when something crashes into water.
func (ps *ParticleSystem) AddWaterSplash(pos rl.Vector2) {
	// Water ripple ring
	ps.particles = append(ps.particles, &Particle{
		BaseSprite: BaseSprite{
			Position: pos,
			Active:   true,
		},
		Kind:       ParticleWaterSplash,
		MaxAge:     0.7,
		RingRadius: 2.0,
		MaxRadius:  38.0,
		Color:      rl.Color{R: 200, G: 240, B: 255, A: 220},
	})

	// Spray drops
	for i := 0; i < 14; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 30.0 + rand.Float64()*70.0
		vel := rl.Vector2{
			X: float32(math.Cos(angle) * speed),
			Y: float32(math.Sin(angle) * speed),
		}
		ps.particles = append(ps.particles, &Particle{
			BaseSprite: BaseSprite{
				Position: pos,
				Velocity: vel,
				Active:   true,
			},
			Kind:      ParticleFire,
			MaxAge:    0.35 + rand.Float32()*0.3,
			StartSize: 3.5,
			EndSize:   0.8,
			Color:     rl.Color{R: 220, G: 245, B: 255, A: 230},
			EndColor:  rl.Color{R: 150, G: 210, B: 240, A: 0},
			Drag:      0.90,
		})
	}
}

// AddContrail adds an exhaust smoke/plasma particle.
func (ps *ParticleSystem) AddContrail(pos rl.Vector2, vel rl.Vector2, isAfterburner bool) {
	col := rl.Color{R: 180, G: 220, B: 240, A: 120}
	endCol := rl.Color{R: 120, G: 160, B: 200, A: 0}
	startSize := float32(3.0)
	endSize := float32(7.0)
	maxAge := float32(0.25)

	if isAfterburner {
		col = rl.Color{R: 0, G: 200, B: 255, A: 220}
		endCol = rl.Color{R: 255, G: 120, B: 20, A: 0}
		startSize = 5.0
		endSize = 1.0
		maxAge = 0.18
	}

	ps.particles = append(ps.particles, &Particle{
		BaseSprite: BaseSprite{
			Position: pos,
			Velocity: vel,
			Active:   true,
		},
		Kind:      ParticleContrail,
		MaxAge:    maxAge,
		StartSize: startSize,
		EndSize:   endSize,
		Color:     col,
		EndColor:  endCol,
		Drag:      0.95,
	})
}

// GetParticles returns the active slice of particles.
func (ps *ParticleSystem) GetParticles() []*Particle {
	return ps.particles
}

// Update advances all particles.
func (ps *ParticleSystem) Update(dt float32) {
	aliveCount := 0
	for i := 0; i < len(ps.particles); i++ {
		p := ps.particles[i]
		if !p.Active {
			continue
		}

		p.Age += dt
		if p.Age >= p.MaxAge {
			p.Active = false
			continue
		}

		// Physics update
		p.Position.X += p.Velocity.X * dt
		p.Position.Y += p.Velocity.Y * dt
		p.Velocity.X *= p.Drag
		p.Velocity.Y *= p.Drag
		p.Rotation += p.RotSpeed * dt

		// Compact in-place
		ps.particles[aliveCount] = p
		aliveCount++
	}
	ps.particles = ps.particles[:aliveCount]
}

// Draw renders all active particles.
func (ps *ParticleSystem) Draw() {
	for _, p := range ps.particles {
		if !p.Active {
			continue
		}
		progress := p.Age / p.MaxAge
		invProgress := 1.0 - progress

		switch p.Kind {
		case ParticleShockwave:
			currentRadius := p.RingRadius + (p.MaxRadius-p.RingRadius)*progress
			alpha := uint8(float32(p.Color.A) * invProgress)
			c := rl.Color{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
			rl.DrawCircleLines(int32(p.Position.X), int32(p.Position.Y), currentRadius, c)
			rl.DrawCircleLines(int32(p.Position.X), int32(p.Position.Y), currentRadius-1, c)

		case ParticleWaterSplash:
			currentRadius := p.RingRadius + (p.MaxRadius-p.RingRadius)*progress
			alpha := uint8(float32(p.Color.A) * invProgress)
			c := rl.Color{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
			rl.DrawCircleLines(int32(p.Position.X), int32(p.Position.Y), currentRadius, c)

		case ParticleFire, ParticleContrail:
			size := p.StartSize + (p.EndSize-p.StartSize)*progress
			alpha := uint8(float32(p.Color.A) * invProgress)
			r := uint8(float32(p.Color.R)*invProgress + float32(p.EndColor.R)*progress)
			g := uint8(float32(p.Color.G)*invProgress + float32(p.EndColor.G)*progress)
			b := uint8(float32(p.Color.B)*invProgress + float32(p.EndColor.B)*progress)
			c := rl.Color{R: r, G: g, B: b, A: alpha}
			rl.DrawCircleV(p.Position, size, c)

		case ParticleSmoke:
			size := p.StartSize + (p.EndSize-p.StartSize)*progress
			alpha := uint8(float32(p.Color.A) * invProgress)
			c := rl.Color{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
			rl.DrawCircleV(p.Position, size, c)

		case ParticleDebris:
			size := p.StartSize + (p.EndSize-p.StartSize)*progress
			alpha := uint8(255 * invProgress)
			c := rl.Color{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
			half := size / 2
			p1 := rl.Vector2{X: p.Position.X - half, Y: p.Position.Y - half}
			p2 := rl.Vector2{X: p.Position.X + half, Y: p.Position.Y - half}
			p3 := rl.Vector2{X: p.Position.X, Y: p.Position.Y + half}
			rl.DrawTriangle(p1, p2, p3, c)
		}
	}
}

// Clear removes all active particles.
func (ps *ParticleSystem) Clear() {
	ps.particles = ps.particles[:0]
}
