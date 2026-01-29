package core

import (
	"fmt"
	"testing"
)

func TestEmulator(t *testing.T) {
	fmt.Printf("%x\n", 0xABCD&0x0FF)
}
