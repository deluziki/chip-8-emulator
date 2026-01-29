# CHIP-8 Emulator

A complete CHIP-8 emulator written in Go using SDL2 for graphics, audio, and input.

## What is CHIP-8?

CHIP-8 is an interpreted programming language developed in the mid-1970s for 8-bit microcomputers. It was designed to allow video games to be more easily programmed and portable. CHIP-8 programs run on a virtual machine with:

- 4KB of memory
- 16 general-purpose 8-bit registers (V0-VF)
- A 16-bit index register (I)
- A 16-bit program counter (PC)
- A 64x32 monochrome display
- A 16-key hexadecimal keypad
- Delay and sound timers

## Features

- Complete implementation of all 35 CHIP-8 opcodes
- Accurate timing with configurable CPU speed
- Delay and sound timer support
- Beeper audio output
- Scalable display window
- Keyboard input mapping
- Pause, reset, and quit controls

## Requirements

- Go 1.20 or later
- SDL2 development libraries

### Installing SDL2

**Ubuntu/Debian:**
```bash
sudo apt-get install libsdl2-dev
```

**macOS (Homebrew):**
```bash
brew install sdl2
```

**Windows:**
Download SDL2 development libraries from https://www.libsdl.org/download-2.0.php

## Building

```bash
# Clone the repository
git clone https://github.com/chip8-emulator
cd chip8-emulator

# Build the emulator
make build
# or
go build -o chip8-emulator .
```

## Usage

```bash
# Run a ROM
./chip8-emulator path/to/rom.ch8

# With options
./chip8-emulator -scale 15 -speed 700 path/to/rom.ch8
```

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-rom` | - | Path to the CHIP-8 ROM file |
| `-scale` | 10 | Display scale factor |
| `-speed` | 500 | CPU speed in Hz (instructions per second) |

### Keyboard Controls

**Emulator Controls:**
- `ESC` - Quit emulator
- `P` - Pause/Resume
- `R` - Reset and reload ROM

**CHIP-8 Keypad Mapping:**

```
CHIP-8 Keypad:       Keyboard:
+---+---+---+---+    +---+---+---+---+
| 1 | 2 | 3 | C |    | 1 | 2 | 3 | 4 |
+---+---+---+---+    +---+---+---+---+
| 4 | 5 | 6 | D | => | Q | W | E | R |
+---+---+---+---+    +---+---+---+---+
| 7 | 8 | 9 | E |    | A | S | D | F |
+---+---+---+---+    +---+---+---+---+
| A | 0 | B | F |    | Z | X | C | V |
+---+---+---+---+    +---+---+---+---+
```

## Project Structure

```
chip8-emulator/
├── main.go           # Entry point and main loop
├── chip8/
│   └── chip8.go      # CPU core and opcode implementation
├── display/
│   └── display.go    # SDL2 graphics rendering
├── input/
│   └── input.go      # Keyboard input handling
├── audio/
│   └── audio.go      # Sound/beeper output
├── Makefile          # Build automation
└── README.md         # This file
```

## Finding ROMs

CHIP-8 ROMs can be found in various public domain ROM collections. Popular games include:
- PONG
- TETRIS
- SPACE INVADERS
- BRIX
- MAZE

## Technical Details

### Memory Map
```
0x000-0x1FF - Reserved (font data)
0x200-0xFFF - Program/Data space
```

### Opcodes Implemented

All standard CHIP-8 opcodes are implemented:
- `0NNN` - Call machine code routine (ignored)
- `00E0` - Clear screen
- `00EE` - Return from subroutine
- `1NNN` - Jump to address
- `2NNN` - Call subroutine
- `3XNN` - Skip if VX == NN
- `4XNN` - Skip if VX != NN
- `5XY0` - Skip if VX == VY
- `6XNN` - Set VX = NN
- `7XNN` - Add NN to VX
- `8XY0-8XYE` - Arithmetic/logic operations
- `9XY0` - Skip if VX != VY
- `ANNN` - Set I = NNN
- `BNNN` - Jump to NNN + V0
- `CXNN` - Random AND NN
- `DXYN` - Draw sprite
- `EX9E` - Skip if key pressed
- `EXA1` - Skip if key not pressed
- `FX07-FX65` - Timer, I/O, and memory operations

## License

MIT License
