package core

import (
	"math/rand"
)

type Emulator struct {
	ProgramCounter uint16
	Ram            [RAM_SIZE]uint8
	Screen         [SCREEN_WIDTH][SCREEN_HEIGHT]bool
	VRegisters     [NUM_REGS]uint8
	IRegister      uint16
	Stack          [STACK_SIZE]uint16
	StackPointer   uint16
	DelayTimer     uint8
	Keys           [NUM_KEYS]bool
}

func NewEmulator() *Emulator {
	emu := &Emulator{
		ProgramCounter: START_ADDRESS,
	}

	copy(emu.Ram[:FONTSET_SIZE], FONTSET[:])
	return emu
}

func (e *Emulator) Reset() {
	e.ProgramCounter = START_ADDRESS
	e.Ram = [RAM_SIZE]uint8{}
	e.Screen = [SCREEN_WIDTH][SCREEN_HEIGHT]bool{}
	e.VRegisters = [NUM_REGS]uint8{}
	e.IRegister = 0
	e.Stack = [STACK_SIZE]uint16{}
	e.StackPointer = 0
	e.DelayTimer = 0
	e.Keys = [NUM_KEYS]bool{}

	copy(e.Ram[:FONTSET_SIZE], FONTSET[:])
}

// fetch-decode-execute cycle
func (e *Emulator) Tick() {
	op := e.fetch()
	e.decodeAndExecute(op)
}

// each instruction in chip-8 is two bytes long
// ram stores values as 8-bit
func (e *Emulator) fetch() uint16 {
	msb := uint16(e.Ram[e.ProgramCounter])
	lsb := uint16(e.Ram[e.ProgramCounter+1])
	op := (msb << 8) | lsb // converting to big endian
	e.ProgramCounter += 2
	return op
}

func (e *Emulator) decodeAndExecute(op uint16) {
	digit1 := (op & 0xF000) >> 12
	digit2 := (op & 0x0F00) >> 8
	digit3 := (op & 0x00F0) >> 4
	digit4 := (op & 0x000F)

	switch {
	// noop (0x0000)
	case op == 0x0000:
		return
	// clear screen (0x00E0)
	case op == 0x00E0:
		e.Screen = [SCREEN_WIDTH][SCREEN_HEIGHT]bool{}
	// return from subroutine (0x00EE)
	case op == 0x00EE:
		returnAddress := e.popFromStack()
		e.ProgramCounter = returnAddress
	// jump (0x1NNN)
	// jump to address 0xNNN
	case digit1 == 1:
		nnn := op & 0xfff
		e.ProgramCounter = op & nnn
	// call subroutine (0x2NNN)
	// enter subroutine at 0xNNN, add current PC to stack so that the executation can be later continued from here
	case digit1 == 2:
		e.pushToStack(e.ProgramCounter)
		nnn := op & 0xFFF
		e.ProgramCounter = nnn
	// skip if vx == nn (0x3XNN)
	case digit1 == 3:
		x := digit2
		nn := op & 0xFF

		if e.VRegisters[x] == uint8(nn) {
			e.ProgramCounter += 2
		}
	// skip if vx != nn (0x4XNN)
	case digit1 == 4:
		x := digit2
		nn := op & 0xFF

		if e.VRegisters[x] != uint8(nn) {
			e.ProgramCounter += 2
		}
	// skip next if vx == xy (0x5XY0)
	case digit1 == 5, digit4 == 0:
		x := digit2
		y := digit3

		if e.VRegisters[x] == e.VRegisters[y] {
			e.ProgramCounter += 2
		}
	// vx = nn (0x6XNN)
	case digit1 == 6:
		x := digit2
		nn := op & 0xFF

		e.VRegisters[x] = uint8(nn)
	// vx = vx + nn (0x7XNN)
	case digit1 == 7:
		x := digit2
		nn := op & 0xFF

		e.VRegisters[x] += uint8(nn)
	case digit1 == 8:
		x := digit2
		y := digit3

		switch digit4 {
		case 0:
			// vx = vy (0x8XY0)
			e.VRegisters[x] = e.VRegisters[y]
		case 1:
			// vx |= vy (0x8XY1)
			e.VRegisters[x] |= e.VRegisters[y]
		case 2:
			// vx &= vy (0x8XY2)
			e.VRegisters[x] &= e.VRegisters[y]
		case 3:
			// vx ^= vy (0x8XY3)
			e.VRegisters[x] ^= e.VRegisters[y]
		case 4:
			// vx += vy (0x8XY4)
			sum := uint16(e.VRegisters[x]) + uint16(e.VRegisters[y])
			e.VRegisters[x] = uint8(sum)

			// if overflows occurs, 0xf v-register is set to 1 else 0
			if sum > 0xff {
				e.VRegisters[0xF] = 1
			} else {
				e.VRegisters[0xF] = 0
			}
		case 5:
			// vx -= vy (0x8XY5)
			diff := int16(e.VRegisters[x]) - int16(e.VRegisters[y])
			e.VRegisters[x] = uint8(diff)

			// if underflow occurs, 0xf v-register is set to 0 else 1
			if diff < 0 {
				e.VRegisters[0xF] = 0
			} else {
				e.VRegisters[0xF] = 1
			}
		case 6:
			// vx >>= 1 (0x8XY6)
			lsb := e.VRegisters[x] & 1
			e.VRegisters[x] >>= 1
			e.VRegisters[0xF] = lsb
		case 7:
			// vx = vy - vx (0x8XY7)
			diff := int16(e.VRegisters[y]) - int16(e.VRegisters[x])
			e.VRegisters[x] = uint8(diff)

			// if underflow occurs, 0xf v-register is set to 0 else 1
			if diff < 0 {
				e.VRegisters[0xF] = 0
			} else {
				e.VRegisters[0xF] = 1
			}
		case 0xE:
			// vx <<= 1 (0x8XYE)
			msb := (e.VRegisters[x] >> 7) & 1
			e.VRegisters[x] <<= 1
			e.VRegisters[0xF] = msb
		}
	// skip if vx != vy (0x9XY0)
	case digit1 == 9 && digit4 == 0:
		x := digit2
		y := digit3

		if e.VRegisters[x] != e.VRegisters[y] {
			e.ProgramCounter += 2
		}
	// jump to 0xNNN in RAM (0xANNN)
	case digit1 == 0xA:
		nnn := op & 0xFFF
		e.IRegister = nnn
	// jump to v0 + nnn (0xBNNN)
	case digit1 == 0xB:
		nnn := op & 0xFFF
		e.ProgramCounter = uint16(e.VRegisters[0]) + nnn
	// vx = rand() & nn (rng) (0xCXNN)
	case digit1 == 0xC:
		x := digit2

		nn := uint8(op & 0xFF)
		rng := uint8(rand.Uint32())
		e.VRegisters[x] = rng & nn
	}
}

func (e *Emulator) tickDelayTimer() {
	if e.DelayTimer > 0 {
		e.DelayTimer--
	}
}

func (e *Emulator) popFromStack() uint16 {
	e.StackPointer--
	return e.Stack[e.StackPointer]
}

func (e *Emulator) pushToStack(val uint16) {
	if e.StackPointer > STACK_SIZE {
		panic("stack overflow")
	}

	e.Stack[e.StackPointer] = val
	e.StackPointer++
}
