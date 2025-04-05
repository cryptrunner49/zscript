//go:build cgo
// +build cgo

package vm

/*
#cgo LDFLAGS: -ldl
#include <stdio.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <string.h>

// Load a library, automatically adding "lib" prefix and ".so" suffix if needed
void* load_library(const char* lib_name) {
    void* handle = dlopen(lib_name, RTLD_LAZY);
    if (!handle) {
        // If direct load fails, try with "lib" prefix and ".so" suffix
        char* prefixed_name = malloc(strlen(lib_name) + 7); // "lib" (3) + ".so" (3) + null terminator (1)
        if (!prefixed_name) {
            fprintf(stderr, "Memory allocation failed for library name\n");
            return NULL;
        }

        // Check if it already starts with "lib" and ends with ".so"
        if (strncmp(lib_name, "lib", 3) != 0 && strstr(lib_name, ".so") == NULL) {
            sprintf(prefixed_name, "lib%s.so", lib_name);
            handle = dlopen(prefixed_name, RTLD_LAZY);
        }

        if (!handle) {
            fprintf(stderr, "Cannot load library '%s' or '%s': %s\n", lib_name, prefixed_name, dlerror());
        }
        free(prefixed_name);
    }
    return handle;
}

// Get a function pointer
void* get_function(void* handle, const char* func_name) {
    void* func = dlsym(handle, func_name);
    if (!func) {
        fprintf(stderr, "Cannot load function '%s': %s\n", func_name, dlerror());
    }
    return func;
}

void close_library(void* handle) {
    dlclose(handle);
}
*/
import "C"

