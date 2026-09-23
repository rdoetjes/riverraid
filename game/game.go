package game

import (
	"math/rand"
	"os"
	"time"

	"riverraid/audio"
	"riverraid/sprites"
	"riverraid/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// GameState defines the current mode of the application.
type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateDying
	StatePaused
	StateGameOver
)

const (
	DefaultScreenWidth  = 720
	DefaultScreenHeight = 650
	BaseScrollSpeed     = 180.0 // Pixels per second
	MinScrollSpeed      = 120.0
	MaxScrollSpeed      = 280.0
	PlayerLateralSpeed  = 240.0
)

// Game encapsulates the full River Raid simulation state.
type Game struct {
	State            GameState
	ScreenWidth      float32
	ScreenHeight     float32
	CameraY          float32
	TargetCameraY    float32
	ScrollSpeed      float32
	ScreenShake      float32
	World            *ProceduralWorld
	Player           *sprites.PlayerJet
	Bullets          []*sprites.Bullet
	Missiles         []*sprites.Missile
	Particles        *sprites.ParticleSystem
	Audio            *audio.SoundManager
	HUD              *ui.HUD
	Menu             *ui.Menu
	Textures         map[string]rl.Texture2D
	HighScore        int
	ScoreForNextLife int
	GameOverReason   string
	LastCheckpointY  float32
	TotalPlayTime    float32
	RespawnTimer     float32
}

// NewGame constructs and initializes all game subsystems.
func NewGame(width, height int32) *Game {
	w := float32(width)
	h := float32(height)
	seed := time.Now().UnixNano()

	g := &Game{
		State:            StateTitle,
		ScreenWidth:      w,
		ScreenHeight:     h,
		CameraY:          0,
		TargetCameraY:    0,
		ScrollSpeed:      BaseScrollSpeed,
		World:            NewProceduralWorld(w, h, seed),
		Player:           sprites.NewPlayerJet(rl.Vector2{X: w / 2, Y: h * 0.75}),
		Bullets:          make([]*sprites.Bullet, 0, 64),
		Missiles:         make([]*sprites.Missile, 0, 16),
		Particles:        sprites.NewParticleSystem(),
		Audio:            audio.NewSoundManager(),
		HUD:              ui.NewHUD(w, h),
		Menu:             ui.NewMenu(w, h),
		Textures:         make(map[string]rl.Texture2D),
		HighScore:        0,
		ScoreForNextLife: 10000,
		LastCheckpointY:  0,
	}

	g.loadAllTextures()

	return g
}

func (g *Game) loadAllTextures() {
	names := []string{"player", "helicopter", "ship", "destroyer", "enemy_jet", "fuel", "sam_site", "missile", "bridge", "deco_pine", "deco_bush", "deco_house", "deco_building", "deco_rock"}
	for _, name := range names {
		path := "assets/sprites/" + name + ".png"
		if _, err := os.Stat(path); err == nil {
			tex := rl.LoadTexture(path)
			if tex.ID > 0 {
				g.Textures[name] = tex
			}
		}
	}
}

// StartNewGame initializes fresh gameplay run.
func (g *Game) StartNewGame() {
	seed := rand.Int63()
	g.World.Reset(seed)

	g.CameraY = 0
	g.TargetCameraY = 0
	g.ScrollSpeed = BaseScrollSpeed
	g.ScreenShake = 0

	startX := g.ScreenWidth / 2
	startY := g.ScreenHeight * 0.75
	g.Player = sprites.NewPlayerJet(rl.Vector2{X: startX, Y: startY})
	g.Bullets = g.Bullets[:0]
	g.Missiles = g.Missiles[:0]
	g.Particles.Clear()

	g.ScoreForNextLife = 10000
	g.LastCheckpointY = 0
	g.GameOverReason = ""
	g.RespawnTimer = 0
	g.State = StatePlaying

	g.HUD.SetAlert("SORTIE INITIATED - GOOD LUCK PILOT", 2.5, rl.Color{R: 0, G: 240, B: 255, A: 255})
}

// RespawnPlayer puts player back at current camera position after crash.
func (g *Game) RespawnPlayer() {
	playerY := g.CameraY + g.ScreenHeight*0.75
	left, right, hasIsland, islLeft, _ := g.World.GetRiverBoundsAt(playerY)
	playerX := (left + right) / 2
	if hasIsland {
		playerX = (left + islLeft) / 2
	}

	g.Player.ResetForRespawn(rl.Vector2{X: playerX, Y: playerY})
	g.Bullets = g.Bullets[:0]
	g.HUD.SetAlert("AIRCRAFT RE-ENGAGED", 2.0, rl.Color{R: 255, G: 220, B: 50, A: 255})
}

// AddScreenShake adds trauma to camera shake.
func (g *Game) AddScreenShake(amount float32) {
	g.ScreenShake += amount
	if g.ScreenShake > 18.0 {
		g.ScreenShake = 18.0
	}
}

// Close cleans up audio and resources.
func (g *Game) Close() {
	if g.Audio != nil {
		g.Audio.Close()
	}
	for _, tex := range g.Textures {
		rl.UnloadTexture(tex)
	}
}
