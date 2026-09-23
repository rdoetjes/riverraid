package game

import (
	"math"
	"math/rand"

	"riverraid/sprites"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SliceStep      = 20.0   // Vertical distance between river mesh sample points
	SectionLength  = 3600.0 // Distance between consecutive bridges
	RiverMinWidth  = 80.0
	RiverMaxWidth  = 440.0
	IslandMinRiver = 340.0 // River must be at least this wide for an island to spawn
)

// RiverSlice represents a cross-section of the river at a specific world Y coordinate.
type RiverSlice struct {
	WorldY         float32
	LeftBankX      float32
	RightBankX     float32
	HasIsland      bool
	IslandLeftX    float32
	IslandRightX   float32
	RiverCenter    float32
	RiverHalfWidth float32
}

// ProceduralWorld manages dynamic river generation, island meshes, scenery, and enemy spawning.
type ProceduralWorld struct {
	Seed           int64
	ScreenWidth    float32
	ScreenHeight   float32
	ActiveSlices   []RiverSlice
	Decorations    []*sprites.TerrainDecoration
	Enemies        []sprites.Sprite
	Bridges        []*sprites.Bridge
	HighestGenY    float32 // Furthest forward Y generated (most negative)
	LowestGenY     float32 // Furthest back Y kept in memory
	NextBridgeY    float32
	CurrentSection int
}

// NewProceduralWorld initializes world generation starting from world origin.
func NewProceduralWorld(screenWidth, screenHeight float32, seed int64) *ProceduralWorld {
	pw := &ProceduralWorld{
		Seed:           seed,
		ScreenWidth:    screenWidth,
		ScreenHeight:   screenHeight,
		ActiveSlices:   make([]RiverSlice, 0, 512),
		Decorations:    make([]*sprites.TerrainDecoration, 0, 256),
		Enemies:        make([]sprites.Sprite, 0, 128),
		Bridges:        make([]*sprites.Bridge, 0, 16),
		HighestGenY:    1000.0,
		LowestGenY:     1000.0,
		NextBridgeY:    -SectionLength,
		CurrentSection: 1,
	}

	// Pre-generate initial river chunks from Y = 1000 down to -screenHeight*2.5
	pw.GenerateAhead(-screenHeight * 2.5)

	return pw
}

