package main

import (
	"flag"

	"github.com/0xmukesh/chippy/gui"
)

var (
	filePath string
)

func main() {
	flag.StringVar(&filePath, "file", "./roms/ibm.ch8", "path of chip-8 rom file")
	flag.Parse()

	gui.Start(filePath)
}
