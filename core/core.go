package core

import (
	"math/rand"
)

type Emulator struct {
	programCounter uint16
	ram            [RAM_SIZE]uint8
	screen         [SCREEN_WIDTH * SCREEN_HEIGHT]bool
	vRegisters     [NUM_REGS]uint8
	iRegister      uint16
	stack          [STACK_SIZE]uint16
	stackPointer   uint16
	delayTimer     uint8
	keys           [NUM_KEYS]bool
}

func NewEmulator() *Emulator {
	emu := &Emulator{
		programCounter: START_ADDRESS,
	}

	copy(emu.ram[:FONTSET_SIZE], FONTSET[:])
	return emu
}

func (e *Emulator) Reset() {
	e.programCounter = START_ADDRESS
	e.ram = [RAM_SIZE]uint8{}
	e.screen = [SCREEN_WIDTH * SCREEN_HEIGHT]bool{}
	e.vRegisters = [NUM_REGS]uint8{}
	e.iRegister = 0
	e.stack = [STACK_SIZE]uint16{}
	e.stackPointer = 0
	e.delayTimer = 0
	e.keys = [NUM_KEYS]bool{}

	copy(e.ram[:FONTSET_SIZE], FONTSET[:])
}

// fetch-decode-execute cycle
func (e *Emulator) Tick() {
	op := e.fetch()
	e.decodeAndExecute(op)
}

func (e *Emulator) GetDisplay() [SCREEN_HEIGHT * SCREEN_WIDTH]bool {
	return e.screen
}

func (e *Emulator) Keypress(idx int, pressed bool) {
	e.keys[idx] = pressed
}

func (e *Emulator) LoadData(data []uint8) {
	start := START_ADDRESS
	end := START_ADDRESS + len(data)

	copy(e.ram[start:end], data[:])
}

func (e *Emulator) TickTimers() {
	if e.delayTimer > 0 {
		e.delayTimer--
	}
}

