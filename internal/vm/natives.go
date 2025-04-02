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
	defineNative("len", arrayLenNative)
	defineNative("push", arrayPushNative)
	defineNative("pop", arrayPopNative)
	defineNative("array_iter", arrayIterNative)
	defineNative("iter_next", iterNextNative)
	defineNative("iter_value", iterValueNative)
	defineNative("iter_done", iterDoneNative)
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

func arrayLenNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'len' expects 1 argument (the array).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'len' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'len' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(len(array.Elements)),
	}
}

func arrayPushNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount < 1 {
		runtimeError("'push' expects at least 1 argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'push' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'push' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}

	for i := 1; i < argCount; i++ {
		array.Elements = append(array.Elements, args[i])
	}
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(len(array.Elements)),
	}
}

func arrayPopNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'pop' expects 1 argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'pop' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'pop' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}

	if len(array.Elements) == 0 {
		return runtime.Value{Type: runtime.VAL_NULL}
	}

	last := array.Elements[len(array.Elements)-1]
	array.Elements = array.Elements[:len(array.Elements)-1]
	return last
}

func arrayIterNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'array_iter' expects 1 argument (the array).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_iter' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_iter' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}

	return runtime.Value{
		Type: runtime.VAL_OBJ,
		Obj:  runtime.NewArrayIterator(array),
	}
}

func iterNextNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'iter_next' expects 1 argument (the iterator).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'iter_next' can only be used on iterators.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	iter, ok := args[0].Obj.(*runtime.ObjArrayIterator)
	if !ok {
		runtimeError("'iter_next' can only be used on iterators.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}

	iter.Index++
	return runtime.Value{Type: runtime.VAL_NULL}
}

func iterValueNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'iter_value' expects 1 argument (the iterator).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'iter_value' can only be used on iterators.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	iter, ok := args[0].Obj.(*runtime.ObjArrayIterator)
	if !ok {
		runtimeError("'iter_value' can only be used on iterators.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}

	if iter.Index >= len(iter.Array.Elements) {
		return runtime.Value{Type: runtime.VAL_NULL}
	}

	return iter.Array.Elements[iter.Index]
}

func iterDoneNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'iter_done' expects 1 argument (the iterator).")
		return runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'iter_done' can only be used on iterators.")
		return runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
	}
	iter, ok := args[0].Obj.(*runtime.ObjArrayIterator)
	if !ok {
		runtimeError("'iter_done' can only be used on iterators.")
		return runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
	}

	return runtime.Value{
		Type: runtime.VAL_BOOL,
		Bool: iter.Index >= len(iter.Array.Elements),
	}
}
