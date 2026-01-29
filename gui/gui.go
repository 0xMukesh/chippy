package gui

import (
	"fmt"
	"os"

	"github.com/0xmukesh/chippy/core"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const SCALE = 25
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

		for range 10 {
			emu.Tick()
		}
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
