// CHIP-8 Emulator in Go
// A complete implementation of the CHIP-8 virtual machine
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/chip8-emulator/audio"
	"github.com/chip8-emulator/chip8"
	"github.com/chip8-emulator/display"
	"github.com/chip8-emulator/input"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	// Default emulation speed (instructions per second)
	DefaultClockSpeed = 500
	// Timer frequency (60 Hz as per CHIP-8 spec)
	TimerFrequency = 60
)

func main() {
	// Parse command line arguments
	romPath := flag.String("rom", "", "Path to the CHIP-8 ROM file")
	scale := flag.Int("scale", 10, "Display scale factor")
	speed := flag.Int("speed", DefaultClockSpeed, "Emulation speed (instructions per second)")
	flag.Parse()

	// Check for ROM path
	if *romPath == "" {
		// Check if ROM path is provided as positional argument
		if flag.NArg() > 0 {
			*romPath = flag.Arg(0)
		} else {
			fmt.Println("CHIP-8 Emulator")
			fmt.Println("Usage: chip8-emulator [options] <rom-file>")
			fmt.Println()
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// Load ROM file
	romData, err := os.ReadFile(*romPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading ROM: %v\n", err)
		os.Exit(1)
	}

	// Initialize CHIP-8
	vm := chip8.New()
	if err := vm.LoadROM(romData); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading ROM into memory: %v\n", err)
		os.Exit(1)
	}

	// Initialize display
	disp, err := display.New("CHIP-8 Emulator", int32(*scale))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing display: %v\n", err)
		os.Exit(1)
	}
	defer disp.Close()

	// Initialize audio
	beeper, err := audio.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not initialize audio: %v\n", err)
		// Continue without audio
	} else {
		defer beeper.Close()
	}

	// Initialize keyboard
	keyboard := input.New()

	// Calculate timing intervals
	cycleInterval := time.Second / time.Duration(*speed)
	timerInterval := time.Second / TimerFrequency

	// Main emulation loop
	running := true
	lastCycleTime := time.Now()
	lastTimerTime := time.Now()

	fmt.Printf("Running %s at %d Hz\n", *romPath, *speed)
	fmt.Println("Keys: 1234 QWER ASDF ZXCV (mapped to CHIP-8 keypad)")
	fmt.Println("Press ESC to quit, P to pause/resume, R to reset")

	paused := false

	for running {
		// Handle SDL events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
					case sdl.K_p:
						paused = !paused
						if paused {
							disp.SetTitle("CHIP-8 Emulator (PAUSED)")
						} else {
							disp.SetTitle("CHIP-8 Emulator")
						}
					case sdl.K_r:
						vm.Reset()
						if err := vm.LoadROM(romData); err != nil {
							fmt.Fprintf(os.Stderr, "Error reloading ROM: %v\n", err)
						}
						keyboard.Reset()
					default:
						if key, ok := keyboard.HandleKeyDown(e.Keysym.Sym); ok {
							vm.SetKey(key, true)
						}
					}
				} else if e.Type == sdl.KEYUP {
					if key, ok := keyboard.HandleKeyUp(e.Keysym.Sym); ok {
						vm.SetKey(key, false)
					}
				}
			}
		}

		if paused {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		now := time.Now()

		// Execute CPU cycles
		if now.Sub(lastCycleTime) >= cycleInterval {
			if err := vm.Cycle(); err != nil {
				fmt.Fprintf(os.Stderr, "Emulation error: %v\n", err)
				running = false
			}
			lastCycleTime = now
		}

		// Update timers at 60Hz
		if now.Sub(lastTimerTime) >= timerInterval {
			vm.UpdateTimers()

			// Update beeper
			if beeper != nil {
				beeper.Update(vm.SoundTimer)
			}

			lastTimerTime = now
		}

		// Update display if needed
		if vm.DrawFlag {
			disp.Render(&vm.Display)
			vm.DrawFlag = false
		}

		// Small sleep to prevent CPU spinning
		time.Sleep(time.Microsecond * 100)
	}

	fmt.Println("Emulator stopped.")
}
