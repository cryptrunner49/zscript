package vm

import (
	"fmt"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/common"
	"github.com/cryptrunner49/gorex/internal/compiler"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/value"
)

const STACK_MAX = 256

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	chunk    *chunk.Chunk
	ip       int
	stack    [STACK_MAX]value.Value
	stackTop int
}

var vm VM

func InitVM() {
	resetStack()
}

func FreeVM() {}

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

func Interpret(source string) InterpretResult {
	ch := chunk.New()
	if !compiler.Compile(source, ch) {
		ch.Free()
		return INTERPRET_COMPILE_ERROR
	}
	vm.chunk = ch
	vm.ip = 0
	result := run()
	ch.Free()
	return result
}

func run() InterpretResult {
	readByte := func() uint8 {
		b := vm.chunk.Code()[vm.ip]
		vm.ip++
		return b
	}
	readConstant := func() value.Value {
		return vm.chunk.Constants().Values()[readByte()]
	}

	for {
		if common.DebugTraceExecution {
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
			Push(readConstant())
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
