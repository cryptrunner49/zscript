package vm

import (
	"fmt"
	"os"
	"strings"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/common"
	"github.com/cryptrunner49/gorex/internal/compiler"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/object"
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
	objects  *object.Obj
	strings  map[uint32]*object.ObjString
}

var vm VM

func InitVM() {
	resetStack()
	vm.objects = nil
	vm.strings = make(map[uint32]*object.ObjString)
}

func FreeVM() {
	vm.strings = nil
	vm.objects = nil
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
			Push(value.Value{Type: value.VAL_BOOL, Bool: value.Equal(a, b)})
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
			if a.Type == value.VAL_OBJ && b.Type == value.VAL_OBJ {
				astr, okA := a.Obj.(*object.ObjString)
				bstr, okB := b.Obj.(*object.ObjString)
				if okA && okB {
					result := astr.Chars + bstr.Chars
					Push(value.Value{Type: value.VAL_OBJ, Obj: object.NewObjString(result)})
					continue
				}
			}
			if a.Type == value.VAL_NUMBER && b.Type == value.VAL_NUMBER {
				Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number + b.Number})
			} else {
				fmt.Fprintln(os.Stderr, "Operands must be two numbers or two strings.")
				return INTERPRET_RUNTIME_ERROR
			}
		case uint8(chunk.OP_SUBTRACT):
			b := Pop()
			a := Pop()
			if a.Type == value.VAL_OBJ && b.Type == value.VAL_OBJ {
				astr, okA := a.Obj.(*object.ObjString)
				bstr, okB := b.Obj.(*object.ObjString)
				if okA && okB {
					idx := strings.Index(astr.Chars, bstr.Chars)
					if idx >= 0 {
						newStr := astr.Chars[:idx] + astr.Chars[idx+len(bstr.Chars):]
						Push(value.Value{Type: value.VAL_OBJ, Obj: object.NewObjString(newStr)})
					} else {
						Push(a)
					}
					continue
				}
			}
			if a.Type == value.VAL_NUMBER && b.Type == value.VAL_NUMBER {
				Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number - b.Number})
			} else {
				fmt.Fprintln(os.Stderr, "Operands must be two numbers or two strings.")
				return INTERPRET_RUNTIME_ERROR
			}
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

func isFalsey(val value.Value) bool {
	return val.Type == value.VAL_NULL || (val.Type == value.VAL_BOOL && !val.Bool)
}