// sampleRiver calculates river boundaries analytically at any world Y coordinate.
func (pw *ProceduralWorld) sampleRiver(worldY float32) RiverSlice {
	// Multi-octave sinusoidal procedural function for smooth river meander
	t := float64(-worldY) * 0.0018
	seedOffset := float64(pw.Seed%1000) * 0.1

	// Center meander
	c1 := math.Sin(t*1.1 + seedOffset)
	c2 := math.Sin(t*2.7+seedOffset*1.4) * 0.45
	c3 := math.Sin(t*5.3+seedOffset*2.1) * 0.18
	normCenter := (c1 + c2 + c3) / 1.63 // -1.0 to 1.0

	// Width oscillation
	w1 := math.Sin(t*1.8 + seedOffset*0.8)
	w2 := math.Cos(t*3.9+seedOffset*1.7) * 0.5
	normWidth := (w1 + w2) / 1.5 // -1.0 to 1.0
	if normWidth < 0 {
		normWidth = -normWidth
	}

	// Near bridge locations, straighten and widen river cleanly (analytical bridge position)
	// Safety: No bridge straightening for the starting runway (worldY > 0)
	distToBridge := float32(1000.0)
	if worldY < -500.0 {
		nearestBridgeIdx := math.Round(float64(-worldY) / SectionLength)
		if nearestBridgeIdx < 1 {
			nearestBridgeIdx = 1
		}
		nearestBridgeY := -float32(nearestBridgeIdx) * SectionLength
		distToBridge = float32(math.Abs(float64(worldY - nearestBridgeY)))
	}

	bridgeFactor := float32(1.0)
	if distToBridge < 220.0 {
		blend := distToBridge / 220.0
		normCenter *= float64(blend)
		bridgeFactor = 1.0 + (1.0-blend)*0.3
	}

	centerX := pw.ScreenWidth/2.0 + float32(normCenter)*(pw.ScreenWidth*0.22)
	halfWidth := (RiverMinWidth + float32(normWidth)*(RiverMaxWidth-RiverMinWidth)*0.85) * bridgeFactor

	// Starting safety: Ensure river is wide and centered for the first 1200 pixels
	if worldY > -1200.0 {
		safeBlend := float32(1.0)
		if worldY < 0 {
			safeBlend = (worldY + 1200.0) / 1200.0 // Fade out safety over 1200px
		}
		// Force wide and centered
		centerX = (pw.ScreenWidth/2.0)*(1.0-safeBlend) + (pw.ScreenWidth/2.0)*safeBlend
		halfWidth = halfWidth*(1.0-safeBlend) + (RiverMaxWidth*0.85)*safeBlend
	}

	leftBank := centerX - halfWidth
	rightBank := centerX + halfWidth

	// Clamp within screen boundaries with margin for banks
	minBankMargin := float32(40.0)
	if leftBank < minBankMargin {
		leftBank = minBankMargin
	}
	if rightBank > pw.ScreenWidth-minBankMargin {
		rightBank = pw.ScreenWidth - minBankMargin
	}

	slice := RiverSlice{
		WorldY:         worldY,
		LeftBankX:      leftBank,
		RightBankX:     rightBank,
		RiverCenter:    (leftBank + rightBank) / 2.0,
		RiverHalfWidth: (rightBank - leftBank) / 2.0,
		HasIsland:      false,
	}

	// Central island generation: Only spawn after Zone 1 and not right at a bridge
	islandFreq := math.Sin(t*3.4 + seedOffset*3.1)
	if worldY < -SectionLength && (rightBank-leftBank) > IslandMinRiver && islandFreq > 0.35 && distToBridge > 250.0 {
		islandWidth := float32((islandFreq - 0.35) * 1.5 * 80.0)
		if islandWidth > (rightBank-leftBank)*0.35 {
			islandWidth = (rightBank - leftBank) * 0.35
		}
		islandMid := slice.RiverCenter + float32(math.Sin(t*6.0))*15.0

		slice.HasIsland = true
		slice.IslandLeftX = islandMid - islandWidth/2.0
		slice.IslandRightX = islandMid + islandWidth/2.0
	}

	return slice
}

// GenerateAhead creates new river slices, bridge checkpoints, decorations, and enemies ahead.
func (pw *ProceduralWorld) GenerateAhead(targetY float32) {
	for pw.HighestGenY > targetY {
		currY := pw.HighestGenY - SliceStep
		slice := pw.sampleRiver(currY)
		pw.ActiveSlices = append(pw.ActiveSlices, slice)
		pw.HighestGenY = currY

		// Check if we reached a bridge location
		if currY <= pw.NextBridgeY && currY > pw.NextBridgeY-SliceStep {
			bridge := sprites.NewBridge(pw.NextBridgeY, slice.LeftBankX, slice.RightBankX, pw.CurrentSection)
			pw.Bridges = append(pw.Bridges, bridge)
			pw.CurrentSection++
			pw.NextBridgeY -= SectionLength
		}

		// Procedural entity spawning
		pw.spawnSliceEntities(slice)
	}
}

// GetSectionAt returns the logical section index for a given world Y coordinate.
func (pw *ProceduralWorld) GetSectionAt(worldY float32) int {
	if worldY > 0 {
		return 1
	}
	return int(math.Abs(float64(worldY))/SectionLength) + 1
}

