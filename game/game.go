package game

import (
	"encoding/json"
	"math/rand"
	"os"
	"sort"
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
	StateEnteringName
)

// HighScoreEntry represents a single record in the top 10.
type HighScoreEntry struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

const (
	DefaultScreenWidth  = 1024
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
	HighScores       []HighScoreEntry
	ScoreForNextLife int
	GameOverReason   string
	LastCheckpointY  float32
	TotalPlayTime    float32
	RespawnTimer     float32
	EnteringName     bool
	EnterNameBuffer  string
	ScoreSubmitted   bool
	WaterShader      rl.Shader
	TimeLoc          int32
	ResLoc           int32
	MainFont         rl.Font
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
		Textures:         make(map[string]rl.Texture2D),
		HighScore:        0,
		ScoreForNextLife: 10000,
		LastCheckpointY:  0,
	}

	g.loadAllTextures()
	g.loadFonts()
	g.loadShaders()
	g.LoadHighScores()

	// Initialize HUD and Menu with the font
	g.HUD = ui.NewHUD(w, h, g.MainFont)
	g.Menu = ui.NewMenu(w, h, g.MainFont)

	return g
}

func (g *Game) LoadHighScores() {
	// Initialize defaults
	g.HighScores = make([]HighScoreEntry, 10)
	for i := 0; i < 10; i++ {
		g.HighScores[i] = HighScoreEntry{Name: "ACE", Score: (10 - i) * 1000}
	}
	g.HighScore = g.HighScores[0].Score

	// Attempt to load from file
	data, err := os.ReadFile("highscores.json")
	if err == nil {
		var loaded []HighScoreEntry
		if json.Unmarshal(data, &loaded) == nil && len(loaded) > 0 {
			g.HighScores = loaded
			g.HighScore = g.HighScores[0].Score
		}
	}
}

func (g *Game) SaveHighScores() {
	data, err := json.Marshal(g.HighScores)
	if err == nil {
		os.WriteFile("highscores.json", data, 0644)
	}
}

func (g *Game) IsNewHighScore(score int) bool {
	return score > g.HighScores[len(g.HighScores)-1].Score
}

func (g *Game) AddHighScore(name string, score int) {
	newEntry := HighScoreEntry{Name: name, Score: score}
	g.HighScores = append(g.HighScores, newEntry)

	// Sort high to low
	sort.Slice(g.HighScores, func(i, j int) bool {
		return g.HighScores[i].Score > g.HighScores[j].Score
	})

	// Keep top 10
	if len(g.HighScores) > 10 {
		g.HighScores = g.HighScores[:10]
	}
	g.HighScore = g.HighScores[0].Score
	g.SaveHighScores()
}

func (g *Game) loadAllTextures() {
	names := []string{"logo", "player", "helicopter", "ship", "destroyer", "enemy_jet", "fuel", "sam_site", "missile", "bridge", "deco_pine", "deco_bush", "deco_house", "deco_building", "deco_rock"}
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
	g.RespawnTimer = 0.1
	g.State = StateDying
	g.Player.Active = false
	g.Player.Lives = 3
	g.ScoreSubmitted = false

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

func (g *Game) loadShaders() {
	if _, err := os.Stat("assets/shaders/water.frag"); err == nil {
		g.WaterShader = rl.LoadShader("", "assets/shaders/water.frag")
		if g.WaterShader.ID > 0 {
			g.TimeLoc = rl.GetShaderLocation(g.WaterShader, "time")
			g.ResLoc = rl.GetShaderLocation(g.WaterShader, "resolution")

			res := []float32{g.ScreenWidth, g.ScreenHeight}
			rl.SetShaderValue(g.WaterShader, g.ResLoc, res, rl.ShaderUniformVec2)
		}
	}
}

func (g *Game) loadFonts() {
	path := "assets/fonts/Army.ttf"
	if _, err := os.Stat(path); err == nil {
		// Load with a larger base size for better TTF quality
		g.MainFont = rl.LoadFontEx(path, 48, nil, 0)
		rl.SetTextureFilter(g.MainFont.Texture, rl.FilterBilinear)
	} else {
		g.MainFont = rl.GetFontDefault()
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
	if g.WaterShader.ID > 0 {
		rl.UnloadShader(g.WaterShader)
	}
	rl.UnloadFont(g.MainFont)
}
