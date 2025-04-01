package vm

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/cryptrunner49/goseedvm/internal/common"
	"github.com/cryptrunner49/goseedvm/internal/compiler"
	"github.com/cryptrunner49/goseedvm/internal/debug"
	"github.com/cryptrunner49/goseedvm/internal/runtime"
)

const (
	FRAMES_MAX = 64
	STACK_MAX  = FRAMES_MAX * 256
)

type CallFrame struct {
	closure *runtime.ObjClosure
	ip      int
	slots   int
}

type InterpretResult int

const (
	INTERPRET_OK InterpretResult = iota
	INTERPRET_COMPILE_ERROR
	INTERPRET_RUNTIME_ERROR
)

type VM struct {
	frames       [FRAMES_MAX]CallFrame
	frameCount   int
	stack        [STACK_MAX]runtime.Value
	stackTop     int
	objects      *runtime.Obj
	globals      map[*runtime.ObjString]runtime.Value
	strings      map[uint32]*runtime.ObjString
	openUpvalues *runtime.ObjUpvalue
}

var vm VM

func InitVM(args []string) {
	resetStack()
	vm.objects = nil
	vm.globals = make(map[*runtime.ObjString]runtime.Value)
	vm.strings = make(map[uint32]*runtime.ObjString)

	// Define built-in globals, including command-line arguments
	defineAllNatives()
	defineArgs(args)
}

func FreeVM() {
	vm.globals = nil
	vm.strings = nil
	vm.objects = nil
}

func resetStack() {
	vm.stackTop = 0
	vm.frameCount = 0
	vm.openUpvalues = nil
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
	resetStack()
	function := compiler.Compile(source)
	if function == nil {
		return INTERPRET_COMPILE_ERROR
	}
	closure := runtime.NewClosure(function)
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: closure})
	callValue(runtime.Value{Type: runtime.VAL_OBJ, Obj: closure}, 0)
	return run()
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

func captureUpvalue(local *runtime.Value) *runtime.ObjUpvalue {
	var prevUpvalue *runtime.ObjUpvalue
	upvalue := vm.openUpvalues
	for upvalue != nil && uintptr(unsafe.Pointer(upvalue.Location)) > uintptr(unsafe.Pointer(local)) {
		prevUpvalue = upvalue
		upvalue = upvalue.Next
	}
	if upvalue != nil && upvalue.Location == local {
		return upvalue
	}
	createdUpvalue := runtime.NewUpvalue(local)
	createdUpvalue.Next = upvalue
	if prevUpvalue == nil {
		vm.openUpvalues = createdUpvalue
	} else {
		prevUpvalue.Next = createdUpvalue
	}
	return createdUpvalue
}

func closeUpvalues(last *runtime.Value) {
	for vm.openUpvalues != nil && uintptr(unsafe.Pointer(vm.openUpvalues.Location)) >= uintptr(unsafe.Pointer(last)) {
		upvalue := vm.openUpvalues
		upvalue.Closed = *upvalue.Location
		upvalue.Location = &upvalue.Closed
		vm.openUpvalues = upvalue.Next
	}
}

