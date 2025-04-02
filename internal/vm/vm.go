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

// Maximum number of call frames allowed and the maximum stack size.
const (
	FRAMES_MAX = 64               // Maximum number of call frames (for nested function calls).
	STACK_MAX  = FRAMES_MAX * 256 // Maximum number of values on the stack.
)

// CallFrame represents an active function call.
type CallFrame struct {
	closure *runtime.ObjClosure // The closure (function with environment) being executed.
	ip      int                 // Instruction pointer into the function's bytecode.
	slots   int                 // Base index in the VM's stack where this call's local variables begin.
}

// InterpretResult indicates the outcome of interpreting code.
type InterpretResult int

const (
	INTERPRET_OK            InterpretResult = iota // Code executed successfully.
	INTERPRET_COMPILE_ERROR                        // A compile-time error occurred.
	INTERPRET_RUNTIME_ERROR                        // A runtime error occurred.
)

// VM represents the virtual machine state.
type VM struct {
	frames       [FRAMES_MAX]CallFrame                // Call frame stack for function calls.
	frameCount   int                                  // Number of active call frames.
	stack        [STACK_MAX]runtime.Value             // Value stack used during execution.
	stackTop     int                                  // Index of the next available slot on the stack.
	objects      *runtime.Obj                         // Linked list of all allocated objects.
	globals      map[*runtime.ObjString]runtime.Value // Global variables table.
	strings      map[uint32]*runtime.ObjString        // Interned strings table.
	openUpvalues *runtime.ObjUpvalue                  // Linked list of open upvalues for closures.
}

var vm VM // Global VM instance.

// InitVM initializes the virtual machine, sets up the stack and built-in globals,
// and processes command-line arguments.
func InitVM(args []string) {
	resetStack()
	vm.objects = nil
	vm.globals = make(map[*runtime.ObjString]runtime.Value)
	vm.strings = make(map[uint32]*runtime.ObjString)

	// Define built-in native functions and globals, including command-line arguments.
	defineAllNatives()
	defineArgs(args)
}

// FreeVM frees resources used by the VM.
func FreeVM() {
	vm.globals = nil
	vm.strings = nil
	vm.objects = nil
}

// resetStack resets the VM's stack, call frame count, and open upvalues.
func resetStack() {
	vm.stackTop = 0
	vm.frameCount = 0
	vm.openUpvalues = nil
}

// Push pushes a value onto the VM's stack.
func Push(val runtime.Value) {
	vm.stack[vm.stackTop] = val
	vm.stackTop++
}

// Pop removes and returns the top value from the VM's stack.
func Pop() runtime.Value {
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

// peek returns a value from the stack without removing it. Distance 0 is the top.
func peek(distance int) runtime.Value {
	return vm.stack[vm.stackTop-1-distance]
}

// Interpret compiles the source code and executes it in the VM.
// It returns an interpretation result indicating success or type of error.
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

// concatenate concatenates the top two string objects from the stack.
func concatenate() {
	b := Pop()
	a := Pop()
	astr := a.Obj.(*runtime.ObjString)
	bstr := b.Obj.(*runtime.ObjString)
	result := astr.Chars + bstr.Chars
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(result)})
}

// crop removes the first occurrence of one string (b) from another (a).
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

// captureUpvalue creates or reuses an upvalue that points to a local variable.
func captureUpvalue(local *runtime.Value) *runtime.ObjUpvalue {
	var prevUpvalue *runtime.ObjUpvalue
	upvalue := vm.openUpvalues
	// Walk the open upvalues list; the list is sorted by pointer addresses.
	for upvalue != nil && uintptr(unsafe.Pointer(upvalue.Location)) > uintptr(unsafe.Pointer(local)) {
		prevUpvalue = upvalue
		upvalue = upvalue.Next
	}
	// Reuse existing upvalue if already capturing this variable.
	if upvalue != nil && upvalue.Location == local {
		return upvalue
	}
	// Create a new upvalue.
	createdUpvalue := runtime.NewUpvalue(local)
	createdUpvalue.Next = upvalue
	if prevUpvalue == nil {
		vm.openUpvalues = createdUpvalue
	} else {
		prevUpvalue.Next = createdUpvalue
	}
	return createdUpvalue
}

// closeUpvalues closes all upvalues that refer to variables at or above the given stack slot.
func closeUpvalues(last *runtime.Value) {
	for vm.openUpvalues != nil && uintptr(unsafe.Pointer(vm.openUpvalues.Location)) >= uintptr(unsafe.Pointer(last)) {
		upvalue := vm.openUpvalues
		upvalue.Closed = *upvalue.Location
		upvalue.Location = &upvalue.Closed
		vm.openUpvalues = upvalue.Next
	}
}

