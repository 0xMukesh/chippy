package gui

import (
	"fmt"
	"os"

	"github.com/0xmukesh/chippy/core"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const SCALE = 10
const TICKS_PER_FRAME = 10
const WINDOW_WIDTH = (core.SCREEN_WIDTH) * SCALE
const WINDOW_HEIGHT = (core.SCREEN_HEIGHT) * SCALE

func Start(file string) {
	emu := core.NewEmulator()
	rom, err := os.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("failed to read %s file - %s", file, err.Error()))
	}

	emu.LoadData(rom)

	rl.InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "chippy")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		handleKeyboardInput(emu)

		for range TICKS_PER_FRAME {
			emu.Tick()
		}

		emu.TickTimers()
		drawScreen(emu)
		rl.EndDrawing()
	}
}

func drawScreen(emu *core.Emulator) {
	screen := emu.GetDisplay()

	for i, pixel := range screen {
		x := int32(i%core.SCREEN_WIDTH) * SCALE
		y := int32(i/core.SCREEN_WIDTH) * SCALE

		if pixel {
			rl.DrawRectangle(x, y, SCALE, SCALE, rl.White)
		}
	}
}

func handleKeyboardInput(emu *core.Emulator) {
	keyMap := map[int32]int{
		rl.KeyOne:   0x1,
		rl.KeyTwo:   0x2,
		rl.KeyThree: 0x3,
		rl.KeyFour:  0xC,
		rl.KeyQ:     0x4,
		rl.KeyW:     0x5,
		rl.KeyE:     0x6,
		rl.KeyR:     0xD,
		rl.KeyA:     0x7,
		rl.KeyS:     0x8,
		rl.KeyD:     0x9,
		rl.KeyF:     0xE,
		rl.KeyZ:     0xA,
		rl.KeyX:     0x0,
		rl.KeyC:     0xB,
		rl.KeyV:     0xF,
	}

	for key, chip8Key := range keyMap {
		emu.Keypress(chip8Key, rl.IsKeyDown(key))
	}
}