// spawnSliceEntities populates the river slice with enemies, fuel depots, and bank scenery.
func (pw *ProceduralWorld) spawnSliceEntities(slice RiverSlice) {
	rng := rand.New(rand.NewSource(int64(math.Abs(float64(slice.WorldY)))*1000 + pw.Seed))
	section := pw.GetSectionAt(slice.WorldY)

	// 1. Terrain Decorations on Left and Right Embankments
	if rng.Float64() < 0.45 {
		// Left bank decoration
		decoX := rng.Float32() * (slice.LeftBankX - 15.0)
		if decoX > 10 {
			decoType := sprites.DecoPineTree
			roll := rng.Float64()
			if roll < 0.25 {
				decoType = sprites.DecoPineTree
			} else if roll < 0.45 {
				decoType = sprites.DecoDeciduousTree
			} else if roll < 0.65 {
				decoType = sprites.DecoBush
			} else if roll < 0.75 {
				decoType = sprites.DecoHouse
			} else if roll < 0.82 {
				decoType = sprites.DecoBuilding
			} else if roll < 0.92 {
				decoType = sprites.DecoRock
			} else if roll < 0.95 {
				decoType = sprites.DecoRadarStation
			} else if roll < 0.97 && section >= 4 {
				// SAM Site spawn on left bank - from Level 4 onwards (~2 per section per bank)
				pw.Enemies = append(pw.Enemies, sprites.NewSAMSite(rl.Vector2{X: decoX, Y: slice.WorldY}))
				return
			} else {
				decoType = sprites.DecoBunker
			}
			scale := 1.0 + rng.Float32()*1.2
			pw.Decorations = append(pw.Decorations, sprites.NewTerrainDecoration(rl.Vector2{X: decoX, Y: slice.WorldY}, decoType, scale))
		}
	}

	if rng.Float64() < 0.45 {
		// Right bank decoration
		margin := pw.ScreenWidth - slice.RightBankX
		if margin > 20 {
			decoX := slice.RightBankX + 15.0 + rng.Float32()*(margin-25.0)
			decoType := sprites.DecoPineTree
			roll := rng.Float64()
			if roll < 0.25 {
				decoType = sprites.DecoPineTree
			} else if roll < 0.45 {
				decoType = sprites.DecoDeciduousTree
			} else if roll < 0.65 {
				decoType = sprites.DecoBush
			} else if roll < 0.75 {
				decoType = sprites.DecoHouse
			} else if roll < 0.82 {
				decoType = sprites.DecoBuilding
			} else if roll < 0.92 {
				decoType = sprites.DecoRock
			} else if roll < 0.95 {
				decoType = sprites.DecoRadarStation
			} else if roll < 0.97 && section >= 4 {
				// SAM Site spawn on right bank - from Level 4 onwards (~2 per section per bank)
				pw.Enemies = append(pw.Enemies, sprites.NewSAMSite(rl.Vector2{X: decoX, Y: slice.WorldY}))
				return
			} else {
				decoType = sprites.DecoBunker
			}
			scale := 1.0 + rng.Float32()*1.5
			pw.Decorations = append(pw.Decorations, sprites.NewTerrainDecoration(rl.Vector2{X: decoX, Y: slice.WorldY}, decoType, scale))
		}
	}

	// 2. Island Decorations
	if slice.HasIsland && rng.Float64() < 0.55 {
		islandWidth := slice.IslandRightX - slice.IslandLeftX
		if islandWidth > 20 {
			decoX := slice.IslandLeftX + 5.0 + rng.Float32()*(islandWidth-10.0)
			decoType := sprites.DecoPineTree
			if rng.Float64() < 0.5 {
				decoType = sprites.DecoBush
			}
			scale := 1.0 + rng.Float32()*1.5
			pw.Decorations = append(pw.Decorations, sprites.NewTerrainDecoration(rl.Vector2{X: decoX, Y: slice.WorldY}, decoType, scale))
		}
	}

	// 3. Spawning Enemies and Fuel Stations (spaced periodically)
	// Don't spawn right on top of a bridge
	distToBridge := float32(math.Abs(float64(slice.WorldY - pw.NextBridgeY)))
	if distToBridge < 180.0 {
		return
	}

	// Spawn density check: Higher density (every 110 pixels instead of 160)
	yInt := int(math.Abs(float64(slice.WorldY)))
	if yInt%110 < int(SliceStep) {
		spawnRoll := rng.Float64()

		if spawnRoll < 0.22 {
			// Fuel Depot spawn
			var fuelX float32
			if slice.HasIsland {
				// Put in left or right channel
				if rng.Float64() < 0.5 {
					fuelX = (slice.LeftBankX + slice.IslandLeftX) / 2.0
				} else {
					fuelX = (slice.IslandRightX + slice.RightBankX) / 2.0
				}
			} else {
				fuelX = slice.RiverCenter + (rng.Float32()-0.5)*(slice.RiverHalfWidth*0.7)
			}
			pw.Enemies = append(pw.Enemies, sprites.NewFuelDepot(rl.Vector2{X: fuelX, Y: slice.WorldY}))

		} else if spawnRoll < 0.45 {
			// Helicopter spawn
			minX := slice.LeftBankX + 20
			maxX := slice.RightBankX - 20
			if slice.HasIsland {
				if rng.Float64() < 0.5 {
					minX = slice.LeftBankX + 15
					maxX = slice.IslandLeftX - 15
				} else {
					minX = slice.IslandRightX + 15
					maxX = slice.RightBankX - 15
				}
			}
			if maxX > minX+30 {
				posX := minX + rng.Float32()*(maxX-minX)
				speed := float32(40.0 + rng.Float64()*45.0 + float64(section)*4.0)
				pw.Enemies = append(pw.Enemies, sprites.NewHelicopter(rl.Vector2{X: posX, Y: slice.WorldY}, minX, maxX, speed))
			}

		} else if spawnRoll < 0.70 {
			// Ship / Destroyer slot
			minX := slice.LeftBankX + 25
			maxX := slice.RightBankX - 25
			if slice.HasIsland {
				if rng.Float64() < 0.5 {
					minX, maxX = slice.LeftBankX+22, slice.IslandLeftX-22
				} else {
					minX, maxX = slice.IslandRightX+22, slice.RightBankX-22
				}
			}

			// Frequency Control: Targeting ~3 destroyers per 3600-pixel section
			// 3600 / 110 (spawn window) = ~32 windows. 3 / 32 = ~9% total probability.
			if section >= 3 && spawnRoll < 0.54 { // 0.54 - 0.45 = 0.09 (9%)
				if maxX > minX+30 {
					posX := minX + rng.Float32()*(maxX-minX)
					speed := float32(60.0 + float64(section)*4.0)
					pw.Enemies = append(pw.Enemies, sprites.NewDestroyer(rl.Vector2{X: posX, Y: slice.WorldY}, minX, maxX, speed))
				}
			} else {
				// Standard Ship
				if maxX > minX+35 {
					posX := minX + rng.Float32()*(maxX-minX)
					speed := float32(25.0 + rng.Float64()*30.0)
					pw.Enemies = append(pw.Enemies, sprites.NewShip(rl.Vector2{X: posX, Y: slice.WorldY}, minX, maxX, speed))
				}
			}

		} else if spawnRoll < 0.95 {
			// Fast Interceptor Jet
			minX := slice.LeftBankX + 15
			maxX := slice.RightBankX - 15
			if maxX > minX+50 {
				posX := minX + rng.Float32()*(maxX-minX)
				speed := float32(110.0 + rng.Float64()*60.0 + float64(section)*8.0)
				pw.Enemies = append(pw.Enemies, sprites.NewEnemyJet(rl.Vector2{X: posX, Y: slice.WorldY}, minX, maxX, speed))
			}
		} else {
			// Extra filler slot (Decoration or Jet)
			if rng.Float64() < 0.5 {
				minX := slice.LeftBankX + 15
				maxX := slice.RightBankX - 15
				if maxX > minX+50 {
					posX := minX + rng.Float32()*(maxX-minX)
					speed := float32(110.0 + rng.Float64()*60.0 + float64(section)*8.0)
					pw.Enemies = append(pw.Enemies, sprites.NewEnemyJet(rl.Vector2{X: posX, Y: slice.WorldY}, minX, maxX, speed))
				}
			}
		}
	}
}

