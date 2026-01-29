package chip8

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New()

	// Check initial state
	if c.PC != ProgramStart {
		t.Errorf("PC should be %#x, got %#x", ProgramStart, c.PC)
	}

	if c.SP != 0 {
		t.Errorf("SP should be 0, got %d", c.SP)
	}

	if c.I != 0 {
		t.Errorf("I should be 0, got %d", c.I)
	}

	// Check fontset is loaded
	if c.Memory[0] != 0xF0 {
		t.Errorf("Fontset not loaded correctly, first byte should be 0xF0, got %#x", c.Memory[0])
	}
}

func TestReset(t *testing.T) {
	c := New()

	// Modify some state
	c.PC = 0x300
	c.V[0] = 42
	c.I = 100
	c.SP = 5
	c.DelayTimer = 10

	// Reset
	c.Reset()

	// Verify reset state
	if c.PC != ProgramStart {
		t.Errorf("After reset, PC should be %#x, got %#x", ProgramStart, c.PC)
	}

	if c.V[0] != 0 {
		t.Errorf("After reset, V0 should be 0, got %d", c.V[0])
	}

	if c.I != 0 {
		t.Errorf("After reset, I should be 0, got %d", c.I)
	}

	if c.SP != 0 {
		t.Errorf("After reset, SP should be 0, got %d", c.SP)
	}

	if c.DelayTimer != 0 {
		t.Errorf("After reset, DelayTimer should be 0, got %d", c.DelayTimer)
	}
}

func TestLoadROM(t *testing.T) {
	c := New()

	rom := []byte{0x00, 0xE0, 0x12, 0x00} // CLS; JP 0x200
	err := c.LoadROM(rom)

	if err != nil {
		t.Errorf("LoadROM failed: %v", err)
	}

	// Check ROM is loaded at correct address
	if c.Memory[ProgramStart] != 0x00 {
		t.Errorf("ROM not loaded at correct address")
	}

	if c.Memory[ProgramStart+1] != 0xE0 {
		t.Errorf("ROM not loaded correctly")
	}
}

func TestLoadROMTooLarge(t *testing.T) {
	c := New()

	// Create ROM larger than available memory
	rom := make([]byte, MemorySize)
	err := c.LoadROM(rom)

	if err == nil {
		t.Error("LoadROM should fail for oversized ROM")
	}
}

func TestOpcode00E0_ClearScreen(t *testing.T) {
	c := New()

	// Set some pixels
	c.Display[0] = 1
	c.Display[100] = 1
	c.Display[500] = 1

	// Load CLS opcode
	c.Memory[ProgramStart] = 0x00
	c.Memory[ProgramStart+1] = 0xE0

	// Execute
	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	// Check display is cleared
	for i, pixel := range c.Display {
		if pixel != 0 {
			t.Errorf("Display[%d] should be 0 after CLS", i)
			break
		}
	}
}

func TestOpcode1NNN_Jump(t *testing.T) {
	c := New()

	// Load JP 0x400
	c.Memory[ProgramStart] = 0x14
	c.Memory[ProgramStart+1] = 0x00

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.PC != 0x400 {
		t.Errorf("PC should be 0x400 after JP, got %#x", c.PC)
	}
}

func TestOpcode2NNN_Call(t *testing.T) {
	c := New()

	// Load CALL 0x400
	c.Memory[ProgramStart] = 0x24
	c.Memory[ProgramStart+1] = 0x00

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.PC != 0x400 {
		t.Errorf("PC should be 0x400 after CALL, got %#x", c.PC)
	}

	if c.SP != 1 {
		t.Errorf("SP should be 1 after CALL, got %d", c.SP)
	}

	if c.Stack[0] != ProgramStart+2 {
		t.Errorf("Stack[0] should be %#x, got %#x", ProgramStart+2, c.Stack[0])
	}
}