func run() InterpretResult {
	readByte := func(frame *CallFrame) uint8 {
		b := frame.closure.Function.Chunk.Code()[frame.ip]
		frame.ip++
		return b
	}
	readShort := func(frame *CallFrame) int {
		b1 := frame.closure.Function.Chunk.Code()[frame.ip]
		b2 := frame.closure.Function.Chunk.Code()[frame.ip+1]
		frame.ip += 2
		return int(b1)<<8 | int(b2)
	}
	readConstant := func(frame *CallFrame) runtime.Value {
		return frame.closure.Function.Chunk.Constants().Values()[readByte(frame)]
	}
	readString := func(frame *CallFrame) *runtime.ObjString {
		return readConstant(frame).Obj.(*runtime.ObjString)
	}
	binaryOp := func(op func(a, b float64) float64, opName string) InterpretResult {
		top := peek(0)
		next := peek(1)
		if top.Type != runtime.VAL_NUMBER || next.Type != runtime.VAL_NUMBER {
			return runtimeError("Both operands for '%s' must be numbers (got %s and %s).", opName, typeName(top), typeName(next))
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
			debug.DisassembleInstruction(&frame.closure.Function.Chunk, frame.ip)
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
				return runtimeError("Cannot assign to undefined global variable '%s'.", name.Chars)
			}
			vm.globals[name] = peek(0)
		case uint8(runtime.OP_GET_GLOBAL):
			name := readString(frame)
			if val, exists := vm.globals[name]; exists {
				Push(val)
			} else {
				return runtimeError("Global variable '%s' is not defined.", name.Chars)
			}
		case uint8(runtime.OP_GET_UPVALUE):
			slot := readByte(frame)
			upvalue := frame.closure.Upvalues[slot]
			Push(*upvalue.Location)
		case uint8(runtime.OP_SET_UPVALUE):
			slot := readByte(frame)
			upvalue := frame.closure.Upvalues[slot]
			*upvalue.Location = peek(0)
		case uint8(runtime.OP_GET_PROPERTY):
			instVal := peek(0)
			if _, ok := instVal.Obj.(*runtime.ObjInstance); !ok {
				return runtimeError("Cannot access property on %s; only struct instances have properties.", typeName(instVal))
			}
			instance := instVal.Obj.(*runtime.ObjInstance)
			name := readString(frame)
			if value, found := instance.Fields[name]; found {
				Pop()
				Push(value)
			} else {
				return runtimeError("Property '%s' does not exist on this instance.", name.Chars)
			}
		case uint8(runtime.OP_SET_PROPERTY):
			instVal := peek(1)
			if _, ok := instVal.Obj.(*runtime.ObjInstance); !ok {
				return runtimeError("Cannot set property on %s; only struct instances have fields.", typeName(instVal))
			}
			instance := instVal.Obj.(*runtime.ObjInstance)
			name := readString(frame)
			instance.Fields[name] = peek(0)
			value := Pop()
			Pop()
			Push(value)
		case uint8(runtime.OP_EQUAL):
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: runtime.Equal(a, b)})
		case uint8(runtime.OP_GREATER):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Both operands for '>' must be numbers (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: a.Number > b.Number})
		case uint8(runtime.OP_LESS):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Both operands for '<' must be numbers (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: a.Number < b.Number})
		case uint8(runtime.OP_ADD):
			if peek(0).Type == runtime.VAL_OBJ && peek(1).Type == runtime.VAL_OBJ {
				concatenate()
			} else if peek(0).Type == runtime.VAL_NUMBER && peek(1).Type == runtime.VAL_NUMBER {
				binaryOp(func(a, b float64) float64 { return a + b }, "+")
			} else {
				return runtimeError("Operator '+' requires two numbers or two strings (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
		case uint8(runtime.OP_SUBTRACT):
			if peek(0).Type == runtime.VAL_OBJ && peek(1).Type == runtime.VAL_OBJ {
				crop()
			} else if peek(0).Type == runtime.VAL_NUMBER && peek(1).Type == runtime.VAL_NUMBER {
				b := Pop()
				a := Pop()
				Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number - b.Number})
			} else {
				return runtimeError("Operator '-' requires two numbers or two strings (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
		case uint8(runtime.OP_MULTIPLY):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Both operands for '*' must be numbers (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number * b.Number})
		case uint8(runtime.OP_DIVIDE):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Both operands for '/' must be numbers (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
			b := Pop()
			if b.Number == 0 {
				return runtimeError("Division by zero is not allowed.")
			}
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number / b.Number})
		case uint8(runtime.OP_MOD):
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Both operands for '%%' must be numbers (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
			b := Pop()
			if b.Number == 0 {
				return runtimeError("Modulo by zero is not allowed.")
			}
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: math.Mod(a.Number, b.Number)})
		case uint8(runtime.OP_NOT):
			val := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: isFalsey(val)})
		case uint8(runtime.OP_NEGATE):
			if peek(0).Type != runtime.VAL_NUMBER {
				return runtimeError("Unary '-' requires a number (got %s).", typeName(peek(0)))
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
		case uint8(runtime.OP_BREAK):
			offset := int(readShort(frame))
			frame.ip += offset
		case uint8(runtime.OP_CONTINUE):
			offset := int(readShort(frame))
			frame.ip -= offset
		case uint8(runtime.OP_CALL):
			argCount := int(readByte(frame))
			if !callValue(peek(argCount), argCount) {
				return INTERPRET_RUNTIME_ERROR
			}
		case uint8(runtime.OP_CLOSURE):
			function := readConstant(frame).Obj.(*runtime.ObjFunction)
			closure := runtime.NewClosure(function)
			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: closure})
			for i := 0; i < closure.UpvalueCount; i++ {
				isLocal := readByte(frame)
				index := readByte(frame)
				if isLocal != 0 {
					closure.Upvalues[i] = captureUpvalue(&vm.stack[frame.slots+int(index)])
				} else {
					closure.Upvalues[i] = frame.closure.Upvalues[index]
				}
			}
		case uint8(runtime.OP_CLOSE_UPVALUE):
			closeUpvalues(&vm.stack[vm.stackTop-1])
			Pop()
		case uint8(runtime.OP_RETURN):
			result := Pop()
			closeUpvalues(&vm.stack[frame.slots])
			vm.frameCount--
			if vm.frameCount == 0 {
				Pop()
				return INTERPRET_OK
			} else {
				vm.stackTop = frame.slots
				Push(result)
				frame = &vm.frames[vm.frameCount-1]
			}
		case uint8(runtime.OP_STRUCT):
			name := readString(frame)
			objStruct := runtime.NewStruct(name)
			fieldCount := int(readByte(frame))
			for i := 0; i < fieldCount; i++ {
				fieldName := readConstant(frame).Obj.(*runtime.ObjString)
				defaultValue := readConstant(frame)
				objStruct.Fields[fieldName] = defaultValue
			}
			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: objStruct})
		}
	}
}

func callValue(callee runtime.Value, argCount int) bool {
	if callee.Type == runtime.VAL_OBJ {
		switch obj := callee.Obj.(type) {
		case *runtime.ObjClosure:
			return call(obj, argCount)
		case *runtime.ObjNative:
			native := obj.Function
			result := native(argCount, vm.stack[vm.stackTop-argCount:vm.stackTop])
			vm.stackTop -= argCount + 1
			Push(result)
			return true
		case *runtime.ObjStruct:
			vm.stack[vm.stackTop-argCount-1] = runtime.ObjVal(runtime.NewInstance(obj))
			return true
		default:
			// Non-callable object type
		}
	}
	runtimeError("Cannot call %s; only functions and structs are callable.", typeName(callee))
	return false
}

func call(closure *runtime.ObjClosure, argCount int) bool {
	if argCount != closure.Function.Arity {
		runtimeError("Function '%s' expects %d arguments but got %d.", closure.Function.Name.Chars, closure.Function.Arity, argCount)
		return false
	}
	if vm.frameCount >= FRAMES_MAX {
		runtimeError("Stack overflow; too many nested function calls (max %d).", FRAMES_MAX)
		return false
	}
	frame := &vm.frames[vm.frameCount]
	vm.frameCount++
	frame.closure = closure
	frame.ip = 0
	frame.slots = vm.stackTop - argCount - 1
	return true
}
