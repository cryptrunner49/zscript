package vm

import (
	"fmt"
	"os"

	"github.com/cryptrunner49/zscript/internal/runtime"
)

// isFalsey returns true if a value is considered false in boolean context.
// In this VM, only null and false are falsey.
func isFalsey(val runtime.Value) bool {
	return val.Type == runtime.VAL_NULL || (val.Type == runtime.VAL_BOOL && !val.Bool)
}

// isTruth returns true if a value is considered true in boolean context.
func isTruth(val runtime.Value) bool {
	return val.Type != runtime.VAL_NULL && (val.Type != runtime.VAL_BOOL || val.Bool)
}

// typeName returns a string representing the type name of a runtime value.
func typeName(val runtime.Value) string {
	switch val.Type {
	case runtime.VAL_BOOL:
		return "boolean"
	case runtime.VAL_NULL:
		return "null"
	case runtime.VAL_NUMBER:
		return "number"
	case runtime.VAL_OBJ:
		switch val.Obj.(type) {
		case *runtime.ObjString:
			return "string"
		case *runtime.ObjFunction:
			return "function"
		case *runtime.ObjClosure:
			return "closure"
		case *runtime.ObjNative:
			return "native function"
		case *runtime.ObjStruct:
			return "struct"
		case *runtime.ObjInstance:
			return "instance"
		default:
			return "object"
		}
	default:
		return "unknown"
	}
}

// runtimeError prints a formatted runtime error message along with a backtrace of call frames.
// It then resets the VM's stack and returns an INTERPRET_RUNTIME_ERROR result.
func runtimeError(format string, args ...interface{}) InterpretResult {
	fmt.Fprintf(os.Stderr, "Runtime Error: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
	// Print the call stack.
	for i := vm.frameCount - 1; i >= 0; i-- {
		frame := &vm.frames[i]
		function := frame.closure.Function
		instruction := frame.ip - 1
		line := function.Chunk.Lines()[instruction]
		fmt.Fprintf(os.Stderr, "  at [line %d] in ", line)
		if function.Name == nil {
			fmt.Fprintln(os.Stderr, "top-level script")
		} else {
			fmt.Fprintf(os.Stderr, "function '%s()'\n", function.Name.Chars)
		}
	}
	resetStack()
	return INTERPRET_RUNTIME_ERROR
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

func createInstance(callee runtime.Value, argCount int) bool {
	if callee.Type == runtime.VAL_OBJ {
		if structObj, ok := callee.Obj.(*runtime.ObjStruct); ok {
			// Create new instance with default values
			instance := runtime.NewInstance(structObj)

			// Pop the force flag
			forceVal := Pop()
			if forceVal.Type != runtime.VAL_BOOL {
				runtimeError("Internal error: Expected boolean force flag for instance creation.")
				return false
			}
			force := forceVal.Bool

			// Base position of the struct on the stack
			base := vm.stackTop - (argCount * 2) - 1 // 2 slots per pair

			// Assign key-value pairs to fields
			for i := 0; i < argCount; i++ {
				value := Pop()  // Pop the value (e.g., 2, then 1)
				keyVal := Pop() // Pop the key (e.g., "y", then "x")
				if keyVal.Type != runtime.VAL_OBJ {
					runtimeError("Field name must be a string in instance initializer.")
					return false
				}
				key, ok := keyVal.Obj.(*runtime.ObjString)
				if !ok {
					runtimeError("Field name must be a string in instance initializer.")
					return false
				}
				if !force {
					// Check if the field exists in the struct
					if _, exists := structObj.Fields[key]; !exists {
						runtimeError("Unknown field '%s' in struct '%s'.", key.Chars, structObj.Name.Chars)
						return false
					}
				}
				instance.Fields[key] = value // Assign value to the specified field
			}

			// Replace the struct with the new instance
			vm.stack[base] = runtime.ObjVal(instance)
			vm.stackTop = base + 1
			return true
		}
	}
	runtimeError("Cannot instantiate %s with '{}'; only structs can be instantiated this way.", typeName(callee))
	return false
}