func TestOpcode00EE_Return(t *testing.T) {
	c := New()

	// Setup: push address to stack
	c.Stack[0] = 0x300
	c.SP = 1
	c.PC = 0x400

	// Load RET opcode at 0x400
	c.Memory[0x400] = 0x00
	c.Memory[0x401] = 0xEE

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.PC != 0x300 {
		t.Errorf("PC should be 0x300 after RET, got %#x", c.PC)
	}

	if c.SP != 0 {
		t.Errorf("SP should be 0 after RET, got %d", c.SP)
	}
}

func TestOpcode3XNN_SkipEqual(t *testing.T) {
	c := New()
	c.V[0] = 0x42

	// Load SE V0, 0x42
	c.Memory[ProgramStart] = 0x30
	c.Memory[ProgramStart+1] = 0x42

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	// Should skip (PC += 4 total: 2 for instruction + 2 for skip)
	if c.PC != ProgramStart+4 {
		t.Errorf("PC should be %#x after SE (equal), got %#x", ProgramStart+4, c.PC)
	}
}

func TestOpcode3XNN_NoSkipNotEqual(t *testing.T) {
	c := New()
	c.V[0] = 0x41

	// Load SE V0, 0x42
	c.Memory[ProgramStart] = 0x30
	c.Memory[ProgramStart+1] = 0x42

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	// Should not skip (PC += 2)
	if c.PC != ProgramStart+2 {
		t.Errorf("PC should be %#x after SE (not equal), got %#x", ProgramStart+2, c.PC)
	}
}

func TestOpcode6XNN_SetRegister(t *testing.T) {
	c := New()

	// Load LD V5, 0xAB
	c.Memory[ProgramStart] = 0x65
	c.Memory[ProgramStart+1] = 0xAB

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.V[5] != 0xAB {
		t.Errorf("V5 should be 0xAB, got %#x", c.V[5])
	}
}

func TestOpcode7XNN_AddToRegister(t *testing.T) {
	c := New()
	c.V[0] = 0x10

	// Load ADD V0, 0x05
	c.Memory[ProgramStart] = 0x70
	c.Memory[ProgramStart+1] = 0x05

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.V[0] != 0x15 {
		t.Errorf("V0 should be 0x15, got %#x", c.V[0])
	}
}

func TestOpcode8XY0_SetVXtoVY(t *testing.T) {
	c := New()
	c.V[1] = 0x42

	// Load LD V0, V1
	c.Memory[ProgramStart] = 0x80
	c.Memory[ProgramStart+1] = 0x10

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.V[0] != 0x42 {
		t.Errorf("V0 should be 0x42, got %#x", c.V[0])
	}
}

func TestOpcode8XY4_AddWithCarry(t *testing.T) {
	c := New()
	c.V[0] = 0xFF
	c.V[1] = 0x02

	// Load ADD V0, V1
	c.Memory[ProgramStart] = 0x80
	c.Memory[ProgramStart+1] = 0x14

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.V[0] != 0x01 {
		t.Errorf("V0 should be 0x01 (overflow), got %#x", c.V[0])
	}

	if c.V[0xF] != 1 {
		t.Errorf("VF should be 1 (carry), got %d", c.V[0xF])
	}
}

func TestOpcode8XY5_SubWithBorrow(t *testing.T) {
	c := New()
	c.V[0] = 0x10
	c.V[1] = 0x05

	// Load SUB V0, V1
	c.Memory[ProgramStart] = 0x80
	c.Memory[ProgramStart+1] = 0x15

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.V[0] != 0x0B {
		t.Errorf("V0 should be 0x0B, got %#x", c.V[0])
	}

	if c.V[0xF] != 1 {
		t.Errorf("VF should be 1 (no borrow), got %d", c.V[0xF])
	}
}

func TestOpcodeANNN_SetI(t *testing.T) {
	c := New()

	// Load LD I, 0x456
	c.Memory[ProgramStart] = 0xA4
	c.Memory[ProgramStart+1] = 0x56

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.I != 0x456 {
		t.Errorf("I should be 0x456, got %#x", c.I)
	}
}

