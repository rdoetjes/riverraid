package audio

import (
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SoundID identifies distinct game audio events.
type SoundID int

const (
	SoundShoot SoundID = iota
	SoundExplosion
	SoundBigExplosion
	SoundFuelPickup
	SoundLowFuel
	SoundExtraLife
	SoundEngine
	SoundMissileWarning
	SoundCount
)

// SoundManager handles loading, procedural fallback, and playback of game audio.
type SoundManager struct {
	sounds      map[SoundID]rl.Sound
	initialized bool
	muted       bool
	lastFuelSfx float64
}

// NewSoundManager initializes RayLib audio device and loads or synthesizes sound effects.
func NewSoundManager() *SoundManager {
	sm := &SoundManager{
		sounds: make(map[SoundID]rl.Sound),
	}

	if !rl.IsAudioDeviceReady() {
		rl.InitAudioDevice()
	}
	sm.initialized = rl.IsAudioDeviceReady()

	if sm.initialized {
		sm.loadAllSounds()
	}

	return sm
}

// loadSound attempts to load from custom asset path; if not found, falls back to synthesized wave.
func (sm *SoundManager) loadSound(id SoundID, filePath string, synthFunc func(int) []byte) {
	if _, err := os.Stat(filePath); err == nil {
		// Custom user file exists
		sound := rl.LoadSound(filePath)
		if sound.FrameCount > 0 {
			sm.sounds[id] = sound
			return
		}
	}

	// Fallback to procedurally generated sound wave
	sampleRate := 44100
	wavData := synthFunc(sampleRate)
	wave := rl.LoadWaveFromMemory(".wav", wavData, int32(len(wavData)))
	if wave.FrameCount > 0 {
		sound := rl.LoadSoundFromWave(wave)
		rl.UnloadWave(wave)
		if sound.FrameCount > 0 {
			sm.sounds[id] = sound
		}
	}
}

func (sm *SoundManager) loadAllSounds() {
	sm.loadSound(SoundShoot, "assets/sounds/shoot.wav", func(sr int) []byte {
		return SynthShoot(sr)
	})
	sm.loadSound(SoundExplosion, "assets/sounds/explosion.wav", func(sr int) []byte {
		return SynthExplosion(sr, false)
	})
	sm.loadSound(SoundBigExplosion, "assets/sounds/big_explosion.wav", func(sr int) []byte {
		return SynthExplosion(sr, true)
	})
	sm.loadSound(SoundFuelPickup, "assets/sounds/fuel.wav", func(sr int) []byte {
		return SynthFuelPickup(sr)
	})
	sm.loadSound(SoundLowFuel, "assets/sounds/low_fuel.wav", func(sr int) []byte {
		return SynthLowFuelWarning(sr)
	})
	sm.loadSound(SoundExtraLife, "assets/sounds/extra_life.wav", func(sr int) []byte {
		return SynthExtraLife(sr)
	})
	sm.loadSound(SoundEngine, "assets/sounds/engine.wav", func(sr int) []byte {
		return SynthEngineHum(sr)
	})
	sm.loadSound(SoundMissileWarning, "assets/sounds/warning.wav", func(sr int) []byte {
		return SynthMissileAlarm(sr)
	})
}

// Play triggers a one-shot sound effect.
func (sm *SoundManager) Play(id SoundID) {
	if sm.muted || !sm.initialized {
		return
	}
	if sound, ok := sm.sounds[id]; ok && sound.FrameCount > 0 {
		rl.PlaySound(sound)
	}
}

// PlayWithPitch plays a sound with custom pitch variation.
func (sm *SoundManager) PlayWithPitch(id SoundID, pitch float32) {
	if sm.muted || !sm.initialized {
		return
	}
	if sound, ok := sm.sounds[id]; ok && sound.FrameCount > 0 {
		rl.SetSoundPitch(sound, pitch)
		rl.PlaySound(sound)
	}
}

// PlayFuelRefuel plays throttled refueling chimes.
func (sm *SoundManager) PlayFuelRefuel(currentTime float64) {
	if currentTime-sm.lastFuelSfx > 0.08 {
		sm.lastFuelSfx = currentTime
		sm.Play(SoundFuelPickup)
	}
}

// ToggleMute switches audio muting state.
func (sm *SoundManager) ToggleMute() {
	sm.muted = !sm.muted
}

// IsMuted returns true if sounds are muted.
func (sm *SoundManager) IsMuted() bool {
	return sm.muted
}

// Close unloads all loaded sounds and closes the RayLib audio device.
func (sm *SoundManager) Close() {
	if !sm.initialized {
		return
	}
	for id, sound := range sm.sounds {
		if sound.FrameCount > 0 {
			rl.UnloadSound(sound)
		}
		delete(sm.sounds, id)
	}
	rl.CloseAudioDevice()
	sm.initialized = false
}

// PrintInfo prints active audio status.
func (sm *SoundManager) PrintInfo() {
	fmt.Printf("[Audio] Initialized: %v, Loaded SFX: %d\n", sm.initialized, len(sm.sounds))
}