import (
	"fmt"
	"os"
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
	libHandles   []unsafe.Pointer                     // List of loaded library handles.
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
	for _, handle := range vm.libHandles {
		C.close_library(handle)
	}

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
func Interpret(source string, scriptPath string) InterpretResult {
	resetStack()
	function := compiler.Compile(source, scriptPath)
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
	var currentLibHandle unsafe.Pointer

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
			// Access a property from an object.
			instVal := peek(0)
			switch obj := instVal.Obj.(type) {
			case *runtime.ObjInstance:
				// For struct instances, look up the property in the fields map.
				name := readString(frame)
				if value, found := obj.Fields[name]; found {
					Pop() // Remove the array from the stack.
					Push(value)
				} else {
					return runtimeError("Property '%s' does not exist on this instance.", name.Chars)
				}
			case *runtime.ObjModule:
				// For struct instances, look up the property in the fields map.
				name := readString(frame)
				if value, found := obj.Fields[name]; found {
					Pop() // Remove the array from the stack.
					Push(value)
				} else {
					return runtimeError("Property '%s' does not exist on this instance.", name.Chars)
				}
			case *runtime.ObjArray:
				// Allow arrays to expose a "length" property.
				name := readString(frame)
				if name.Chars == "length" {
					Pop()
					Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(len(obj.Elements))})
				} else {
					return runtimeError("Cannot access property '%s' on array; only 'length' is supported.", name.Chars)
				}
			default:
				return runtimeError("Cannot access property on %s; only struct instances, arrays, and modules have properties.", typeName(instVal))
			}
		case uint8(runtime.OP_SET_PROPERTY):
			// Set a property on a struct instance.
			instVal := peek(1)
			switch obj := instVal.Obj.(type) {
			case *runtime.ObjInstance:
				name := readString(frame)
				obj.Fields[name] = peek(0)
				value := Pop()
				Pop()
				Push(value)
			case *runtime.ObjModule:
				name := readString(frame)
				obj.Fields[name] = peek(0)
				value := Pop()
				Pop()
				Push(value)
			default:
				return runtimeError("Cannot set property on %s; only struct instances and modules have fields.", typeName(instVal))
			}
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
			if peek(0).Type == runtime.VAL_NUMBER && peek(1).Type == runtime.VAL_NUMBER {
				b := Pop()
				a := Pop()
				Push(addNumbers(a, b))
			} else if peek(0).Type == runtime.VAL_OBJ && peek(1).Type == runtime.VAL_OBJ {
				b := peek(0)
				a := peek(1)
				switch obj2 := b.Obj.(type) {
				case *runtime.ObjInstance:
					if inst1, ok := a.Obj.(*runtime.ObjInstance); ok {
						result, err := addInstances(inst1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						return runtimeError("Operands must be of the same type for '+'. Got %s and %s.", typeName(a), typeName(b))
					}
				case *runtime.ObjMap:
					if map1, ok := a.Obj.(*runtime.ObjMap); ok {
						result := addMaps(map1, obj2)
						Pop()
						Pop()
						Push(result)
					} else {
						return runtimeError("Operands must be of the same type for '+'. Got %s and %s.", typeName(a), typeName(b))
					}
				case *runtime.ObjArray:
					if arr1, ok := a.Obj.(*runtime.ObjArray); ok {
						result := addArrays(arr1, obj2)
						Pop()
						Pop()
						Push(result)
					} else {
						return runtimeError("Operands must be of the same type for '+'. Got %s and %s.", typeName(a), typeName(b))
					}
				case *runtime.ObjString:
					if _, ok := a.Obj.(*runtime.ObjString); ok {
						b := Pop()
						a := Pop()
						Push(addStrings(a, b))
					} else {
						return runtimeError("Operands must be of the same type for '+'. Got %s and %s.", typeName(a), typeName(b))
					}
				default:
					b := Pop()
					a := Pop()
					Push(addStrings(a, b)) // Fallback to string concatenation
				}
			} else {
				b := Pop()
				a := Pop()
				Push(addStrings(a, b)) // Mixed types fallback to string concatenation
			}

		case uint8(runtime.OP_SUBTRACT):
			if peek(0).Type == runtime.VAL_NUMBER && peek(1).Type == runtime.VAL_NUMBER {
				b := Pop()
				a := Pop()
				Push(subtractNumbers(a, b))
			} else if peek(0).Type == runtime.VAL_OBJ && peek(1).Type == runtime.VAL_OBJ {
				b := peek(0)
				a := peek(1)
				switch obj2 := b.Obj.(type) {
				case *runtime.ObjInstance:
					if inst1, ok := a.Obj.(*runtime.ObjInstance); ok {
						result, err := subtractInstances(inst1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						return runtimeError("Operands must be of the same type for '-'. Got %s and %s.", typeName(a), typeName(b))
					}
				case *runtime.ObjMap:
					if map1, ok := a.Obj.(*runtime.ObjMap); ok {
						result := subtractMaps(map1, obj2)
						Pop()
						Pop()
						Push(result)
					} else {
						return runtimeError("Operands must be of the same type for '-'. Got %s and %s.", typeName(a), typeName(b))
					}
				case *runtime.ObjArray:
					if arr1, ok := a.Obj.(*runtime.ObjArray); ok {
						result := subtractArrays(arr1, obj2)
						Pop()
						Pop()
						Push(result)
					} else {
						return runtimeError("Operands must be of the same type for '-'. Got %s and %s.", typeName(a), typeName(b))
					}
				case *runtime.ObjString:
					if _, ok := a.Obj.(*runtime.ObjString); ok {
						b := Pop()
						a := Pop()
						Push(subtractStrings(a, b))
					} else {
						return runtimeError("Operands must be of the same type for '-'. Got %s and %s.", typeName(a), typeName(b))
					}
				default:
					b := Pop()
					a := Pop()
					Push(subtractStrings(a, b)) // Fallback to string cropping
				}
			} else {
				b := Pop()
				a := Pop()
				Push(subtractStrings(a, b)) // Mixed types fallback to string cropping
			}

		case uint8(runtime.OP_MULTIPLY):
			b := peek(0)
			a := peek(1)
			switch {
			case b.Type == runtime.VAL_NUMBER && a.Type == runtime.VAL_NUMBER:
				bVal := Pop()
				aVal := Pop()
				Push(multiplyNumbers(aVal, bVal))
			case b.Type == runtime.VAL_OBJ && a.Type == runtime.VAL_OBJ:
				switch obj2 := b.Obj.(type) {
				case *runtime.ObjInstance:
					if inst1, ok := a.Obj.(*runtime.ObjInstance); ok {
						result, err := multiplyInstances(inst1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						bVal := Pop()
						aVal := Pop()
						runtimeError("Operator '*' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
						Push(aVal)
					}
				case *runtime.ObjArray:
					if arr1, ok := a.Obj.(*runtime.ObjArray); ok {
						result, err := multiplyArrays(arr1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						bVal := Pop()
						aVal := Pop()
						runtimeError("Operator '*' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
						Push(aVal)
					}
				default:
					bVal := Pop()
					aVal := Pop()
					runtimeError("Operator '*' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
					Push(aVal)
				}
			default:
				bVal := Pop()
				aVal := Pop()
				runtimeError("Operator '*' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
				Push(aVal)
			}

		case uint8(runtime.OP_DIVIDE):
			b := peek(0)
			a := peek(1)
			switch {
			case b.Type == runtime.VAL_NUMBER && a.Type == runtime.VAL_NUMBER:
				bVal := Pop()
				aVal := Pop()
				Push(divideNumbers(aVal, bVal))
			case b.Type == runtime.VAL_OBJ && a.Type == runtime.VAL_OBJ:
				switch obj2 := b.Obj.(type) {
				case *runtime.ObjInstance:
					if inst1, ok := a.Obj.(*runtime.ObjInstance); ok {
						result, err := divideInstances(inst1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						bVal := Pop()
						aVal := Pop()
						runtimeError("Operator '/' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
						Push(aVal)
					}
				case *runtime.ObjArray:
					if arr1, ok := a.Obj.(*runtime.ObjArray); ok {
						result, err := divideArrays(arr1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						bVal := Pop()
						aVal := Pop()
						runtimeError("Operator '/' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
						Push(aVal)
					}
				default:
					bVal := Pop()
					aVal := Pop()
					runtimeError("Operator '/' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
					Push(aVal)
				}
			default:
				bVal := Pop()
				aVal := Pop()
				runtimeError("Operator '/' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
				Push(aVal)
			}

		case uint8(runtime.OP_MOD):
			b := peek(0)
			a := peek(1)
			switch {
			case b.Type == runtime.VAL_NUMBER && a.Type == runtime.VAL_NUMBER:
				bVal := Pop()
				aVal := Pop()
				Push(modNumbers(aVal, bVal))
			case b.Type == runtime.VAL_OBJ && a.Type == runtime.VAL_OBJ:
				switch obj2 := b.Obj.(type) {
				case *runtime.ObjInstance:
					if inst1, ok := a.Obj.(*runtime.ObjInstance); ok {
						result, err := modInstances(inst1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						bVal := Pop()
						aVal := Pop()
						runtimeError("Operator '%%' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
						Push(aVal)
					}
				case *runtime.ObjArray:
					if arr1, ok := a.Obj.(*runtime.ObjArray); ok {
						result, err := modArrays(arr1, obj2)
						if err != INTERPRET_OK {
							return err
						}
						Pop()
						Pop()
						Push(result)
					} else {
						bVal := Pop()
						aVal := Pop()
						runtimeError("Operator '%%' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
						Push(aVal)
					}
				default:
					bVal := Pop()
					aVal := Pop()
					runtimeError("Operator '%%' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
					Push(aVal)
				}
			default:
				bVal := Pop()
				aVal := Pop()
				runtimeError("Operator '%%' requires numbers or arrays of numbers (got %s and %s).", typeName(aVal), typeName(bVal))
				Push(aVal)
			}

		case uint8(runtime.OP_NOT):
			// Logical NOT: converts a value to its boolean negation.
			val := Pop()
			Push(runtime.Value{Type: runtime.VAL_BOOL, Bool: isFalsey(val)})
		case uint8(runtime.OP_NEGATE):
			// Negation: applies unary minus to a number or an array of numbers.
			if peek(0).Type == runtime.VAL_NUMBER {
				val := Pop()
				Push(runtime.Value{Type: runtime.VAL_NUMBER, Number: -val.Number})
			} else if peek(0).Type == runtime.VAL_OBJ {
				// Check if the object is an array.
				if array, ok := peek(0).Obj.(*runtime.ObjArray); ok {
					// Verify that every element is a number.
					for _, elem := range array.Elements {
						if elem.Type != runtime.VAL_NUMBER {
							return runtimeError("Unary '-' requires a number or an array of numbers (got non-number element).")
						}
					}
					// Create a new array with negated numbers.
					newElements := make([]runtime.Value, len(array.Elements))
					for i, elem := range array.Elements {
						newElements[i] = runtime.Value{Type: runtime.VAL_NUMBER, Number: -elem.Number}
					}
					// Pop the original array and push the new negated array.
					Pop()
					Push(runtime.ObjVal(runtime.NewArray(newElements)))
				} else {
					return runtimeError("Unary '-' requires a number or an array of numbers (got %s).", typeName(peek(0)))
				}
			} else {
				return runtimeError("Unary '-' requires a number (got %s).", typeName(peek(0)))
			}
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

		case uint8(runtime.OP_GET_VALUE):
			index := Pop()
			obj := Pop()

			if obj.Type != runtime.VAL_OBJ {
				runtimeError("Cannot index non-object type.")
				break
			}

			switch o := obj.Obj.(type) {
			case *runtime.ObjArray:
				if index.Type != runtime.VAL_NUMBER {
					runtimeError("Array index must be a number.")
					break
				}
				idx := int(index.Number)
				if idx < 0 || idx >= len(o.Elements) {
					runtimeError("Array index out of bounds.")
					break
				}
				Push(o.Elements[idx])
			case *runtime.ObjMap:
				if index.Type != runtime.VAL_OBJ {
					runtimeError("Map key must be a string.")
					break
				}
				key, ok := index.Obj.(*runtime.ObjString)
				if !ok {
					runtimeError("Map key must be a string.")
					break
				}
				val, exists := o.Entries[key]
				if exists {
					Push(val)
				} else {
					Push(runtime.Value{Type: runtime.VAL_NULL})
				}
			default:
				runtimeError("Object does not support indexing.")
			}

		case uint8(runtime.OP_SET_VALUE):
			value := Pop()
			index := Pop()
			obj := Pop()

			if obj.Type != runtime.VAL_OBJ {
				runtimeError("Cannot index non-object type.")
				break
			}

			switch o := obj.Obj.(type) {
			case *runtime.ObjArray:
				if index.Type != runtime.VAL_NUMBER {
					runtimeError("Array index must be a number.")
					break
				}
				idx := int(index.Number)
				if idx < 0 || idx >= len(o.Elements) {
					runtimeError("Array index out of bounds.")
					break
				}
				o.Elements[idx] = value
				Push(value)
			case *runtime.ObjMap:
				if index.Type != runtime.VAL_OBJ {
					runtimeError("Map key must be a string.")
					break
				}
				key, ok := index.Obj.(*runtime.ObjString)
				if !ok {
					runtimeError("Map key must be a string.")
					break
				}
				o.Entries[key] = value
				Push(value)
			default:
				runtimeError("Object does not support indexing.")
			}

		case uint8(runtime.OP_ARRAY):
			// Create a new array object from a list of elements.
			elementCount := int(readByte(frame))
			elements := make([]runtime.Value, elementCount)
			for i := elementCount - 1; i >= 0; i-- {
				elements[i] = Pop()
			}
			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewArray(elements)})

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

		case uint8(runtime.OP_MODULE):
			// Create a new module type instance.
			name := readString(frame)
			objModule := runtime.NewModule(name)
			fieldCount := int(readByte(frame))

			// For each field, read its name and value from the stack in the correct order.
			for i := 0; i < fieldCount; i++ {
				fieldName := readConstant(frame).Obj.(*runtime.ObjString)
				defaultValue := readConstant(frame)
				objModule.Fields[fieldName] = defaultValue
			}

			Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: objModule})
		case uint8(runtime.OP_IMPORT):
			path := readString(frame).Chars
			pathObj := runtime.NewObjString(path)
			if cached, exists := vm.globals[pathObj]; exists {
				Push(cached)
				break
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return runtimeError("Failed to load module '%s': %v", path, err)
			}
			function := compiler.Compile(string(content), path)
			if function == nil {
				return INTERPRET_COMPILE_ERROR
			}
			closure := runtime.NewClosure(function)
			vm.globals[pathObj] = runtime.ObjVal(closure)
			Push(runtime.ObjVal(closure))
			if !callValue(runtime.ObjVal(closure), 0) {
				return INTERPRET_RUNTIME_ERROR
			}
			Pop()                     // Pop the null return value
			Push(vm.globals[pathObj]) // Push the closure back
		case uint8(runtime.OP_USE):
			libName := readString(frame).Chars
			// Use the full library name as provided (e.g., "libmylib.so" or "mylib.dll")
			currentLibHandle = C.load_library(C.CString(libName))
			if currentLibHandle == nil {
				return runtimeError("Failed to load library '%s'.", libName)
			}
			vm.libHandles = append(vm.libHandles, currentLibHandle)

		case uint8(runtime.OP_DEFINE_EXTERN):
			returnTypeConstant := readConstant(frame)
			returnType := returnTypeConstant.Obj.(*runtime.ObjString).Chars
			paramCount := int(readByte(frame))
			paramTypes := make([]string, paramCount)
			for i := 0; i < paramCount; i++ {
				paramTypeConstant := readConstant(frame)
				paramTypes[i] = paramTypeConstant.Obj.(*runtime.ObjString).Chars
			}
			funcNameConstant := readConstant(frame)
			funcName := funcNameConstant.Obj.(*runtime.ObjString).Chars
			cFunc := C.get_function(currentLibHandle, C.CString(funcName))
			if cFunc == nil {
				return runtimeError("Failed to load function '%s' from library.", funcName)
			}
			nativeFunc := createNativeFunc(funcName, cFunc, returnType, paramTypes)
			nameObj := runtime.NewObjString(funcName)
			vm.globals[nameObj] = runtime.Value{Type: runtime.VAL_OBJ, Obj: nativeFunc}

		case uint8(runtime.OP_MAP):
			pairCount := int(readByte(frame))
			mapObj := runtime.NewMap()
			for i := 0; i < pairCount; i++ {
				value := Pop()
				keyVal := Pop()
				if keyVal.Type != runtime.VAL_OBJ {
					runtimeError("Map key must be a string")
					continue
				}
				key, ok := keyVal.Obj.(*runtime.ObjString)
				if !ok {
					runtimeError("Map key must be a string")
					continue
				}
				mapObj.Entries[key] = value
			}
			Push(runtime.ObjVal(mapObj))
		}
	}
}