// CleanupBehind removes world slices and objects that have scrolled far behind the camera.
func (pw *ProceduralWorld) CleanupBehind(camBottomY float32) {
	threshold := camBottomY + 300.0

	// Slices cleanup
	sliceIdx := 0
	for sliceIdx < len(pw.ActiveSlices) && pw.ActiveSlices[sliceIdx].WorldY > threshold {
		sliceIdx++
	}
	if sliceIdx > 0 {
		pw.ActiveSlices = pw.ActiveSlices[sliceIdx:]
	}

	// Decorations cleanup
	decoCount := 0
	for i := 0; i < len(pw.Decorations); i++ {
		d := pw.Decorations[i]
		if d.Position.Y <= threshold {
			pw.Decorations[decoCount] = d
			decoCount++
		}
	}
	pw.Decorations = pw.Decorations[:decoCount]

	// Enemies cleanup
	enemyCount := 0
	for i := 0; i < len(pw.Enemies); i++ {
		e := pw.Enemies[i]
		if e.IsActive() && e.GetPosition().Y <= threshold {
			pw.Enemies[enemyCount] = e
			enemyCount++
		}
	}
	pw.Enemies = pw.Enemies[:enemyCount]

	// Bridges cleanup
	bridgeCount := 0
	for i := 0; i < len(pw.Bridges); i++ {
		b := pw.Bridges[i]
		if b.Position.Y <= threshold {
			pw.Bridges[bridgeCount] = b
			bridgeCount++
		}
	}
	pw.Bridges = pw.Bridges[:bridgeCount]
}

