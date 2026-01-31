package gui

import (
	"fmt"
	"math"
	"os"

	"github.com/0xmukesh/chippy/core"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const SCALE = 10
const TICKS_PER_FRAME = 20
const WINDOW_WIDTH = (core.SCREEN_WIDTH) * SCALE
const WINDOW_HEIGHT = (core.SCREEN_HEIGHT) * SCALE

var keyMap = map[int32]int{
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

func Start(file string) {
	emu := core.NewEmulator()
	rom, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read ROM file '%s': %v\n", file, err)
		os.Exit(1)
	}

	emu.LoadData(rom)

	rl.InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "chippy")
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	wave := generateBeepSound(440.0, 0.2, 44100)
	beep := rl.LoadSoundFromWave(wave)
	defer rl.UnloadSound(beep)

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		emu.TickTimers()
		handleKeyboardInput(emu)
		for range TICKS_PER_FRAME {
			emu.Tick()
		}

		if emu.SoundTimer() > 0 {
			if !rl.IsSoundPlaying(beep) {
				rl.PlaySound(beep)
			}
		} else {
			if rl.IsSoundPlaying(beep) {
				rl.StopSound(beep)
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
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
	for key, chip8Key := range keyMap {
		emu.Keypress(chip8Key, rl.IsKeyDown(key))
	}
}

func generateBeepSound(freq, duration float64, sampleRate uint32) rl.Wave {
	sampleCount := uint32(float64(sampleRate) * duration)

	samples := make([]float32, sampleCount)
	for i := range samples {
		t := float64(i) / float64(sampleRate)
		samples[i] = float32(0.5 * math.Sin(2*math.Pi*freq*t))
	}

	data := make([]byte, sampleCount*4)
	for i, sample := range samples {
		bits := math.Float32bits(sample)
		data[i*4] = byte(bits)
		data[i*4+1] = byte(bits >> 8)
		data[i*4+2] = byte(bits >> 16)
		data[i*4+3] = byte(bits >> 24)
	}

	return rl.NewWave(sampleCount, sampleRate, 32, 1, data)
}
