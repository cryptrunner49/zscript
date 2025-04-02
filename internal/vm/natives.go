package vm

import (
	"fmt"
	"sort"
	"time"

	"github.com/cryptrunner49/goseedvm/internal/common"
	"github.com/cryptrunner49/goseedvm/internal/runtime"
)

// defineAllNatives registers all native functions (built-in functions) to the VM.
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
	defineNative("to_chars", toCharsNative)
	// New native functions:
	defineNative("array_sort", arraySortNative)
	defineNative("array_split", arraySplitNative)
	defineNative("array_join", arrayJoinNative)
	defineNative("array_sorted_push", arraySortedPushNative)
	// New search functions:
	defineNative("array_linear_search", arrayLinearSearchNative)
	defineNative("array_binary_search", arrayBinarySearchNative)
}

// defineNative registers a single native function in the VM's global table.
// It creates a string object for the function name, wraps the native function in an ObjNative,
// and then stores it in the globals map.
func defineNative(name string, function runtime.NativeFn) {
	nameObj := runtime.NewObjString(name)
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: nameObj})
	nativeObj := &runtime.ObjNative{Function: function}
	Push(runtime.Value{Type: runtime.VAL_OBJ, Obj: nativeObj})
	vm.globals[nameObj] = vm.stack[vm.stackTop-1]
	Pop()
	Pop()
}

// enableDebugPrint turns on bytecode debug printing.
func enableDebugPrint(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}
	common.DebugPrintCode = true
	return runtime.Value{}
}

// enableTraceExecution turns on instruction-level execution tracing.
func enableTraceExecution(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}
	common.DebugTraceExecution = true
	return runtime.Value{}
}

// disableDebugPrint turns off bytecode debug printing.
func disableDebugPrint(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}
	common.DebugPrintCode = false
	return runtime.Value{}
}

// disableTraceExecution turns off instruction-level execution tracing.
func disableTraceExecution(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString("Error: to_str expects 1 argument")}
	}
	common.DebugTraceExecution = false
	return runtime.Value{}
}

// clockNative returns the current time in seconds as a number.
func clockNative(argCount int, args []runtime.Value) runtime.Value {
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(time.Now().UnixNano()) / 1e9,
	}
}

