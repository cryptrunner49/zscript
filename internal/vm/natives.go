package vm

import (
	"fmt"
	"time"

	"github.com/cryptrunner49/goseedvm/internal/common"
	"github.com/cryptrunner49/goseedvm/internal/runtime"
)

func defineAllNatives() {
	defineNative("clock", clockNative)
	defineNative("to_str", toStr)
	defineNative("enable_debug", enableDebugPrint)
	defineNative("enable_trace", enableTraceExecution)
	defineNative("disable_debug", disableDebugPrint)
	defineNative("disable_trace", disableTraceExecution)
}

func defineNative(name string, function runtime.NativeFn) {
	nameObj := runtime.NewObjString(name)
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: nameObj})
	nativeObj := &runtime.ObjNative{Function: function}
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: nativeObj})
	vm.globals[nameObj] = vm.stack[vm.stackTop-1]
	Pop()
	Pop()
}

func enableDebugPrint(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		// Handle incorrect argument count (though runtimeError isn't directly callable here)
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}

	common.DebugPrintCode = true
	return runtime.Value{}
}

func enableTraceExecution(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		// Handle incorrect argument count (though runtimeError isn't directly callable here)
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}

	common.DebugTraceExecution = true
	return runtime.Value{}
}

func disableDebugPrint(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		// Handle incorrect argument count (though runtimeError isn't directly callable here)
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}

	common.DebugPrintCode = false
	return runtime.Value{}
}

func disableTraceExecution(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		// Handle incorrect argument count (though runtimeError isn't directly callable here)
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}

	common.DebugTraceExecution = false
	return runtime.Value{}
}

func clockNative(argCount int, args []runtime.Value) runtime.Value {
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(time.Now().UnixNano()) / 1e9,
	}
}

func toStr(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		// Handle incorrect argument count (though runtimeError isn't directly callable here)
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}

	value := args[0]
	var str string

	switch value.Type {
	case runtime.VAL_BOOL:
		if value.Bool {
			str = "true"
		} else {
			str = "false"
		}
	case runtime.VAL_NULL:
		str = "null"
	case runtime.VAL_NUMBER:
		str = fmt.Sprintf("%g", value.Number) // Matches PrintValue behavior
	case runtime.VAL_OBJ:
		switch obj := value.Obj.(type) {
		case *runtime.ObjString:
			str = obj.Chars
		case *runtime.ObjFunction:
			if obj.Name != nil {
				str = "<fn " + obj.Name.Chars + ">"
			} else {
				str = "<fn>"
			}
		case *runtime.ObjClosure:
			if obj.Function.Name != nil {
				str = "<fn " + obj.Function.Name.Chars + ">"
			} else {
				str = "<fn>"
			}
		case *runtime.ObjNative:
			str = "<native fn>"
		case *runtime.ObjStruct:
			str = "<struct " + obj.Name.Chars + ">"
		case *runtime.ObjInstance:
			// Assuming the instance has a struct type with a name
			str = "<instance>"
		case *runtime.ObjUpvalue:
			str = "<upvalue>"
		default:
			str = "<object>"
		}
	default:
		str = "unknown"
	}

	return runtime.Value{
		Type: runtime.VAL_OBJ,
		Obj:  runtime.NewObjString(str),
	}
}
