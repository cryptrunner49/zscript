package vm

import (
	"fmt"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/value"
)

const (
	STACK_MAX = 256
)

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	chunk    *chunk.Chunk
	ip       int // Using int as index instead of pointer
	stack    [STACK_MAX]value.Value
	stackTop int
}

var vm VM

func InitVM() {
	resetStack()
}

func FreeVM() {
	// No-op in Go version since memory is managed automatically
}

func resetStack() {
	vm.stackTop = 0
}

func Push(val value.Value) {
	vm.stack[vm.stackTop] = val
	vm.stackTop++
}

func Pop() value.Value {
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func Interpret(ch *chunk.Chunk) InterpretResult {
	vm.chunk = ch
	vm.ip = 0
	return run()
}

func run() InterpretResult {
	const debugTraceExecution = true // Matches C's DEBUG_TRACE_EXECUTION

	readByte := func() uint8 {
		b := vm.chunk.Code()[vm.ip]
		vm.ip++
		return b
	}
	readConstant := func() value.Value {
		return vm.chunk.Constants().Values()[readByte()]
	}

	for {
		if debugTraceExecution {
			fmt.Print("      ")
			for i := 0; i < vm.stackTop; i++ {
				fmt.Print("[ ")
				value.PrintValue(vm.stack[i])
				fmt.Print(" ]")
			}
			fmt.Println()
			debug.DisassembleInstruction(vm.chunk, vm.ip)
		}

		instruction := readByte()
		switch instruction {
		case uint8(chunk.OP_CONSTANT):
			constant := readConstant()
			Push(constant)
		case uint8(chunk.OP_ADD):
			b := Pop()
			a := Pop()
			Push(a + b)
		case uint8(chunk.OP_SUBTRACT):
			b := Pop()
			a := Pop()
			Push(a - b)
		case uint8(chunk.OP_MULTIPLY):
			b := Pop()
			a := Pop()
			Push(a * b)
		case uint8(chunk.OP_DIVIDE):
			b := Pop()
			a := Pop()
			Push(a / b)
		case uint8(chunk.OP_NEGATE):
			Push(-Pop())
		case uint8(chunk.OP_RETURN):
			value.PrintValue(Pop())
			fmt.Println()
			return INTERPRET_OK
		}
	}
}
