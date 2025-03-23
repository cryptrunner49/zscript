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
		case uint8(chunk.OP_NULL):
			Push(value.Value{Type: value.VAL_NULL})
		case uint8(chunk.OP_TRUE):
			Push(value.Value{Type: value.VAL_BOOL, Bool: true})
		case uint8(chunk.OP_FALSE):
			Push(value.Value{Type: value.VAL_BOOL, Bool: false})
		case uint8(chunk.OP_EQUAL):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: valuesEqual(a, b)})
		case uint8(chunk.OP_GREATER):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: a.Number > b.Number})
		case uint8(chunk.OP_LESS):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: a.Number < b.Number})
		case uint8(chunk.OP_ADD):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number + b.Number})
		case uint8(chunk.OP_SUBTRACT):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number - b.Number})
		case uint8(chunk.OP_MULTIPLY):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number * b.Number})
		case uint8(chunk.OP_DIVIDE):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number / b.Number})
		case uint8(chunk.OP_NOT):
			val := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: isFalsey(val)})
		case uint8(chunk.OP_NEGATE):
			val := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: -val.Number})
		case uint8(chunk.OP_RETURN):
			value.PrintValue(Pop())
			fmt.Println()
			return INTERPRET_OK
		}
	}
}

func valuesEqual(a, b value.Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case value.VAL_BOOL:
		return a.Bool == b.Bool
	case value.VAL_NULL:
		return true
	case value.VAL_NUMBER:
		return a.Number == b.Number
	default:
		return false
	}
}

func isFalsey(val value.Value) bool {
	return val.Type == value.VAL_NULL || (val.Type == value.VAL_BOOL && !val.Bool)
}
