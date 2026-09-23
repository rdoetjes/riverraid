package audio

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
)

// GenerateWavData creates a valid WAV file in memory with the given 16-bit PCM samples.
func GenerateWavData(samples []int16, sampleRate int) []byte {
	buf := new(bytes.Buffer)
	numChannels := uint16(1)
	bitsPerSample := uint16(16)
	subchunk2Size := uint32(len(samples) * int(numChannels) * int(bitsPerSample/8))
	chunkSize := uint32(36 + subchunk2Size)
	byteRate := uint32(int(sampleRate) * int(numChannels) * int(bitsPerSample/8))
	blockAlign := uint16(numChannels * (bitsPerSample / 8))

	// RIFF header
	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, chunkSize)
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16)) // Subchunk1Size for PCM
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))  // AudioFormat (PCM)
	_ = binary.Write(buf, binary.LittleEndian, numChannels)
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(buf, binary.LittleEndian, byteRate)
	_ = binary.Write(buf, binary.LittleEndian, blockAlign)
	_ = binary.Write(buf, binary.LittleEndian, bitsPerSample)

	// data chunk
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, subchunk2Size)

	for _, s := range samples {
		_ = binary.Write(buf, binary.LittleEndian, s)
	}

	return buf.Bytes()
}

// SynthShoot generates a laser/cannon blast wave.
func SynthShoot(sampleRate int) []byte {
	duration := 0.16
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := t / duration

		// Pitch drops rapidly from 1200Hz down to 180Hz
		freq := 1200.0 * math.Exp(-progress*4.0)
		phase := freq * t * 2.0 * math.Pi

		// Square wave + sine harmonic
		val := math.Sin(phase)
		if val > 0 {
			val = 0.7
		} else {
			val = -0.7
		}
		val += 0.3 * math.Sin(phase*0.5)

		// Exponential amplitude envelope
		amp := math.Pow(1.0-progress, 1.8)
		samples[i] = int16(val * amp * 26000)
	}

	return GenerateWavData(samples, sampleRate)
}

// SynthExplosion generates an explosive detonation sound with noise burst and low rumble.
func SynthExplosion(sampleRate int, isBig bool) []byte {
	duration := 0.45
	if isBig {
		duration = 0.95
	}
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]int16, numSamples)

	var lastNoise float64
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := t / duration

		// White noise filtered
		noise := (rand.Float64()*2.0 - 1.0)
		// Low-pass smooth
		filterFactor := 0.15 + (1.0-progress)*0.2
		lastNoise = lastNoise + filterFactor*(noise-lastNoise)

		// Sub-bass thump
		bassFreq := 90.0 * (1.0 - progress*0.6)
		bass := math.Sin(bassFreq * t * 2.0 * math.Pi)

		combined := lastNoise*0.75 + bass*0.55
		amp := math.Pow(1.0-progress, 1.5)
		if isBig {
			amp = math.Pow(1.0-progress, 1.1)
		}

		// Clamp
		val := combined * amp * 29000
		if val > 32000 {
			val = 32000
		} else if val < -32000 {
			val = -32000
		}
		samples[i] = int16(val)
	}

	return GenerateWavData(samples, sampleRate)
}

// SynthFuelPickup generates a bright melodic chime sweep for refueling.
func SynthFuelPickup(sampleRate int) []byte {
	duration := 0.12
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := t / duration

		// Arpeggiated high chime
		freq := 600.0 + progress*800.0
		val := 0.6*math.Sin(freq*t*2.0*math.Pi) + 0.4*math.Sin(freq*2.0*t*2.0*math.Pi)
		amp := math.Sin(progress * math.Pi) // Smooth in and out

		samples[i] = int16(val * amp * 22000)
	}

	return GenerateWavData(samples, sampleRate)
}

// SynthLowFuelWarning generates a cautionary two-tone alert beep.
func SynthLowFuelWarning(sampleRate int) []byte {
	duration := 0.18
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := t / duration

		freq := 880.0
		if progress > 0.5 {
			freq = 660.0
		}
		// Pulse square wave
		val := math.Sin(freq * t * 2.0 * math.Pi)
		if val > 0 {
			val = 0.6
		} else {
			val = -0.6
		}
		amp := math.Pow(math.Sin(progress*math.Pi), 0.5)

		samples[i] = int16(val * amp * 20000)
	}

	return GenerateWavData(samples, sampleRate)
}

// SynthExtraLife generates a fanfare chord burst.
func SynthExtraLife(sampleRate int) []byte {
	duration := 0.6
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]int16, numSamples)

	notes := []float64{523.25, 659.25, 783.99, 1046.50} // C5, E5, G5, C6
	noteDuration := duration / float64(len(notes))

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		noteIdx := int(t / noteDuration)
		if noteIdx >= len(notes) {
			noteIdx = len(notes) - 1
		}
		noteProgress := math.Mod(t, noteDuration) / noteDuration
		freq := notes[noteIdx]

		val := 0.7*math.Sin(freq*t*2.0*math.Pi) + 0.3*math.Sin(freq*1.5*t*2.0*math.Pi)
		amp := math.Pow(1.0-noteProgress, 0.8)
		samples[i] = int16(val * amp * 24000)
	}

	return GenerateWavData(samples, sampleRate)
}

// SynthEngineHum generates a looping engine turbine sample.
func SynthEngineHum(sampleRate int) []byte {
	duration := 0.5
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		freq1 := 75.0
		freq2 := 150.0
		freq3 := 225.0

		// Blend bass rumble, turbine whine, and soft air noise
		noise := (rand.Float64()*2.0 - 1.0) * 0.15
		val := 0.5*math.Sin(freq1*t*2.0*math.Pi) +
			0.3*math.Sin(freq2*t*2.0*math.Pi) +
			0.15*math.Sin(freq3*t*2.0*math.Pi) +
			noise

		samples[i] = int16(val * 16000)
	}

	return GenerateWavData(samples, sampleRate)
}

// SynthMissileAlarm generates a piercing high-pitched warning buzzer.
func SynthMissileAlarm(sampleRate int) []byte {
	duration := 0.12
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]int16, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := t / duration

		// High pitched piercing tone
		freq := 1800.0
		// Add a bit of frequency modulation to make it sound like a buzzer
		if int(t*40)%2 == 0 {
			freq = 1500.0
		}

		val := math.Sin(freq * t * 2.0 * math.Pi)
		// Square wave for that retro buzzer feel
		if val > 0 {
			val = 0.5
		} else {
			val = -0.5
		}

		// Fast fade in/out envelope
		amp := math.Sin(progress * math.Pi)
		samples[i] = int16(val * amp * 18000)
	}

	return GenerateWavData(samples, sampleRate)
}