// each instruction in chip-8 is two bytes long
// ram stores values as 8-bit
func (e *Emulator) fetch() uint16 {
	msb := uint16(e.ram[e.programCounter])
	lsb := uint16(e.ram[e.programCounter+1])
	op := (msb << 8) | lsb // converting to big endian
	e.programCounter += 2
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
		e.screen = [SCREEN_WIDTH * SCREEN_HEIGHT]bool{}
	// return from subroutine (0x00EE)
	case op == 0x00EE:
		returnAddress := e.popFromStack()
		e.programCounter = returnAddress
	// jump (0x1NNN)
	// jump to address 0xNNN
	case digit1 == 1:
		e.programCounter = op & 0xfff
	// call subroutine (0x2NNN)
	// enter subroutine at 0xNNN, add current PC to stack so that the executation can be later continued from here
	case digit1 == 2:
		e.pushToStack(e.programCounter)
		nnn := op & 0xFFF
		e.programCounter = nnn
	// skip if vx == nn (0x3XNN)
	case digit1 == 3:
		x := digit2
		nn := op & 0xFF

		if e.vRegisters[x] == uint8(nn) {
			e.programCounter += 2
		}
	// skip if vx != nn (0x4XNN)
	case digit1 == 4:
		x := digit2
		nn := op & 0xFF

		if e.vRegisters[x] != uint8(nn) {
			e.programCounter += 2
		}
	// skip next if vx == xy (0x5XY0)
	case digit1 == 5 && digit4 == 0:
		x := digit2
		y := digit3

		if e.vRegisters[x] == e.vRegisters[y] {
			e.programCounter += 2
		}
	// vx = nn (0x6XNN)
	case digit1 == 6:
		x := digit2
		nn := op & 0xFF

		e.vRegisters[x] = uint8(nn)
	// vx = vx + nn (0x7XNN)
	case digit1 == 7:
		x := digit2
		nn := op & 0xFF

		e.vRegisters[x] += uint8(nn)
	case digit1 == 8:
		x := digit2
		y := digit3

		switch digit4 {
		case 0:
			// vx = vy (0x8XY0)
			e.vRegisters[x] = e.vRegisters[y]
		case 1:
			// vx |= vy (0x8XY1)
			e.vRegisters[x] |= e.vRegisters[y]
		case 2:
			// vx &= vy (0x8XY2)
			e.vRegisters[x] &= e.vRegisters[y]
		case 3:
			// vx ^= vy (0x8XY3)
			e.vRegisters[x] ^= e.vRegisters[y]
		case 4:
			// vx += vy (0x8XY4)
			sum := uint16(e.vRegisters[x]) + uint16(e.vRegisters[y])
			e.vRegisters[x] = uint8(sum)

			// if overflows occurs, 0xf v-register is set to 1 else 0
			if sum > 0xff {
				e.vRegisters[0xF] = 1
			} else {
				e.vRegisters[0xF] = 0
			}
		case 5:
			// vx -= vy (0x8XY5)
			diff := int16(e.vRegisters[x]) - int16(e.vRegisters[y])
			e.vRegisters[x] = uint8(diff)

			// if underflow occurs, 0xf v-register is set to 0 else 1
			if diff < 0 {
				e.vRegisters[0xF] = 0
			} else {
				e.vRegisters[0xF] = 1
			}
		case 6:
			// vx >>= 1 (0x8XY6)
			lsb := e.vRegisters[x] & 1
			e.vRegisters[x] >>= 1
			e.vRegisters[0xF] = lsb
		case 7:
			// vx = vy - vx (0x8XY7)
			diff := int16(e.vRegisters[y]) - int16(e.vRegisters[x])
			e.vRegisters[x] = uint8(diff)

			// if underflow occurs, 0xf v-register is set to 0 else 1
			if diff < 0 {
				e.vRegisters[0xF] = 0
			} else {
				e.vRegisters[0xF] = 1
			}
		case 0xE:
			// vx <<= 1 (0x8XYE)
			msb := (e.vRegisters[x] >> 7) & 1
			e.vRegisters[x] <<= 1
			e.vRegisters[0xF] = msb
		}
	// skip if vx != vy (0x9XY0)
	case digit1 == 9 && digit4 == 0:
		x := digit2
		y := digit3

		if e.vRegisters[x] != e.vRegisters[y] {
			e.programCounter += 2
		}
	// jump to 0xNNN in ram (0xANNN)
	case digit1 == 0xA:
		nnn := op & 0xFFF
		e.iRegister = nnn
	// jump to v0 + nnn (0xBNNN)
	case digit1 == 0xB:
		nnn := op & 0xFFF
		e.programCounter = uint16(e.vRegisters[0]) + nnn
	// vx = rand() & nn (rng) (0xCXNN)
	case digit1 == 0xC:
		x := digit2

		nn := uint8(op & 0xFF)
		rng := uint8(rand.Uint32())
		e.vRegisters[x] = rng & nn
	// draw a pixel at (VX, VY) of height N pixels (0xDXYN)
	case digit1 == 0xD:
		xCoord := uint16(e.vRegisters[digit2])
		yCoord := uint16(e.vRegisters[digit3])
		numRows := digit4

		flipped := false

		for i := range numRows {
			address := e.iRegister + i
			pixels := e.ram[address]

			for bitIdx := range 8 {
				bit := pixels & (0b1000_0000 >> bitIdx)

				if bit != 0 {
					x := (xCoord + uint16(bitIdx)) % SCREEN_WIDTH
					y := (yCoord + i) % SCREEN_HEIGHT

					idx := x + (SCREEN_WIDTH * y)

					// check if the pixel is about to be flipped and compute that over the entire screen
					flipped = flipped || e.screen[idx]
					// if the current pixel in the sprite row is on and the pixel at (X, Y) on screen is also on
					// turn off the pixel
					// XOR with true = NOT
					e.screen[idx] = !e.screen[idx]
				}
			}
		}

		if flipped {
			e.vRegisters[0xF] = 1
		} else {
			e.vRegisters[0xF] = 0
		}
	// skip if key pressed (0xEX9E)
	case digit1 == 0xE && digit3 == 9 && digit4 == 0xE:
		x := digit2
		vx := e.vRegisters[x]
		key := e.keys[vx]

		if key {
			e.programCounter += 2
		}
	// skip if key not pressed (0xEXA1)
	case digit1 == 0xE && digit3 == 0xA && digit4 == 1:
		x := digit2
		vx := e.vRegisters[x]
		key := e.keys[vx]

		if !key {
			e.programCounter += 2
		}
	// vx = dt (0xFX07)
	case digit1 == 0xF && digit3 == 0 && digit4 == 7:
		x := digit2
		e.vRegisters[x] = e.delayTimer
	// wait for key press (0xFX0A)
	case digit1 == 0xF && digit3 == 0 && digit4 == 0xA:
		x := digit2
		pressed := false

		for i, key := range e.keys {
			if key {
				e.vRegisters[x] = uint8(i)
				pressed = true
				break
			}
		}

		// go back if there was no key press
		if !pressed {
			e.programCounter -= 2
		}
	// dt = vx (0xFX15)
	case digit1 == 0xF && digit3 == 1 && digit4 == 5:
		x := digit2
		e.delayTimer = e.vRegisters[x]
	// i += vx (0xFX1E)
	case digit1 == 0xF && digit3 == 1 && digit4 == 0xE:
		x := digit2
		e.iRegister += uint16(e.vRegisters[x])
	// get font address for character stored at vx (0xFX29)
	case digit1 == 0xF && digit3 == 2 && digit4 == 9:
		x := digit2
		c := e.vRegisters[x]
		e.iRegister = uint16(c * 5)
	// binary-coded decimal value of vx (0xFX33)
	case digit1 == 0xF && digit3 == 3 && digit4 == 3:
		x := digit2
		vx := float32(e.vRegisters[x])

		hundreds := uint8(vx / 100.0)
		tens := uint8(vx/10.0) % 10
		ones := uint8(uint32(vx) % 10)

		e.ram[e.iRegister] = hundreds
		e.ram[e.iRegister+1] = tens
		e.ram[e.iRegister+2] = ones
	// store v0 to vx into ram starting from i register (0xFX55)
	case digit1 == 0xF && digit3 == 5 && digit4 == 5:
		x := digit2
		iReg := e.iRegister

		for i := range x + 1 {
			e.ram[iReg+i] = e.vRegisters[i]
		}
	// load i into v0 to vx (0xFX65)
	case digit1 == 0xF && digit3 == 6 && digit4 == 5:
		x := digit2
		iReg := e.iRegister

		for i := range x + 1 {
			e.vRegisters[i] = e.ram[iReg+i]
		}
	}
}

func (e *Emulator) popFromStack() uint16 {
	e.stackPointer--
	return e.stack[e.stackPointer]
}

func (e *Emulator) pushToStack(val uint16) {
	if e.stackPointer > STACK_SIZE {
		panic("stack overflow")
	}

	e.stack[e.stackPointer] = val
	e.stackPointer++
}
