//go:build cgo
// +build cgo

package vm

/*
#cgo pkg-config: readline
#cgo LDFLAGS: -ldl
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <dlfcn.h>

// Wrapper functions for calling C functions with different signatures
void call_void(void* func) {
    ((void (*)())func)();
}

int32_t call_int32(void* func) {
    return ((int32_t (*)())func)();
}

double call_double(void* func) {
    return ((double (*)())func)();
}

_Bool call_bool(void* func) {
    return ((_Bool (*)())func)();
}

uint32_t call_uint32(void* func) {
    return ((uint32_t (*)())func)();
}

void call_void_int32(void* func, int32_t arg) {
    ((void (*)(int32_t))func)(arg);
}

int32_t call_int32_int32(void* func, int32_t arg) {
    return ((int32_t (*)(int32_t))func)(arg);
}

double call_double_int32(void* func, int32_t arg) {
    return ((double (*)(int32_t))func)(arg);
}

_Bool call_bool_int32(void* func, int32_t arg) {
    return ((_Bool (*)(int32_t))func)(arg);
}

void call_void_double(void* func, double arg) {
    ((void (*)(double))func)(arg);
}

int32_t call_int32_double(void* func, double arg) {
    return ((int32_t (*)(double))func)(arg);
}

double call_double_double(void* func, double arg) {
    return ((double (*)(double))func)(arg);
}

_Bool call_bool_double(void* func, double arg) {
    return ((_Bool (*)(double))func)(arg);
}

uint32_t call_uint32_size_t(void* func, size_t arg) {
    return ((uint32_t (*)(size_t))func)(arg);
}
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/cryptrunner49/goseedvm/internal/runtime"
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

func createNativeFunc(funcName string, cFunc unsafe.Pointer, returnType string, paramTypes []string) *runtime.ObjNative {
	return &runtime.ObjNative{
		Function: func(argCount int, args []runtime.Value) runtime.Value {
			if argCount != len(paramTypes) {
				runtimeError("Function '%s' expects %d arguments but got %d.", funcName, len(paramTypes), argCount)
				return runtime.Value{Type: runtime.VAL_NULL}
			}
			cArgs := make([]interface{}, len(paramTypes))
			for i, pt := range paramTypes {
				switch pt {
				case "int32_t", "int":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					cArgs[i] = C.int32_t(args[i].Number)
				case "double":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					cArgs[i] = C.double(args[i].Number)
				case "bool":
					if args[i].Type != runtime.VAL_BOOL {
						runtimeError("Argument %d of '%s' must be a boolean.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					cArgs[i] = C.bool(args[i].Bool)
				case "size_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					cArgs[i] = C.size_t(args[i].Number)
				default:
					runtimeError("Unsupported parameter type '%s' for '%s'.", pt, funcName)
					return runtime.Value{Type: runtime.VAL_NULL}
				}
			}
			if len(paramTypes) == 0 {
				switch returnType {
				case "void":
					C.call_void(cFunc)
					return runtime.Value{Type: runtime.VAL_NULL}
				case "int32_t", "int":
					result := C.call_int32(cFunc)
					return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
				case "double":
					result := C.call_double(cFunc)
					return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
				case "bool":
					result := C.call_bool(cFunc)
					return runtime.Value{Type: runtime.VAL_BOOL, Bool: bool(result)}
				case "uint32_t":
					result := C.call_uint32(cFunc)
					return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
				default:
					runtimeError("Unsupported return type '%s' for '%s'.", returnType, funcName)
					return runtime.Value{Type: runtime.VAL_NULL}
				}
			} else if len(paramTypes) == 1 {
				switch paramTypes[0] {
				case "int32_t", "int":
					switch returnType {
					case "void":
						C.call_void_int32(cFunc, cArgs[0].(C.int32_t))
						return runtime.Value{Type: runtime.VAL_NULL}
					case "int32_t", "int":
						result := C.call_int32_int32(cFunc, cArgs[0].(C.int32_t))
						return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
					case "double":
						result := C.call_double_int32(cFunc, cArgs[0].(C.int32_t))
						return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
					case "bool":
						result := C.call_bool_int32(cFunc, cArgs[0].(C.int32_t))
						return runtime.Value{Type: runtime.VAL_BOOL, Bool: bool(result)}
					default:
						runtimeError("Unsupported return type '%s' for '%s' with int32_t parameter.", returnType, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
				case "double":
					switch returnType {
					case "void":
						C.call_void_double(cFunc, cArgs[0].(C.double))
						return runtime.Value{Type: runtime.VAL_NULL}
					case "int32_t", "int":
						result := C.call_int32_double(cFunc, cArgs[0].(C.double))
						return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
					case "double":
						result := C.call_double_double(cFunc, cArgs[0].(C.double))
						return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
					case "bool":
						result := C.call_bool_double(cFunc, cArgs[0].(C.double))
						return runtime.Value{Type: runtime.VAL_BOOL, Bool: bool(result)}
					default:
						runtimeError("Unsupported return type '%s' for '%s' with double parameter.", returnType, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
				case "size_t":
					switch returnType {
					case "uint32_t":
						result := C.call_uint32_size_t(cFunc, cArgs[0].(C.size_t))
						return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(result)}
					default:
						runtimeError("Unsupported return type '%s' for '%s' with size_t parameter.", returnType, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
				default:
					runtimeError("Unsupported parameter type '%s' for '%s'.", paramTypes[0], funcName)
					return runtime.Value{Type: runtime.VAL_NULL}
				}
			}
			runtimeError("Unsupported function signature for '%s' (only 0 or 1 parameters supported).", funcName)
			return runtime.Value{Type: runtime.VAL_NULL}
		},
	}
}
