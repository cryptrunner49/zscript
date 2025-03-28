package vm

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cryptrunner49/gorex/internal/common"
	"github.com/cryptrunner49/gorex/internal/compiler"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/runtime"
)

const (
	FRAMES_MAX = 64
	STACK_MAX  = FRAMES_MAX * 256
)

type CallFrame struct {
	function *runtime.ObjFunction
	ip       int
	slots    int
}

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	frames     [FRAMES_MAX]CallFrame
	frameCount int
	chunk      *runtime.Chunk
	ip         int
	stack      [STACK_MAX]runtime.Value
	stackTop   int
	objects    *runtime.Obj
	globals    map[*runtime.ObjString]runtime.Value
	strings    map[uint32]*runtime.ObjString
}

var vm VM

func InitVM() {
	resetStack()
	vm.objects = nil
	vm.globals = make(map[*runtime.ObjString]runtime.Value)
	vm.strings = make(map[uint32]*runtime.ObjString)

	defineNative("clock", clockNative)
}

func FreeVM() {
	vm.globals = nil
	vm.strings = nil
	vm.objects = nil
}

func resetStack() {
	vm.stackTop = 0
	vm.frameCount = 0
}

func Push(val runtime.Value) {
	vm.stack[vm.stackTop] = val
	vm.stackTop++
}

func Pop() runtime.Value {
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func peek(distance int) runtime.Value {
	return vm.stack[vm.stackTop-1-distance]
}

func Interpret(source string) InterpretResult {
	resetStack() // Reset stack and frame count before interpretation

	function := compiler.Compile(source)
	if function == nil {
		return INTERPRET_COMPILE_ERROR
	}

	// Push the function onto the stack as an object.
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: function})
	callValue(runtime.Value{Type: runtime.VAL_OBJ, Obj: function}, 0)

	return run()
}

func runtimeError(format string, args ...interface{}) InterpretResult {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)

	for i := vm.frameCount - 1; i >= 0; i-- {
		frame := &vm.frames[i]
		function := frame.function

		// -1 because the IP is sitting on the next instruction to be executed.
		instruction := frame.ip - 1
		line := function.Chunk.Lines()[instruction]

		fmt.Fprintf(os.Stderr, "[line %d] in ", line)

		if function.Name == nil {
			fmt.Fprintln(os.Stderr, "script")
		} else {
			fmt.Fprintf(os.Stderr, "%s()\n", function.Name.Chars)
		}
	}

	resetStack()
	return INTERPRET_RUNTIME_ERROR
}

func concatenate() {
	b := Pop()
	a := Pop()
	astr := a.Obj.(*runtime.ObjString)
	bstr := b.Obj.(*runtime.ObjString)
	result := astr.Chars + bstr.Chars
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(result)})
}

func crop() {
	b := Pop()
	a := Pop()
	astr := a.Obj.(*runtime.ObjString)
	bstr := b.Obj.(*runtime.ObjString)
	idx := strings.Index(astr.Chars, bstr.Chars)
	if idx >= 0 {
		newStr := astr.Chars[:idx] + astr.Chars[idx+len(bstr.Chars):]
		Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(newStr)})
	} else {
		Push(a)
	}
}

