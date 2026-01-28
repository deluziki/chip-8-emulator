// Package audio handles sound output for the CHIP-8 emulator
package audio

import (
	"math"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	// Audio configuration
	SampleRate = 44100
	Frequency  = 440 // A4 note
	Amplitude  = 0.3 // Volume (0.0 - 1.0)
)

// Beeper handles audio playback for the CHIP-8 sound timer
type Beeper struct {
	deviceID  sdl.AudioDeviceID
	isPlaying bool
	phase     float64
	mu        sync.Mutex
}

// audioCallback is called by SDL when it needs more audio data
func (b *Beeper) audioCallback(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.isPlaying {
		// Fill with silence
		for i := range data {
			data[i] = 0
		}
		return
	}

	// Generate square wave
	phaseIncrement := 2 * math.Pi * Frequency / SampleRate

	for i := 0; i < len(data); i += 2 {
		var sample int16
		if math.Sin(b.phase) >= 0 {
			sample = int16(Amplitude * 32767)
		} else {
			sample = int16(-Amplitude * 32767)
		}

		// Write 16-bit sample (little-endian)
		data[i] = byte(sample)
		data[i+1] = byte(sample >> 8)

		b.phase += phaseIncrement
		if b.phase >= 2*math.Pi {
			b.phase -= 2 * math.Pi
		}
	}
}

// New creates a new Beeper instance
func New() (*Beeper, error) {
	b := &Beeper{}

	spec := &sdl.AudioSpec{
		Freq:     SampleRate,
		Format:   sdl.AUDIO_S16LSB,
		Channels: 1,
		Samples:  512,
		Callback: sdl.AudioCallback(b.audioCallbackWrapper),
	}

	var obtainedSpec sdl.AudioSpec
	deviceID, err := sdl.OpenAudioDevice("", false, spec, &obtainedSpec, 0)
	if err != nil {
		return nil, err
	}

	b.deviceID = deviceID

	// Start audio (paused initially)
	sdl.PauseAudioDevice(b.deviceID, false)

	return b, nil
}

// audioCallbackWrapper wraps the callback for SDL
func (b *Beeper) audioCallbackWrapper(userdata interface{}, stream []byte) {
	b.audioCallback(stream)
}

// Play starts the beep sound
func (b *Beeper) Play() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isPlaying = true
}

// Stop stops the beep sound
func (b *Beeper) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isPlaying = false
}

// IsPlaying returns whether the beeper is currently playing
func (b *Beeper) IsPlaying() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.isPlaying
}

// Close cleans up audio resources
func (b *Beeper) Close() {
	b.Stop()
	if b.deviceID != 0 {
		sdl.CloseAudioDevice(b.deviceID)
	}
}

// Update updates the beeper state based on the sound timer
func (b *Beeper) Update(soundTimer uint8) {
	if soundTimer > 0 {
		if !b.IsPlaying() {
			b.Play()
		}
	} else {
		if b.IsPlaying() {
			b.Stop()
		}
	}
}
