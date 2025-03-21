package main

import (
	"fmt"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/debug"
)

func main() {
	ch := chunk.New()
	if err := ch.Write(chunk.OP_RETURN); err != nil {
		fmt.Printf("Failed to write chunk: %v\n", err)
		return
	}

	debug.Disassemble(ch, "test chunk")
	ch.Free()
}