func run() InterpretResult {
	// Helper functions to read instructions (adjusted to take frame as parameter)
	readByte := func(frame *CallFrame) uint8 {
		b := frame.function.Chunk.Code()[frame.ip]
		frame.ip++
		return b
	}
	readShort := func(frame *CallFrame) int {
		b1 := frame.function.Chunk.Code()[frame.ip]
		b2 := frame.function.Chunk.Code()[frame.ip+1]
		frame.ip += 2
		return int(b1)<<8 | int(b2)
	}
	readConstant := func(frame *CallFrame) runtime.Value {
		return frame.function.Chunk.Constants().Values()[readByte(frame)]
	}
	readString := func(frame *CallFrame) *runtime.ObjString {
		return readConstant(frame).Obj.(*runtime.ObjString)
	}
	binaryOp := func(op func(a, b float64) float64) InterpretResult {
		top := peek(0)
		next := peek(1)
		if top.Type != runtime.VAL_NUMBER || next.Type != runtime.VAL_NUMBER {
			runtimeError("Operands must be numbers.")
			return INTERPRET_RUNTIME_ERROR
		}
		b := Pop().Number
		a := Pop().Number
		result := op(a, b)
		Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: result})
		return INTERPRET_OK
	}

	for {
		if vm.frameCount == 0 {
			return INTERPRET_OK
		}
		frame := &vm.frames[vm.frameCount-1]

		if common.DebugTraceExecution {
			fmt.Print("      ")
			for i := 0; i < vm.stackTop; i++ {
				fmt.Print("[ ")
				runtime.PrintValue(vm.stack[i])
				fmt.Print(" ]")
			}
			fmt.Println()
			debug.DisassembleInstruction(&frame.function.Chunk, frame.ip)
		}

		instruction := readByte(frame)
		switch instruction {
		case uint8(runtime.OP_CONSTANT):
			Push(readConstant(frame))
		case uint8(runtime.OP_NULL):
			Push(runtime.Value{Type: runtime.VAL_NULL})
		case uint8(runtime.OP_TRUE):
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: true})
		case uint8(runtime.OP_FALSE):
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: false})
		case uint8(runtime.OP_POP):
			Pop()
		case uint8(runtime.OP_SET_LOCAL):
			slot := readByte(frame)
			vm.stack[frame.slots+int(slot)] = peek(0)
		case uint8(runtime.OP_GET_LOCAL):
			slot := readByte(frame)
			Push(vm.stack[frame.slots+int(slot)])
		case uint8(runtime.OP_DEFINE_GLOBAL):
			name := readString(frame)
			vm.globals[name] = peek(0)
			Pop()
		case uint8(runtime.OP_SET_GLOBAL):
			name := readString(frame)
			if _, exists := vm.globals[name]; !exists {
				return runtimeError("Undefined variable '%s'.", name.Chars)
			}
			vm.globals[name] = peek(0)
		case uint8(runtime.OP_GET_GLOBAL):
			name := readString(frame)
			if val, exists := vm.globals[name]; exists {
				Push(val)
			} else {
				return runtimeError("Undefined variable '%s'.", name.Chars)
			}
		case uint8(runtime.OP_EQUAL):
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: runtime.Equal(a, b)})
		case uint8(runtime.OP_GREATER):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: a.Number > b.Number})
		case uint8(runtime.OP_LESS):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: a.Number < b.Number})
		case uint8(runtime.OP_ADD):
			if peek(0).Type == runtime.VAL_OBJ && peek(1).Type == runtime.VAL_OBJ {
				concatenate()
			} else if peek(0).Type == runtime.VAL_NUMBER && peek(1).Type == runtime.VAL_NUMBER {
				binaryOp(func(a, b float64) float64 { return a + b })
			} else {
				return runtimeError("Operands must be two numbers or two strings.")
			}
		case uint8(runtime.OP_SUBTRACT):
			if peek(0).Type == runtime.VAL_OBJ && peek(1).Type == runtime.VAL_OBJ {
				crop()
			} else if peek(0).Type == runtime.VAL_NUMBER && peek(1).Type == runtime.VAL_NUMBER {
				b := Pop()
				a := Pop()
				Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number - b.Number})
			} else {
				return runtimeError("Operands must be two numbers or two strings.")
			}
		case uint8(runtime.OP_MULTIPLY):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number * b.Number})
		case uint8(runtime.OP_DIVIDE):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Operands must be numbers.")
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number / b.Number})
		case uint8(runtime.OP_NOT):
			val := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: isFalsey(val)})
		case uint8(runtime.OP_NEGATE):
			if peek(0).Type != runtime.VAL_NUMBER {
				return runtimeError("Operand must be a number.")
			}
			val := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: -val.Number})
		case uint8(runtime.OP_PRINT):
			runtime.PrintValue(Pop())
			fmt.Println()
		case uint8(runtime.OP_JUMP):
			offset := int(readShort(frame))
			frame.ip += offset
		case uint8(runtime.OP_JUMP_IF_FALSE):
			offset := int(readShort(frame))
			if isFalsey(peek(0)) {
				frame.ip += offset
			}
		case uint8(runtime.OP_LOOP):
			offset := int(readShort(frame))
			frame.ip -= offset
		case uint8(runtime.OP_CALL):
			argCount := int(readByte(frame))
			if !callValue(peek(argCount), argCount) {
				return INTERPRET_RUNTIME_ERROR
			}
		case uint8(runtime.OP_RETURN):
			result := Pop()
			vm.frameCount--

			if vm.frameCount == 0 {
				Pop()
				return INTERPRET_OK
			} else {
				vm.stackTop = frame.slots
				Push(result)

				frame = &vm.frames[vm.frameCount]
			}
		}
	}
}

func isFalsey(val runtime.Value) bool {
	return val.Type == runtime.VAL_NULL || (val.Type == runtime.VAL_BOOL && !val.Bool)
}

func callValue(callee runtime.Value, argCount int) bool {
	if callee.Type == runtime.VAL_OBJ {
		switch obj := callee.Obj.(type) {
		case *runtime.ObjFunction:
			return call(obj, argCount)
		case *runtime.ObjNative:
			native := callee.Obj.(*runtime.ObjNative).Function
			result := native(argCount, vm.stack[vm.stackTop-argCount:])
			vm.stackTop -= argCount + 1
			Push(result)
			return true
		default:
			// Non-callable object type.
		}
	}
	runtimeError("Can only call functions and classes.")
	return false
}

func call(function *runtime.ObjFunction, argCount int) bool {
	if argCount != function.Arity {
		runtimeError("Expected %d arguments but got %d", function.Arity, argCount)
		return false
	}
	if vm.frameCount >= FRAMES_MAX {
		runtimeError("Stack overflow.")
		return false
	}

	frame := &vm.frames[vm.frameCount]
	vm.frameCount++

	frame.function = function
	frame.ip = 0 // Assuming `ip` should start at 0 for new functions
	frame.slots = vm.stackTop - argCount - 1

	return true
}

func defineNative(name string, function runtime.NativeFn) {
	// Create a new ObjString for the function name.
	nameObj := runtime.NewObjString(name)
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: nameObj})

	// Create a new ObjNative function object.
	nativeObj := &runtime.ObjNative{Function: function}
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: nativeObj})

	// Store the function in the global variables table.
	vm.globals[nameObj] = vm.stack[vm.stackTop-1]

	// Pop the function and its name from the stack.
	Pop()
	Pop()
}

func clockNative(argCount int, args []runtime.Value) runtime.Value {
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(time.Now().UnixNano()) / 1e9, // Convert nanoseconds to seconds
	}
}
