package sprites

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// DecorationType represents different landscape props.
type DecorationType int

const (
	DecoPineTree DecorationType = iota
	DecoDeciduousTree
	DecoRock
	DecoRadarStation
	DecoBunker
	DecoHouse
	DecoBuilding
	DecoBush
)

// TerrainDecoration represents an environmental vector prop on the riverbanks.
type TerrainDecoration struct {
	BaseSprite
	DecoType DecorationType
	Scale    float32
	Rotation float32
}

// NewTerrainDecoration creates an environmental decoration.
func NewTerrainDecoration(pos rl.Vector2, decoType DecorationType, scale float32, rotation float32) *TerrainDecoration {
	return &TerrainDecoration{
		BaseSprite: BaseSprite{
			Position: pos,
			Active:   true,
		},
		DecoType: decoType,
		Scale:    scale,
		Rotation: rotation,
	}
}

func (td *TerrainDecoration) GetType() SpriteType {
	return TypeDecoration
}

func (td *TerrainDecoration) Update(dt float32) {
	td.Age += dt
}

func (td *TerrainDecoration) Draw(tex rl.Texture2D) {
	if !td.Active {
		return
	}

	destRec := rl.Rectangle{X: td.Position.X, Y: td.Position.Y, Width: td.Scale * 32, Height: td.Scale * 32}
	origin := rl.Vector2{X: destRec.Width / 2, Y: destRec.Height / 2}
	rl.DrawTexturePro(tex, rl.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)}, destRec, origin, td.Rotation, rl.White)
}
