package main

import (
	"fmt"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/value"
)

func main() {
	ch := chunk.New()

	constant := ch.AddConstant(value.Value(1.2))
	if err := ch.Write(chunk.OP_CONSTANT, 123); err != nil {
		fmt.Printf("Failed to write chunk: %v\n", err)
		return
	}
	if err := ch.Write(chunk.OpCode(constant), 123); err != nil {
		fmt.Printf("Failed to write chunk: %v\n", err)
		return
	}
	if err := ch.Write(chunk.OP_RETURN, 123); err != nil {
		fmt.Printf("Failed to write chunk: %v\n", err)
		return
	}

	debug.Disassemble(ch, "test chunk")
	ch.Free()
}
