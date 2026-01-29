// Package chip8 implements the CHIP-8 virtual machine
package chip8

import (
	"fmt"
	"math/rand"
)

const (
	// Memory size (4KB)
	MemorySize = 4096
	// Number of general purpose registers
	NumRegisters = 16
	// Stack size (16 levels)
	StackSize = 16
	// Display width in pixels
	DisplayWidth = 64
	// Display height in pixels
	DisplayHeight = 32
	// Number of keys on the keypad
	NumKeys = 16
	// Program start address (programs are loaded at 0x200)
	ProgramStart = 0x200
)

// CHIP8 represents the CHIP-8 virtual machine
type CHIP8 struct {
	// Memory (4KB)
	Memory [MemorySize]uint8

	// General purpose registers V0-VF
	V [NumRegisters]uint8

	// Index register
	I uint16

	// Program counter
	PC uint16

	// Stack
	Stack [StackSize]uint16

	// Stack pointer
	SP uint8

	// Delay timer
	DelayTimer uint8

	// Sound timer
	SoundTimer uint8

	// Display buffer (64x32 monochrome)
	Display [DisplayWidth * DisplayHeight]uint8

	// Keypad state (16 keys)
	Keys [NumKeys]bool

	// Flag indicating if display needs to be redrawn
	DrawFlag bool

	// Flag indicating if we're waiting for a key press
	WaitingForKey bool

	// Register to store the pressed key
	KeyRegister uint8
}

// Fontset contains the built-in CHIP-8 font sprites (0-F)
// Each character is 5 bytes tall
var Fontset = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

// New creates and initializes a new CHIP-8 virtual machine
func New() *CHIP8 {
	c := &CHIP8{}
	c.Reset()
	return c
}

// Reset resets the CHIP-8 to its initial state
func (c *CHIP8) Reset() {
	// Clear memory
	for i := range c.Memory {
		c.Memory[i] = 0
	}

	// Clear registers
	for i := range c.V {
		c.V[i] = 0
	}

	// Clear stack
	for i := range c.Stack {
		c.Stack[i] = 0
	}

	// Clear display
	for i := range c.Display {
		c.Display[i] = 0
	}

	// Clear keys
	for i := range c.Keys {
		c.Keys[i] = false
	}

	// Reset other state
	c.I = 0
	c.PC = ProgramStart
	c.SP = 0
	c.DelayTimer = 0
	c.SoundTimer = 0
	c.DrawFlag = true
	c.WaitingForKey = false
	c.KeyRegister = 0

	// Load fontset into memory (starting at 0x000)
	for i, b := range Fontset {
		c.Memory[i] = b
	}
}

// LoadROM loads a ROM file into memory starting at 0x200
func (c *CHIP8) LoadROM(data []byte) error {
	if len(data) > MemorySize-ProgramStart {
		return fmt.Errorf("ROM too large: %d bytes (max %d)", len(data), MemorySize-ProgramStart)
	}

	for i, b := range data {
		c.Memory[ProgramStart+i] = b
	}

	return nil
}

// SetKey sets the state of a key (0-15)
func (c *CHIP8) SetKey(key uint8, pressed bool) {
	if key < NumKeys {
		c.Keys[key] = pressed

		// If we're waiting for a key and a key was pressed
		if c.WaitingForKey && pressed {
			c.V[c.KeyRegister] = key
			c.WaitingForKey = false
		}
	}
}

// UpdateTimers decrements the delay and sound timers (should be called at 60Hz)
func (c *CHIP8) UpdateTimers() {
	if c.DelayTimer > 0 {
		c.DelayTimer--
	}
	if c.SoundTimer > 0 {
		c.SoundTimer--
	}
}

// ShouldBeep returns true if the sound timer is active
func (c *CHIP8) ShouldBeep() bool {
	return c.SoundTimer > 0
}

// Cycle executes one CPU cycle (fetch, decode, execute)
func (c *CHIP8) Cycle() error {
	// Don't execute if waiting for key press
	if c.WaitingForKey {
		return nil
	}

	// Fetch opcode (2 bytes, big-endian)
	opcode := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])

	// Increment program counter before execution
	c.PC += 2

	// Execute opcode
	return c.executeOpcode(opcode)
}