func TestOpcodeFX33_BCD(t *testing.T) {
	c := New()
	c.V[0] = 123
	c.I = 0x300

	// Load LD B, V0
	c.Memory[ProgramStart] = 0xF0
	c.Memory[ProgramStart+1] = 0x33

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.Memory[0x300] != 1 {
		t.Errorf("Memory[I] should be 1, got %d", c.Memory[0x300])
	}

	if c.Memory[0x301] != 2 {
		t.Errorf("Memory[I+1] should be 2, got %d", c.Memory[0x301])
	}

	if c.Memory[0x302] != 3 {
		t.Errorf("Memory[I+2] should be 3, got %d", c.Memory[0x302])
	}
}

func TestOpcodeFX55_StoreRegisters(t *testing.T) {
	c := New()
	c.I = 0x300
	c.V[0] = 0xAA
	c.V[1] = 0xBB
	c.V[2] = 0xCC

	// Load LD [I], V2
	c.Memory[ProgramStart] = 0xF2
	c.Memory[ProgramStart+1] = 0x55

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.Memory[0x300] != 0xAA {
		t.Errorf("Memory[I] should be 0xAA, got %#x", c.Memory[0x300])
	}

	if c.Memory[0x301] != 0xBB {
		t.Errorf("Memory[I+1] should be 0xBB, got %#x", c.Memory[0x301])
	}

	if c.Memory[0x302] != 0xCC {
		t.Errorf("Memory[I+2] should be 0xCC, got %#x", c.Memory[0x302])
	}
}

func TestOpcodeFX65_LoadRegisters(t *testing.T) {
	c := New()
	c.I = 0x300
	c.Memory[0x300] = 0xAA
	c.Memory[0x301] = 0xBB
	c.Memory[0x302] = 0xCC

	// Load LD V2, [I]
	c.Memory[ProgramStart] = 0xF2
	c.Memory[ProgramStart+1] = 0x65

	err := c.Cycle()
	if err != nil {
		t.Errorf("Cycle failed: %v", err)
	}

	if c.V[0] != 0xAA {
		t.Errorf("V0 should be 0xAA, got %#x", c.V[0])
	}

	if c.V[1] != 0xBB {
		t.Errorf("V1 should be 0xBB, got %#x", c.V[1])
	}

	if c.V[2] != 0xCC {
		t.Errorf("V2 should be 0xCC, got %#x", c.V[2])
	}
}

func TestUpdateTimers(t *testing.T) {
	c := New()
	c.DelayTimer = 5
	c.SoundTimer = 3

	c.UpdateTimers()

	if c.DelayTimer != 4 {
		t.Errorf("DelayTimer should be 4, got %d", c.DelayTimer)
	}

	if c.SoundTimer != 2 {
		t.Errorf("SoundTimer should be 2, got %d", c.SoundTimer)
	}
}

func TestSetKey(t *testing.T) {
	c := New()

	c.SetKey(5, true)
	if !c.Keys[5] {
		t.Error("Key 5 should be pressed")
	}

	c.SetKey(5, false)
	if c.Keys[5] {
		t.Error("Key 5 should be released")
	}
}

func TestWaitingForKey(t *testing.T) {
	c := New()
	c.WaitingForKey = true
	c.KeyRegister = 3

	// Pressing a key should store it and stop waiting
	c.SetKey(0xA, true)

	if c.WaitingForKey {
		t.Error("Should no longer be waiting for key")
	}

	if c.V[3] != 0xA {
		t.Errorf("V3 should be 0xA, got %#x", c.V[3])
	}
}

func TestShouldBeep(t *testing.T) {
	c := New()

	if c.ShouldBeep() {
		t.Error("Should not beep when SoundTimer is 0")
	}

	c.SoundTimer = 5
	if !c.ShouldBeep() {
		t.Error("Should beep when SoundTimer > 0")
	}
}
