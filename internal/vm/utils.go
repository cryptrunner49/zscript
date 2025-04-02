package vm

import (
	"fmt"
	"os"

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