// executeOpcode decodes and executes a single opcode
func (c *CHIP8) executeOpcode(opcode uint16) error {
	// Extract common opcode parts
	x := uint8((opcode & 0x0F00) >> 8)  // Second nibble
	y := uint8((opcode & 0x00F0) >> 4)  // Third nibble
	n := uint8(opcode & 0x000F)         // Fourth nibble
	nn := uint8(opcode & 0x00FF)        // Second byte
	nnn := opcode & 0x0FFF              // Last three nibbles

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // 00E0: Clear screen
			for i := range c.Display {
				c.Display[i] = 0
			}
			c.DrawFlag = true
		case 0x00EE: // 00EE: Return from subroutine
			if c.SP == 0 {
				return fmt.Errorf("stack underflow")
			}
			c.SP--
			c.PC = c.Stack[c.SP]
		default:
			// 0NNN: Call machine code routine (ignored on modern interpreters)
		}

	case 0x1000: // 1NNN: Jump to address NNN
		c.PC = nnn

	case 0x2000: // 2NNN: Call subroutine at NNN
		if c.SP >= StackSize {
			return fmt.Errorf("stack overflow")
		}
		c.Stack[c.SP] = c.PC
		c.SP++
		c.PC = nnn

	case 0x3000: // 3XNN: Skip next instruction if VX == NN
		if c.V[x] == nn {
			c.PC += 2
		}

	case 0x4000: // 4XNN: Skip next instruction if VX != NN
		if c.V[x] != nn {
			c.PC += 2
		}

	case 0x5000: // 5XY0: Skip next instruction if VX == VY
		if c.V[x] == c.V[y] {
			c.PC += 2
		}

	case 0x6000: // 6XNN: Set VX to NN
		c.V[x] = nn

	case 0x7000: // 7XNN: Add NN to VX (no carry flag)
		c.V[x] += nn

	case 0x8000:
		switch n {
		case 0x0: // 8XY0: Set VX to VY
			c.V[x] = c.V[y]
		case 0x1: // 8XY1: Set VX to VX OR VY
			c.V[x] |= c.V[y]
		case 0x2: // 8XY2: Set VX to VX AND VY
			c.V[x] &= c.V[y]
		case 0x3: // 8XY3: Set VX to VX XOR VY
			c.V[x] ^= c.V[y]
		case 0x4: // 8XY4: Add VY to VX, VF = carry
			sum := uint16(c.V[x]) + uint16(c.V[y])
			c.V[x] = uint8(sum)
			if sum > 255 {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
		case 0x5: // 8XY5: Subtract VY from VX, VF = NOT borrow
			if c.V[x] >= c.V[y] {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[x] -= c.V[y]
		case 0x6: // 8XY6: Shift VX right, VF = LSB before shift
			c.V[0xF] = c.V[x] & 0x1
			c.V[x] >>= 1
		case 0x7: // 8XY7: Set VX to VY - VX, VF = NOT borrow
			if c.V[y] >= c.V[x] {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[x] = c.V[y] - c.V[x]
		case 0xE: // 8XYE: Shift VX left, VF = MSB before shift
			c.V[0xF] = (c.V[x] & 0x80) >> 7
			c.V[x] <<= 1
		default:
			return fmt.Errorf("unknown opcode: 0x%04X", opcode)
		}

	case 0x9000: // 9XY0: Skip next instruction if VX != VY
		if c.V[x] != c.V[y] {
			c.PC += 2
		}

	case 0xA000: // ANNN: Set I to NNN
		c.I = nnn

	case 0xB000: // BNNN: Jump to NNN + V0
		c.PC = nnn + uint16(c.V[0])

	case 0xC000: // CXNN: Set VX to random byte AND NN
		c.V[x] = uint8(rand.Intn(256)) & nn

	case 0xD000: // DXYN: Draw sprite at (VX, VY) with N bytes of sprite data starting at I
		c.V[0xF] = 0
		for row := uint8(0); row < n; row++ {
			sprite := c.Memory[c.I+uint16(row)]
			for col := uint8(0); col < 8; col++ {
				if (sprite & (0x80 >> col)) != 0 {
					px := (c.V[x] + col) % DisplayWidth
					py := (c.V[y] + row) % DisplayHeight
					idx := int(py)*DisplayWidth + int(px)
					if c.Display[idx] == 1 {
						c.V[0xF] = 1
					}
					c.Display[idx] ^= 1
				}
			}
		}
		c.DrawFlag = true

	case 0xE000:
		switch nn {
		case 0x9E: // EX9E: Skip next instruction if key VX is pressed
			if c.Keys[c.V[x]] {
				c.PC += 2
			}
		case 0xA1: // EXA1: Skip next instruction if key VX is not pressed
			if !c.Keys[c.V[x]] {
				c.PC += 2
			}
		default:
			return fmt.Errorf("unknown opcode: 0x%04X", opcode)
		}

	case 0xF000:
		switch nn {
		case 0x07: // FX07: Set VX to delay timer
			c.V[x] = c.DelayTimer
		case 0x0A: // FX0A: Wait for key press, store in VX
			c.WaitingForKey = true
			c.KeyRegister = x
		case 0x15: // FX15: Set delay timer to VX
			c.DelayTimer = c.V[x]
		case 0x18: // FX18: Set sound timer to VX
			c.SoundTimer = c.V[x]
		case 0x1E: // FX1E: Add VX to I
			c.I += uint16(c.V[x])
		case 0x29: // FX29: Set I to location of font character VX
			c.I = uint16(c.V[x]) * 5
		case 0x33: // FX33: Store BCD of VX at I, I+1, I+2
			c.Memory[c.I] = c.V[x] / 100
			c.Memory[c.I+1] = (c.V[x] / 10) % 10
			c.Memory[c.I+2] = c.V[x] % 10
		case 0x55: // FX55: Store V0-VX in memory starting at I
			for i := uint8(0); i <= x; i++ {
				c.Memory[c.I+uint16(i)] = c.V[i]
			}
		case 0x65: // FX65: Load V0-VX from memory starting at I
			for i := uint8(0); i <= x; i++ {
				c.V[i] = c.Memory[c.I+uint16(i)]
			}
		default:
			return fmt.Errorf("unknown opcode: 0x%04X", opcode)
		}

	default:
		return fmt.Errorf("unknown opcode: 0x%04X", opcode)
	}

	return nil
}