// GetSliceAt finds the interpolated river bounds for any Y position.
func (pw *ProceduralWorld) GetSliceAt(worldY float32) RiverSlice {
	return pw.sampleRiver(worldY)
}

// GetRiverBoundsAt linearly interpolates exact river boundaries from active slices to match screen rendering.
func (pw *ProceduralWorld) GetRiverBoundsAt(worldY float32) (leftBank, rightBank float32, hasIsland bool, islLeft, islRight float32) {
	slices := pw.ActiveSlices
	n := len(slices)
	if n >= 2 {
		for i := 0; i < n-1; i++ {
			s0 := slices[i]
			s1 := slices[i+1]
			// Slices are ordered from higher Y to lower Y
			if s0.WorldY >= worldY && worldY >= s1.WorldY {
				denom := s0.WorldY - s1.WorldY
				t := float32(0.0)
				if denom > 0.001 {
					t = (s0.WorldY - worldY) / denom
				}
				leftBank = s0.LeftBankX + t*(s1.LeftBankX-s0.LeftBankX)
				rightBank = s0.RightBankX + t*(s1.RightBankX-s0.RightBankX)

				// Island linear interpolation matching display.go exactly
				if s0.HasIsland || s1.HasIsland {
					il0, ir0 := s0.IslandLeftX, s0.IslandRightX
					if !s0.HasIsland {
						il0, ir0 = s0.RiverCenter, s0.RiverCenter
					}
					il1, ir1 := s1.IslandLeftX, s1.IslandRightX
					if !s1.HasIsland {
						il1, ir1 = s1.RiverCenter, s1.RiverCenter
					}
					islLeft = il0 + t*(il1-il0)
					islRight = ir0 + t*(ir1-ir0)
					if islRight-islLeft > 2.0 {
						hasIsland = true
					}
				}
				return leftBank, rightBank, hasIsland, islLeft, islRight
			}
		}
	}

	// Fallback to analytical sampling if coordinates are outside cached active slices
	s := pw.sampleRiver(worldY)
	return s.LeftBankX, s.RightBankX, s.HasIsland, s.IslandLeftX, s.IslandRightX
}

// IsPointInWater checks if (x, y) is inside the navigable river water (and not on land or an island).
func (pw *ProceduralWorld) IsPointInWater(x, y float32) bool {
	leftBank, rightBank, hasIsland, islLeft, islRight := pw.GetRiverBoundsAt(y)

	// 2px safety tolerance to prevent frustrating edge clipping
	tolerance := float32(2.0)

	// Outside main river banks -> Land
	if x <= leftBank+tolerance || x >= rightBank-tolerance {
		return false
	}

	// Inside central island -> Land
	if hasIsland && x >= islLeft-tolerance && x <= islRight+tolerance {
		return false
	}

	return true
}

// Reset clears world progress and restarts generation at section 1.
func (pw *ProceduralWorld) Reset(seed int64) {
	pw.Seed = seed
	pw.ActiveSlices = pw.ActiveSlices[:0]
	pw.Decorations = pw.Decorations[:0]
	pw.Enemies = pw.Enemies[:0]
	pw.Bridges = pw.Bridges[:0]
	pw.HighestGenY = 1000.0
	pw.LowestGenY = 1000.0
	pw.NextBridgeY = -SectionLength
	pw.CurrentSection = 1

	pw.GenerateAhead(-pw.ScreenHeight * 2.5)
}