// run executes the bytecode instructions in a loop and returns an interpretation result.
func run() InterpretResult {
	// Helper functions to read bytes and constants from the current call frame.
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
	// binaryOp applies a binary arithmetic operation to two number operands.
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

	// Main instruction dispatch loop.
	for {
		// When no more call frames remain, execution is complete.
		if vm.frameCount == 0 {
			return INTERPRET_OK
		}
		frame := &vm.frames[vm.frameCount-1]
		// Optionally print debug info if tracing is enabled.
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

		// Read the next opcode.
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
			// Access a property from a struct instance.
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
			// Set a property on a struct instance.
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
			// Addition: support both numeric addition and string concatenation.
			if peek(0).Type == runtime.VAL_OBJ && peek(1).Type == runtime.VAL_OBJ {
				concatenate()
			} else if peek(0).Type == runtime.VAL_NUMBER && peek(1).Type == runtime.VAL_NUMBER {
				binaryOp(func(a, b float64) float64 { return a + b }, "+")
			} else {
				return runtimeError("Operator '+' requires two numbers or two strings (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
		case uint8(runtime.OP_SUBTRACT):
			// Subtraction: supports numeric subtraction and a special "crop" operation for strings.
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
			// Multiplication: only defined for numbers.
			if peek(0).Type != runtime.VAL_NUMBER || peek(1).Type != runtime.VAL_NUMBER {
				return runtimeError("Both operands for '*' must be numbers (got %s and %s).", typeName(peek(1)), typeName(peek(0)))
			}
			b := Pop()
			a := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number * b.Number})
		case uint8(runtime.OP_DIVIDE):
			// Division: check for division by zero.
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
			// Modulo operation: check for modulo by zero.
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
			// Logical NOT: converts a value to its boolean negation.
			val := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: isFalsey(val)})
		case uint8(runtime.OP_NEGATE):
			// Negation: applies unary minus to a number.
			if peek(0).Type != runtime.VAL_NUMBER {
				return runtimeError("Unary '-' requires a number (got %s).", typeName(peek(0)))
			}
			val := Pop()
			Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: -val.Number})
		case uint8(runtime.OP_PRINT):
			// Print the top value on the stack.
			runtime.PrintValue(Pop())
			fmt.Println()
		case uint8(runtime.OP_JUMP):
			// Unconditional jump: move the instruction pointer by a given offset.
			offset := int(readShort(frame))
			frame.ip += offset
		case uint8(runtime.OP_JUMP_IF_FALSE):
			// Conditional jump: jump if the top of the stack is falsey.
			offset := int(readShort(frame))
			if isFalsey(peek(0)) {
				frame.ip += offset
			}
		case uint8(runtime.OP_JUMP_IF_TRUE):
			// Conditional jump: jump if the top of the stack is truthy.
			offset := int(readShort(frame))
			if isTruth(peek(0)) {
				frame.ip += offset
			}
		case uint8(runtime.OP_LOOP):
			// Loop back: subtract offset from the instruction pointer.
			offset := int(readShort(frame))
			frame.ip -= offset
		case uint8(runtime.OP_BREAK):
			// Break out of a loop by adding an offset.
			offset := int(readShort(frame))
			frame.ip += offset
		case uint8(runtime.OP_CONTINUE):
			// Continue to next loop iteration by subtracting an offset.
			offset := int(readShort(frame))
			frame.ip -= offset
		case uint8(runtime.OP_CALL):
			// Function call: read argument count and attempt to call the callee.
			argCount := int(readByte(frame))
			if !callValue(peek(argCount), argCount) {
				return INTERPRET_RUNTIME_ERROR
			}
		case uint8(runtime.OP_CLOSURE):
			// Create a closure from a function constant and capture its upvalues.
			function := readConstant(frame).Obj.(*runtime.ObjFunction)
			closure := runtime.NewClosure(function)
			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: closure})
			// For each upvalue, determine if it is a local or an upvalue from the enclosing function.
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
			// Close any upvalues that reference a variable going out of scope.
			closeUpvalues(&vm.stack[vm.stackTop-1])
			Pop()
		case uint8(runtime.OP_RETURN):
			// Return from the current function call.
			result := Pop()
			closeUpvalues(&vm.stack[frame.slots])
			vm.frameCount--
			if vm.frameCount == 0 {
				// End of the top-level script.
				Pop()
				return INTERPRET_OK
			} else {
				// Restore the caller's stack and push the return value.
				vm.stackTop = frame.slots
				Push(result)
				frame = &vm.frames[vm.frameCount-1]
			}
		case uint8(runtime.OP_STRUCT):
			// Create a new struct type instance.
			name := readString(frame)
			objStruct := runtime.NewStruct(name)
			fieldCount := int(readByte(frame))
			// For each field, read its name and default value.
			for i := 0; i < fieldCount; i++ {
				fieldName := readConstant(frame).Obj.(*runtime.ObjString)
				defaultValue := readConstant(frame)
				objStruct.Fields[fieldName] = defaultValue
			}
			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: objStruct})
		case uint8(runtime.OP_ARRAY_GET):
			// Array indexing: retrieve an element from an array.
			indexVal := Pop()
			arrayVal := Pop()

			if arrayVal.Type != runtime.VAL_OBJ {
				return runtimeError("Only arrays can be indexed with [].")
			}
			array, ok := arrayVal.Obj.(*runtime.ObjArray)
			if !ok {
				return runtimeError("Only arrays can be indexed with [].")
			}
			if indexVal.Type != runtime.VAL_NUMBER {
				return runtimeError("Array index must be a number.")
			}

			index := int(indexVal.Number)
			if index < 0 || index >= len(array.Elements) {
				return runtimeError("Array index out of bounds.")
			}

			Push(array.Elements[index])
		case uint8(runtime.OP_ARRAY):
			// Create a new array object from a list of elements.
			elementCount := int(readByte(frame))
			elements := make([]runtime.Value, elementCount)
			for i := elementCount - 1; i >= 0; i-- {
				elements[i] = Pop()
			}
			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewArray(elements)})
		case uint8(runtime.OP_ARRAY_SET):
			// Set an array element at a specified index.
			value := Pop()
			indexVal := Pop()
			arrayVal := Pop()

			if arrayVal.Type != runtime.VAL_OBJ {
				return runtimeError("Only arrays can be indexed with [].")
			}
			array, ok := arrayVal.Obj.(*runtime.ObjArray)
			if !ok {
				return runtimeError("Only arrays can be indexed with [].")
			}
			if indexVal.Type != runtime.VAL_NUMBER {
				return runtimeError("Array index must be a number.")
			}

			index := int(indexVal.Number)
			if index < 0 || index >= len(array.Elements) {
				return runtimeError("Array index out of bounds.")
			}

			array.Elements[index] = value
			Push(value)
		case uint8(runtime.OP_ARRAY_LEN):
			// Get the length of an array.
			arrayVal := Pop()
			if arrayVal.Type != runtime.VAL_OBJ {
				return runtimeError("Can only get length of arrays.")
			}
			array, ok := arrayVal.Obj.(*runtime.ObjArray)
			if !ok {
				return runtimeError("Can only get length of arrays.")
			}
			Push(runtime.Value{
				Type:   runtime.VAL_NUMBER,
				Number: float64(len(array.Elements)),
			})
		case uint8(runtime.OP_ARRAY_SLICE):
			// Create a slice of an array given start and end indices.
			endVal := Pop()
			startVal := Pop()
			arrayVal := Pop()

			if arrayVal.Type != runtime.VAL_OBJ {
				return runtimeError("Expected array for slice operation.")
			}
			array, ok := arrayVal.Obj.(*runtime.ObjArray)
			if !ok {
				return runtimeError("Expected array for slice operation.")
			}

			start := 0
			if startVal.Type != runtime.VAL_NULL {
				if startVal.Type != runtime.VAL_NUMBER {
					return runtimeError("Slice start must be a number.")
				}
				start = int(startVal.Number)
				if start < 0 {
					start += len(array.Elements)
				}
			}

			end := len(array.Elements)
			if endVal.Type != runtime.VAL_NULL {
				if endVal.Type != runtime.VAL_NUMBER {
					return runtimeError("Slice end must be a number.")
				}
				end = int(endVal.Number)
				if end < 0 {
					end += len(array.Elements)
				}
			}

			// Clamp indices to valid range.
			if start < 0 {
				start = 0
			} else if start > len(array.Elements) {
				start = len(array.Elements)
			}

			if end < 0 {
				end = 0
			} else if end > len(array.Elements) {
				end = len(array.Elements)
			}

			if start > end {
				start, end = end, start
			}

			elements := array.Elements[start:end]
			newArray := runtime.NewArray(elements)
			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: newArray})
		}
	}
}

// callValue attempts to call a value, which can be a function, native function, or struct constructor.
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
			// For struct constructors, create a new instance.
			vm.stack[vm.stackTop-argCount-1] = runtime.ObjVal(runtime.NewInstance(obj))
			return true
		default:
			// Non-callable object type.
		}
	}
	runtimeError("Cannot call %s; only functions and structs are callable.", typeName(callee))
	return false
}

// call sets up a new call frame for a closure, verifying the argument count.
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
