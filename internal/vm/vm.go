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
	globals  map[*object.ObjString]value.Value
	strings  map[uint32]*object.ObjString
}

var vm VM

func InitVM() {
	resetStack()
	vm.objects = nil
	vm.globals = make(map[*object.ObjString]value.Value)
	vm.strings = make(map[uint32]*object.ObjString)
}

func FreeVM() {
	vm.globals = nil
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

func peek(distance int) value.Value {
	return vm.stack[vm.stackTop-1-distance]
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

func runtimeError(format string, args ...interface{}) InterpretResult {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
	instruction := vm.ip - 1
	line := vm.chunk.Lines()[instruction]
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
	resetStack()
	return INTERPRET_RUNTIME_ERROR
}

func concatenate() {
	b := Pop()
	a := Pop()
	astr := a.Obj.(*object.ObjString)
	bstr := b.Obj.(*object.ObjString)
	result := astr.Chars + bstr.Chars
	Push(value.Value{Type: value.VAL_OBJ, Obj: object.NewObjString(result)})
}

func crop() {
	b := Pop()
	a := Pop()
	astr := a.Obj.(*object.ObjString)
	bstr := b.Obj.(*object.ObjString)
	idx := strings.Index(astr.Chars, bstr.Chars)
	if idx >= 0 {
		newStr := astr.Chars[:idx] + astr.Chars[idx+len(bstr.Chars):]
		Push(value.Value{Type: value.VAL_OBJ, Obj: object.NewObjString(newStr)})
	} else {
		Push(a)
	}
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
	readString := func() *object.ObjString {
		return readConstant().Obj.(*object.ObjString)
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
		case uint8(chunk.OP_POP):
			Pop()
		case uint8(chunk.OP_DEFINE_GLOBAL):
			name := readString()
			vm.globals[name] = peek(0)
			Pop()
		case uint8(chunk.OP_SET_GLOBAL):
			name := readString()
			if _, exists := vm.globals[name]; !exists {
				return runtimeError("Undefined variable '%s'.", name.Chars)
			}
			vm.globals[name] = peek(0)
		case uint8(chunk.OP_GET_GLOBAL):
			name := readString()
			if val, exists := vm.globals[name]; exists {
				Push(val)
			} else {
				return runtimeError("Undefined variable '%s'.", name.Chars)
			}
		case uint8(chunk.OP_EQUAL):
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: value.Equal(a, b)})
		case uint8(chunk.OP_GREATER):
			if peek(0).Type != value.VAL_NUMBER || peek(1).Type != value.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: a.Number > b.Number})
		case uint8(chunk.OP_LESS):
			if peek(0).Type != value.VAL_NUMBER || peek(1).Type != value.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: a.Number < b.Number})
		case uint8(chunk.OP_ADD):
			if peek(0).Type == value.VAL_OBJ && peek(1).Type == value.VAL_OBJ {
				concatenate()
			} else if peek(0).Type == value.VAL_NUMBER && peek(1).Type == value.VAL_NUMBER {
				b := Pop()
				a := Pop()
				Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number + b.Number})
			} else {
				return runtimeError("Operands must be two numbers or two strings.")
			}
		case uint8(chunk.OP_SUBTRACT):
			if peek(0).Type == value.VAL_OBJ && peek(1).Type == value.VAL_OBJ {
				crop()
			} else if peek(0).Type == value.VAL_NUMBER && peek(1).Type == value.VAL_NUMBER {
				b := Pop()
				a := Pop()
				Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number - b.Number})
			} else {
				return runtimeError("Operands must be two numbers or two strings.")
			}
		case uint8(chunk.OP_MULTIPLY):
			if peek(0).Type != value.VAL_NUMBER || peek(1).Type != value.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number * b.Number})
		case uint8(chunk.OP_DIVIDE):
			if peek(0).Type != value.VAL_NUMBER || peek(1).Type != value.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: a.Number / b.Number})
		case uint8(chunk.OP_NOT):
			val := Pop()
			Push(value.Value{Type: value.VAL_BOOL, Bool: isFalsey(val)})
		case uint8(chunk.OP_NEGATE):
			if peek(0).Type != value.VAL_NUMBER {
				return runtimeError("Operand must be a number.")
			}
			val := Pop()
			Push(value.Value{Type: value.VAL_NUMBER, Number: -val.Number})
		case uint8(chunk.OP_PRINT):
			value.PrintValue(Pop())
			fmt.Println()
		case uint8(chunk.OP_RETURN):
			return INTERPRET_OK
		}
	}
}

func isFalsey(val value.Value) bool {
	return val.Type == value.VAL_NULL || (val.Type == value.VAL_BOOL && !val.Bool)
}
