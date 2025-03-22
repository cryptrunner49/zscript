package main

import (
	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/value"
	"github.com/cryptrunner49/gorex/internal/vm"
)

func main() {
	vm.InitVM()

	ch := chunk.New()

	constant := ch.AddConstant(value.Value(1.2))
	ch.Write(uint8(chunk.OP_CONSTANT), 123)
	ch.Write(uint8(constant), 123)

	constant = ch.AddConstant(value.Value(3.4))
	ch.Write(uint8(chunk.OP_CONSTANT), 123)
	ch.Write(uint8(constant), 123)

	ch.Write(uint8(chunk.OP_ADD), 123)

	constant = ch.AddConstant(value.Value(5.6))
	ch.Write(uint8(chunk.OP_CONSTANT), 123)
	ch.Write(uint8(constant), 123)

	ch.Write(uint8(chunk.OP_DIVIDE), 123)

	constant = ch.AddConstant(value.Value(2))
	ch.Write(uint8(chunk.OP_CONSTANT), 123)
	ch.Write(uint8(constant), 123)

	ch.Write(uint8(chunk.OP_MULTIPLY), 123)
	ch.Write(uint8(chunk.OP_NEGATE), 123)

	ch.Write(uint8(chunk.OP_RETURN), 123)

	debug.Disassemble(ch, "test chunk")
	vm.Interpret(ch)

	vm.FreeVM()
	ch.Free()
}