// toStr converts a value to its string representation.
func toStr(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
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
		str = fmt.Sprintf("%g", value.Number)
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

// arrayLenNative returns the length of an array.
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

// arrayPushNative pushes one or more elements onto an array and returns the new length.
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

// arrayPopNative removes the last element from an array and returns it.
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

// arrayIterNative creates an iterator for an array.
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

// iterNextNative advances the iterator to the next element.
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

// iterValueNative returns the current value pointed to by the iterator.
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

// iterDoneNative returns a boolean indicating whether the iterator has finished iterating over the array.
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

// toCharsNative converts a string into an array of single-character strings.
func toCharsNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("to_chars() expects exactly 1 argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strVal := args[0]
	if strVal.Type != runtime.VAL_OBJ {
		runtimeError("to_chars() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := strVal.Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("to_chars() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	chars := make([]runtime.Value, len(strObj.Chars))
	for i, r := range strObj.Chars {
		chars[i] = runtime.ObjVal(runtime.NewObjString(string(r)))
	}
	return runtime.ObjVal(runtime.NewArray(chars))
}

// ============================================================================
// New Native Functions: Search and Extended Array Operations
// ============================================================================

// arraySortNative sorts an array in ascending order.
// It works with any type by comparing the string representation of elements.
// The sort is done in-place and the sorted array is returned.
func arraySortNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'array_sort' expects 1 argument (the array).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_sort' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_sort' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	valueToString := func(v runtime.Value) string {
		switch v.Type {
		case runtime.VAL_BOOL:
			if v.Bool {
				return "true"
			}
			return "false"
		case runtime.VAL_NULL:
			return "null"
		case runtime.VAL_NUMBER:
			return fmt.Sprintf("%g", v.Number)
		case runtime.VAL_OBJ:
			if strObj, ok := v.Obj.(*runtime.ObjString); ok {
				return strObj.Chars
			}
			return toStr(1, []runtime.Value{v}).Obj.(*runtime.ObjString).Chars
		default:
			return "unknown"
		}
	}
	sort.Slice(array.Elements, func(i, j int) bool {
		return valueToString(array.Elements[i]) < valueToString(array.Elements[j])
	})
	return runtime.ObjVal(array)
}

// arraySplitNative splits an array into subarrays using a separator element.
// Every occurrence of the separator starts a new subarray.
func arraySplitNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'array_split' expects 2 arguments: an array and a separator.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_split' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_split' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	separator := args[1]
	var resultElements []runtime.Value
	var currentSplit []runtime.Value
	for _, elem := range array.Elements {
		if runtime.Equal(elem, separator) {
			currentSubArray := runtime.ObjVal(runtime.NewArray(currentSplit))
			resultElements = append(resultElements, currentSubArray)
			currentSplit = []runtime.Value{}
		} else {
			currentSplit = append(currentSplit, elem)
		}
	}
	currentSubArray := runtime.ObjVal(runtime.NewArray(currentSplit))
	resultElements = append(resultElements, currentSubArray)
	return runtime.ObjVal(runtime.NewArray(resultElements))
}

// arrayJoinNative joins two or more arrays into one.
// It concatenates the elements of each array in the order provided.
func arrayJoinNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount < 2 {
		runtimeError("'array_join' expects at least 2 arguments (arrays).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	var joinedElements []runtime.Value
	for i := 0; i < argCount; i++ {
		if args[i].Type != runtime.VAL_OBJ {
			runtimeError("'array_join' can only join arrays.")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
		arr, ok := args[i].Obj.(*runtime.ObjArray)
		if !ok {
			runtimeError("'array_join' can only join arrays.")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
		joinedElements = append(joinedElements, arr.Elements...)
	}
	return runtime.ObjVal(runtime.NewArray(joinedElements))
}

// arraySortedPushNative inserts a new element into a sorted array while keeping it sorted.
// The ordering is determined by comparing the string representation of elements.
func arraySortedPushNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'array_sorted_push' expects 2 arguments: an array and a value.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_sorted_push' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_sorted_push' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	newVal := args[1]
	valueToString := func(v runtime.Value) string {
		switch v.Type {
		case runtime.VAL_BOOL:
			if v.Bool {
				return "true"
			}
			return "false"
		case runtime.VAL_NULL:
			return "null"
		case runtime.VAL_NUMBER:
			return fmt.Sprintf("%g", v.Number)
		case runtime.VAL_OBJ:
			if strObj, ok := v.Obj.(*runtime.ObjString); ok {
				return strObj.Chars
			}
			return toStr(1, []runtime.Value{v}).Obj.(*runtime.ObjString).Chars
		default:
			return "unknown"
		}
	}
	inserted := false
	newStr := valueToString(newVal)
	for i, elem := range array.Elements {
		if newStr < valueToString(elem) {
			array.Elements = append(array.Elements[:i], append([]runtime.Value{newVal}, array.Elements[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		array.Elements = append(array.Elements, newVal)
	}
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(len(array.Elements)),
	}
}

// arrayLinearSearchNative performs a linear search on an array.
// It returns the index of the first occurrence of the search value, or -1 if not found.
func arrayLinearSearchNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'array_linear_search' expects 2 arguments: an array and a search value.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_linear_search' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_linear_search' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	searchVal := args[1]
	for i, elem := range array.Elements {
		if runtime.Equal(elem, searchVal) {
			return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(i)}
		}
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: -1}
}

// arrayBinarySearchNative performs a binary search on a sorted array.
// It returns the index of the search value, or -1 if not found.
// Comparison is based on the string representation of elements.
func arrayBinarySearchNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'array_binary_search' expects 2 arguments: a sorted array and a search value.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_binary_search' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_binary_search' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	searchVal := args[1]
	valueToString := func(v runtime.Value) string {
		switch v.Type {
		case runtime.VAL_BOOL:
			if v.Bool {
				return "true"
			}
			return "false"
		case runtime.VAL_NULL:
			return "null"
		case runtime.VAL_NUMBER:
			return fmt.Sprintf("%g", v.Number)
		case runtime.VAL_OBJ:
			if strObj, ok := v.Obj.(*runtime.ObjString); ok {
				return strObj.Chars
			}
			return toStr(1, []runtime.Value{v}).Obj.(*runtime.ObjString).Chars
		default:
			return "unknown"
		}
	}
	low := 0
	high := len(array.Elements) - 1
	searchStr := valueToString(searchVal)
	for low <= high {
		mid := (low + high) / 2
		midStr := valueToString(array.Elements[mid])
		if midStr == searchStr {
			return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(mid)}
		} else if midStr < searchStr {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: -1}
}
