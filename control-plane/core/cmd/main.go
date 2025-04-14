package main

import (
	"log"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/internal/process"
)

func main() {
	proc := process.NewProcess()

	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("❌ Fatal error: %v", r)
		}
	}()

	proc.Launch()
}
