// Package input handles keyboard input mapping for the CHIP-8 emulator
package input

import "github.com/veandco/go-sdl2/sdl"

/*
CHIP-8 Keypad Layout:    Keyboard Mapping:
+---+---+---+---+        +---+---+---+---+
| 1 | 2 | 3 | C |        | 1 | 2 | 3 | 4 |
+---+---+---+---+        +---+---+---+---+
| 4 | 5 | 6 | D |   =>   | Q | W | E | R |
+---+---+---+---+        +---+---+---+---+
| 7 | 8 | 9 | E |        | A | S | D | F |
+---+---+---+---+        +---+---+---+---+
| A | 0 | B | F |        | Z | X | C | V |
+---+---+---+---+        +---+---+---+---+
*/

// KeyMap maps SDL keycodes to CHIP-8 key indices (0x0-0xF)
var KeyMap = map[sdl.Keycode]uint8{
	sdl.K_1: 0x1, sdl.K_2: 0x2, sdl.K_3: 0x3, sdl.K_4: 0xC,
	sdl.K_q: 0x4, sdl.K_w: 0x5, sdl.K_e: 0x6, sdl.K_r: 0xD,
	sdl.K_a: 0x7, sdl.K_s: 0x8, sdl.K_d: 0x9, sdl.K_f: 0xE,
	sdl.K_z: 0xA, sdl.K_x: 0x0, sdl.K_c: 0xB, sdl.K_v: 0xF,
}

// Keyboard handles keyboard input state
type Keyboard struct {
	// Keys tracks the current state of each CHIP-8 key
	Keys [16]bool
}

// New creates a new Keyboard instance
func New() *Keyboard {
	return &Keyboard{}
}

// HandleKeyDown processes a key down event
func (k *Keyboard) HandleKeyDown(keycode sdl.Keycode) (uint8, bool) {
	if chip8Key, ok := KeyMap[keycode]; ok {
		k.Keys[chip8Key] = true
		return chip8Key, true
	}
	return 0, false
}

// HandleKeyUp processes a key up event
func (k *Keyboard) HandleKeyUp(keycode sdl.Keycode) (uint8, bool) {
	if chip8Key, ok := KeyMap[keycode]; ok {
		k.Keys[chip8Key] = false
		return chip8Key, true
	}
	return 0, false
}

// IsKeyPressed returns true if the specified CHIP-8 key is currently pressed
func (k *Keyboard) IsKeyPressed(key uint8) bool {
	if key < 16 {
		return k.Keys[key]
	}
	return false
}

// Reset resets all key states to unpressed
func (k *Keyboard) Reset() {
	for i := range k.Keys {
		k.Keys[i] = false
	}
}

// GetKeyState returns a copy of the current key state array
func (k *Keyboard) GetKeyState() [16]bool {
	return k.Keys
}
