package vm

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cryptrunner49/goseedvm/internal/common"
	"github.com/cryptrunner49/goseedvm/internal/runtime"
)

// Seed the random number generator once during initialization
func init() {
	rand.Seed(time.Now().UnixNano())
}

// unescapeString replaces common escape sequences like \n with their actual characters.
func unescapeString(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\\`, "\\")
	// Add more escape sequences as needed
	return s
}

// defineAllNatives registers all native functions (built-in functions) to the VM.
func defineAllNatives() {
	// Debug
	defineNative("enable_debug", enableDebugPrint)
	defineNative("enable_trace", enableTraceExecution)
	defineNative("disable_debug", disableDebugPrint)
	defineNative("disable_trace", disableTraceExecution)

	// String
	defineNative("to_str", toStr)
	defineNative("to_chars", toCharsNative)
	defineNative("char_at", charAtNative)
	defineNative("substring", substringNative)
	defineNative("str_index_of", strIndexOfNative)
	defineNative("str_last_index_of", strLastIndexOfNative)
	defineNative("str_contains", strContainsNative)
	defineNative("starts_with", startsWithNative)
	defineNative("ends_with", endsWithNative)
	defineNative("to_upper", toUpperNative)
	defineNative("to_lower", toLowerNative)
	defineNative("trim", trimNative)
	defineNative("split", splitNative)
	defineNative("replace", replaceNative)
	defineNative("str_length", strLengthNative)

	// Array
	defineNative("len", arrayLenNative)
	defineNative("push", arrayPushNative)
	defineNative("pop", arrayPopNative)
	defineNative("array_sort", arraySortNative)
	defineNative("array_split", arraySplitNative)
	defineNative("array_join", arrayJoinNative)
	defineNative("array_sorted_push", arraySortedPushNative)
	defineNative("array_linear_search", arrayLinearSearchNative)
	defineNative("array_binary_search", arrayBinarySearchNative)
	defineNative("index_of", arrayIndexOfNative)
	defineNative("last_index_of", arrayLastIndexOfNative)
	defineNative("array_contains", arrayContainsNative)
	defineNative("array_clear", arrayClearNative)
	defineNative("array_reverse", arrayReverseNative)
	defineNative("array_to_string", arrayToStringNative)
	defineNative("array_remove", arrayRemoveNative)

	// Iterator
	defineNative("array_iter", arrayIterNative)
	defineNative("iter_next", iterNextNative)
	defineNative("iter_value", iterValueNative)
	defineNative("iter_done", iterDoneNative)

	// Map
	defineNative("map_remove", mapRemoveNative)
	defineNative("map_contains_key", mapContainsKeyNative)
	defineNative("map_contains_value", mapContainsValueNative)
	defineNative("map_size", mapSizeNative)
	defineNative("map_clear", mapClearNative)
	defineNative("map_keys", mapKeysNative)
	defineNative("map_values", mapValuesNative)

	// Date
	defineNative("Date", dateNew)
	defineNative("date_now", dateNow)
	defineNative("date_parse_datetime", dateParseDateTime)
	defineNative("date_format_datetime", dateFormatDateTime)
	defineNative("date_add_datetime", dateAddDateTime)
	defineNative("date_subtract_datetime", dateSubtractDateTime)
	defineNative("date_get_component", dateGetDateTimeComponent)
	defineNative("date_set_component", dateSetDateTimeComponent)
	defineNative("date_add_days", dateAddDays)
	defineNative("date_subtract_days", dateSubtractDays)

	// Time
	defineNative("Time", timeNew)
	defineNative("time_now", timeNow)
	defineNative("time_parse", timeParseTime)
	defineNative("time_format", timeFormatTime)
	defineNative("time_add", timeAddTime)
	defineNative("time_subtract", timeSubtractTime)
	defineNative("time_get_timezone", timeGetTimeZone)
	defineNative("time_convert_timezone", timeConvertTimeZone)

	// DateTime
	defineNative("DateTime", dateTimeNew)
	defineNative("datetime_now", dateTimeNow)
	defineNative("datetime_parse", dateTimeParseDateTime)
	defineNative("datetime_format", dateTimeFormatDateTime)
	defineNative("datetime_add", dateTimeAddDateTime)
	defineNative("datetime_subtract", dateTimeSubtractDateTime)
	defineNative("datetime_get_component", dateTimeGetDateTimeComponent)
	defineNative("datetime_set_component", dateTimeSetDateTimeComponent)
	defineNative("datetime_add_days", dateTimeAddDays)
	defineNative("datetime_subtract_days", dateTimeSubtractDays)

	// Random Functions
	defineNative("shuffle", shuffleNative)
	defineNative("random_between", randomBetweenNative)
	defineNative("random_string", randomStringNative)

	// Output Functions
	defineNative("print", printNative)
	defineNative("println", printlnNative)
	defineNative("printf", printfNative)

	// Input Functions
	defineNative("scan", scanNative)
	defineNative("scanln", scanlnNative)
	defineNative("scanf", scanfNative)

	// Formatting Functions
	defineNative("sprintf", sprintfNative)
	defineNative("errorf", errorfNative)

	// File Operations
	defineNative("read_file", readFileNative)
	defineNative("write_file", writeFileNative)

	// Utility Functions
	defineNative("parse_int", parseIntNative)

	// Others
	defineNative("clock", clockNative)
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

// ============================================================================
// Native Functions: Debug
// ============================================================================

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

// ============================================================================
// Native Functions: String Operations
// ============================================================================

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

// charAtNative returns character at given index
func charAtNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'char_at' expects 2 arguments: a string and an index.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'char_at' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_NUMBER {
		runtimeError("'char_at' requires a number as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	index := int(args[1].Number)
	if index < 0 || index >= len(strObj.Chars) {
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewObjString(string(strObj.Chars[index])))
}

// substringNative returns a substring between start and end indices
func substringNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 3 {
		runtimeError("'substring' expects 3 arguments: a string, start index, and end index.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'substring' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_NUMBER || args[2].Type != runtime.VAL_NUMBER {
		runtimeError("'substring' requires numbers as second and third arguments.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	start := int(args[1].Number)
	end := int(args[2].Number)
	if start < 0 {
		start = 0
	}
	if end > len(strObj.Chars) {
		end = len(strObj.Chars)
	}
	if start >= end || start >= len(strObj.Chars) {
		return runtime.ObjVal(runtime.NewObjString(""))
	}
	return runtime.ObjVal(runtime.NewObjString(strObj.Chars[start:end]))
}

// strIndexOfNative returns first occurrence of substring
func strIndexOfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'str_index_of' expects 2 arguments: a string and a substring.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'str_index_of' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	subStrObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("'str_index_of' requires a string as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	index := strings.Index(strObj.Chars, subStrObj.Chars)
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(index)}
}

// strLastIndexOfNative returns last occurrence of substring
func strLastIndexOfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'str_last_index_of' expects 2 arguments: a string and a substring.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'str_last_index_of' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	subStrObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("'str_last_index_of' requires a string as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	index := strings.LastIndex(strObj.Chars, subStrObj.Chars)
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(index)}
}

// strContainsNative checks if substring exists in string
func strContainsNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'str_contains' expects 2 arguments: a string and a substring.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'str_contains' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	subStrObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("'str_contains' requires a string as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{Type: runtime.VAL_BOOL, Bool: strings.Contains(strObj.Chars, subStrObj.Chars)}
}

// startsWithNative checks if string starts with prefix
func startsWithNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'starts_with' expects 2 arguments: a string and a prefix.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'starts_with' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	prefixObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("'starts_with' requires a string as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{Type: runtime.VAL_BOOL, Bool: strings.HasPrefix(strObj.Chars, prefixObj.Chars)}
}

// endsWithNative checks if string ends with suffix
func endsWithNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'ends_with' expects 2 arguments: a string and a suffix.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'ends_with' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	suffixObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("'ends_with' requires a string as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{Type: runtime.VAL_BOOL, Bool: strings.HasSuffix(strObj.Chars, suffixObj.Chars)}
}

// toUpperNative converts string to uppercase
func toUpperNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'to_upper' expects 1 argument: a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'to_upper' requires a string argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewObjString(strings.ToUpper(strObj.Chars)))
}

// toLowerNative converts string to lowercase
func toLowerNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'to_lower' expects 1 argument: a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'to_lower' requires a string argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewObjString(strings.ToLower(strObj.Chars)))
}

// trimNative removes leading and trailing whitespace
func trimNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'trim' expects 1 argument: a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'trim' requires a string argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewObjString(strings.TrimSpace(strObj.Chars)))
}

// splitNative splits string into array of substrings
func splitNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'split' expects 2 arguments: a string and a delimiter.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'split' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	delimiterObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("'split' requires a string as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	split := strings.Split(strObj.Chars, delimiterObj.Chars)
	result := make([]runtime.Value, len(split))
	for i, s := range split {
		result[i] = runtime.ObjVal(runtime.NewObjString(s))
	}
	return runtime.ObjVal(runtime.NewArray(result))
}

// replaceNative replaces all occurrences of old with new
func replaceNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 3 {
		runtimeError("'replace' expects 3 arguments: a string, old substring, and new substring.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'replace' requires a string as first argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	oldObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("'replace' requires a string as second argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	newObj, ok := args[2].Obj.(*runtime.ObjString)
	if !ok || args[2].Type != runtime.VAL_OBJ {
		runtimeError("'replace' requires a string as third argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewObjString(strings.ReplaceAll(strObj.Chars, oldObj.Chars, newObj.Chars)))
}

// strLengthNative returns the length of the string
func strLengthNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'str_length' expects 1 argument: a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("'str_length' requires a string argument.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(len(strObj.Chars))}
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

// ============================================================================
// Native Functions: Array Operations
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

// arrayIndexOfNative returns the index of the first occurrence of an element
func arrayIndexOfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'index_of' expects 2 arguments: an array and an element.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'index_of' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'index_of' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	element := args[1]
	for i, elem := range array.Elements {
		if runtime.Equal(elem, element) {
			return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(i)}
		}
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: -1}
}

// arrayLastIndexOfNative returns the index of the last occurrence of an element
func arrayLastIndexOfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'last_index_of' expects 2 arguments: an array and an element.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'last_index_of' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'last_index_of' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	element := args[1]
	for i := len(array.Elements) - 1; i >= 0; i-- {
		if runtime.Equal(array.Elements[i], element) {
			return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(i)}
		}
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: -1}
}

// arrayContainsNative checks if an element exists in the array
func arrayContainsNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'array_contains' expects 2 arguments: an array and an element.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_contains' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_contains' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	element := args[1]
	for _, elem := range array.Elements {
		if runtime.Equal(elem, element) {
			return runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
		}
	}
	return runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
}

// arrayClearNative removes all elements from the array
func arrayClearNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'array_clear' expects 1 argument: an array.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_clear' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_clear' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array.Elements = []runtime.Value{}
	return runtime.Value{Type: runtime.VAL_NULL}
}

// arrayReverseNative reverses the array in place
func arrayReverseNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'array_reverse' expects 1 argument: an array.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_reverse' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_reverse' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	for i, j := 0, len(array.Elements)-1; i < j; i, j = i+1, j-1 {
		array.Elements[i], array.Elements[j] = array.Elements[j], array.Elements[i]
	}
	return runtime.ObjVal(array)
}

// arrayToStringNative converts array to string representation
func arrayToStringNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'array_to_string' expects 1 argument: an array.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_to_string' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_to_string' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, elem := range array.Elements {
		if i > 0 {
			sb.WriteString(", ")
		}
		str := toStr(1, []runtime.Value{elem}).Obj.(*runtime.ObjString).Chars
		sb.WriteString(str)
	}
	sb.WriteString("]")
	return runtime.ObjVal(runtime.NewObjString(sb.String()))
}

// arrayRemoveNative removes the first occurrence of an element
func arrayRemoveNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'array_remove' expects 2 arguments: an array and an element.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'array_remove' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'array_remove' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	element := args[1]
	for i, elem := range array.Elements {
		if runtime.Equal(elem, element) {
			array.Elements = append(array.Elements[:i], array.Elements[i+1:]...)
			return runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
		}
	}
	return runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
}

// ============================================================================
// Native Functions: Map Operations
// ============================================================================

// mapRemoveNative removes a key-value pair from a map
func mapRemoveNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'map_remove' expects 2 arguments: a map and a key.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'map_remove' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj, ok := args[0].Obj.(*runtime.ObjMap)
	if !ok {
		runtimeError("'map_remove' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_OBJ {
		runtimeError("Map key must be a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	key, ok := args[1].Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("Map key must be a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	delete(mapObj.Entries, key)
	return runtime.Value{Type: runtime.VAL_NULL}
}

// mapContainsKeyNative checks if a key exists in a map
func mapContainsKeyNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'map_contains_key' expects 2 arguments: a map and a key.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'map_contains_key' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj, ok := args[0].Obj.(*runtime.ObjMap)
	if !ok {
		runtimeError("'map_contains_key' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_OBJ {
		runtimeError("Map key must be a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	key, ok := args[1].Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("Map key must be a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	_, exists := mapObj.Entries[key]
	return runtime.Value{Type: runtime.VAL_BOOL, Bool: exists}
}

// mapContainsValueNative checks if a value exists in a map
func mapContainsValueNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'map_contains_value' expects 2 arguments: a map and a value.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'map_contains_value' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj, ok := args[0].Obj.(*runtime.ObjMap)
	if !ok {
		runtimeError("'map_contains_value' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	searchVal := args[1]
	for _, val := range mapObj.Entries {
		if runtime.Equal(val, searchVal) {
			return runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
		}
	}
	return runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
}

// mapSizeNative returns the number of key-value pairs in a map
func mapSizeNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'map_size' expects 1 argument: a map.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'map_size' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj, ok := args[0].Obj.(*runtime.ObjMap)
	if !ok {
		runtimeError("'map_size' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(len(mapObj.Entries)),
	}
}

// mapClearNative removes all key-value pairs from a map
func mapClearNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'map_clear' expects 1 argument: a map.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'map_clear' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj, ok := args[0].Obj.(*runtime.ObjMap)
	if !ok {
		runtimeError("'map_clear' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj.Entries = make(map[*runtime.ObjString]runtime.Value)
	return runtime.Value{Type: runtime.VAL_NULL}
}

// mapKeysNative returns an array of all keys in a map
func mapKeysNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'map_keys' expects 1 argument: a map.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'map_keys' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj, ok := args[0].Obj.(*runtime.ObjMap)
	if !ok {
		runtimeError("'map_keys' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	keys := make([]runtime.Value, 0, len(mapObj.Entries))
	for key := range mapObj.Entries {
		keys = append(keys, runtime.ObjVal(key))
	}
	return runtime.ObjVal(runtime.NewArray(keys))
}

// mapValuesNative returns an array of all values in a map
func mapValuesNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'map_values' expects 1 argument: a map.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'map_values' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	mapObj, ok := args[0].Obj.(*runtime.ObjMap)
	if !ok {
		runtimeError("'map_values' can only be used on maps.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	values := make([]runtime.Value, 0, len(mapObj.Entries))
	for _, value := range mapObj.Entries {
		values = append(values, value)
	}
	return runtime.ObjVal(runtime.NewArray(values))
}

// ============================================================================
// Native Functions: Date
// ============================================================================

func dateNew(argCount int, args []runtime.Value) runtime.Value {
	switch argCount {
	case 0:
		// Return current date
		now := time.Now()
		year, month, day := now.Date()
		return runtime.ObjVal(runtime.NewDate(year, month, day))
	case 1:
		// Set year, default month to January (1), day to 1
		if args[0].Type != runtime.VAL_NUMBER {
			runtimeError("Argument must be a number")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
		year := int(args[0].Number)
		return runtime.ObjVal(runtime.NewDate(year, time.January, 1))
	case 3:
		// Set year, month, day
		for i := 0; i < 3; i++ {
			if args[i].Type != runtime.VAL_NUMBER {
				runtimeError("Arguments must be numbers")
				return runtime.Value{Type: runtime.VAL_NULL}
			}
		}
		year := int(args[0].Number)
		month := time.Month(args[1].Number) // Assumes month is 1-12
		day := int(args[2].Number)
		return runtime.ObjVal(runtime.NewDate(year, month, day))
	default:
		runtimeError("Date requires 0, 1, or 3 arguments")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
}

func dateNow(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		runtimeError("date_now() expects 0 arguments")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	now := time.Now()
	year, month, day := now.Date()
	return runtime.ObjVal(runtime.NewDate(year, month, day))
}

func dateParseDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("date_parse_datetime() expects 1 argument (string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_parse_datetime() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("date_parse_datetime() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	t, err := time.Parse("2006-01-02", strObj.Chars) // Standard date format
	if err != nil {
		runtimeError("Invalid date format: %s (use 'YYYY-MM-DD')", strObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewDate(t.Year(), t.Month(), t.Day()))
}

func dateFormatDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("date_format_datetime() expects 2 arguments (Date, format string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dateObj, ok := args[0].Obj.(*runtime.ObjDate)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_format_datetime() first argument must be a Date")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("date_format_datetime() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatted := dateObj.Time.Format(formatObj.Chars)
	return runtime.ObjVal(runtime.NewObjString(formatted))
}

func dateAddDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 4 {
		runtimeError("date_add_datetime() expects 4 arguments (Date, years, months, days)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dateObj, ok := args[0].Obj.(*runtime.ObjDate)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_add_datetime() first argument must be a Date")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	for i := 1; i < 4; i++ {
		if args[i].Type != runtime.VAL_NUMBER {
			runtimeError("date_add_datetime() arguments 2-4 must be numbers")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
	}
	years := int(args[1].Number)
	months := int(args[2].Number)
	days := int(args[3].Number)
	newTime := dateObj.Time.AddDate(years, months, days)
	return runtime.ObjVal(runtime.NewDate(newTime.Year(), newTime.Month(), newTime.Day()))
}

func dateSubtractDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 4 {
		runtimeError("date_subtract_datetime() expects 4 arguments (Date, years, months, days)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dateObj, ok := args[0].Obj.(*runtime.ObjDate)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_subtract_datetime() first argument must be a Date")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	for i := 1; i < 4; i++ {
		if args[i].Type != runtime.VAL_NUMBER {
			runtimeError("date_subtract_datetime() arguments 2-4 must be numbers")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
	}
	years := int(args[1].Number)
	months := int(args[2].Number)
	days := int(args[3].Number)
	newTime := dateObj.Time.AddDate(-years, -months, -days)
	return runtime.ObjVal(runtime.NewDate(newTime.Year(), newTime.Month(), newTime.Day()))
}

func dateGetDateTimeComponent(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("date_get_component() expects 2 arguments (Date, component string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dateObj, ok := args[0].Obj.(*runtime.ObjDate)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_get_component() first argument must be a Date")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	compObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("date_get_component() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	switch compObj.Chars {
	case "year":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dateObj.Time.Year())}
	case "month":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dateObj.Time.Month())}
	case "day":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dateObj.Time.Day())}
	default:
		runtimeError("Invalid component '%s' for Date (use 'year', 'month', 'day')", compObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
}

func dateSetDateTimeComponent(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 3 {
		runtimeError("date_set_component() expects 3 arguments (Date, component string, value)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dateObj, ok := args[0].Obj.(*runtime.ObjDate)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_set_component() first argument must be a Date")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	compObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("date_set_component() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[2].Type != runtime.VAL_NUMBER {
		runtimeError("date_set_component() third argument must be a number")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	value := int(args[2].Number)
	switch compObj.Chars {
	case "year":
		dateObj.Time = time.Date(value, dateObj.Time.Month(), dateObj.Time.Day(), 0, 0, 0, 0, dateObj.Time.Location())
	case "month":
		dateObj.Time = time.Date(dateObj.Time.Year(), time.Month(value), dateObj.Time.Day(), 0, 0, 0, 0, dateObj.Time.Location())
	case "day":
		dateObj.Time = time.Date(dateObj.Time.Year(), dateObj.Time.Month(), value, 0, 0, 0, 0, dateObj.Time.Location())
	default:
		runtimeError("Invalid component '%s' for Date (use 'year', 'month', 'day')", compObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(dateObj)
}

func dateAddDays(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("date_add_days() expects 2 arguments (Date, days)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dateObj, ok := args[0].Obj.(*runtime.ObjDate)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_add_days() first argument must be a Date")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_NUMBER {
		runtimeError("date_add_days() second argument must be a number")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	days := int(args[1].Number)
	newTime := dateObj.Time.AddDate(0, 0, days)
	return runtime.ObjVal(runtime.NewDate(newTime.Year(), newTime.Month(), newTime.Day()))
}

func dateSubtractDays(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("date_subtract_days() expects 2 arguments (Date, days)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dateObj, ok := args[0].Obj.(*runtime.ObjDate)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("date_subtract_days() first argument must be a Date")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_NUMBER {
		runtimeError("date_subtract_days() second argument must be a number")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	days := int(args[1].Number)
	newTime := dateObj.Time.AddDate(0, 0, -days)
	return runtime.ObjVal(runtime.NewDate(newTime.Year(), newTime.Month(), newTime.Day()))
}

// ============================================================================
// Native Functions: Time
// ============================================================================

func timeNew(argCount int, args []runtime.Value) runtime.Value {
	switch argCount {
	case 0:
		// Return current time
		now := time.Now()
		hour, minute, second := now.Hour(), now.Minute(), now.Second()
		return runtime.ObjVal(runtime.NewTime(hour, minute, second))
	case 1:
		// Set hour, default minute and second to 0
		if args[0].Type != runtime.VAL_NUMBER {
			runtimeError("Argument must be a number")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
		hour := int(args[0].Number)
		return runtime.ObjVal(runtime.NewTime(hour, 0, 0))
	case 3:
		// Set hour, minute, second
		for i := 0; i < 3; i++ {
			if args[i].Type != runtime.VAL_NUMBER {
				runtimeError("Arguments must be numbers")
				return runtime.Value{Type: runtime.VAL_NULL}
			}
		}
		hour := int(args[0].Number)
		minute := int(args[1].Number)
		second := int(args[2].Number)
		return runtime.ObjVal(runtime.NewTime(hour, minute, second))
	default:
		runtimeError("Time requires 0, 1, or 3 arguments")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
}

func timeNow(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		runtimeError("time_now() expects 0 arguments")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	now := time.Now()
	return runtime.ObjVal(runtime.NewTime(now.Hour(), now.Minute(), now.Second()))
}

func timeParseTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("time_parse() expects 1 argument (string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("time_parse() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("time_parse() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	t, err := time.Parse("15:04:05", strObj.Chars) // Standard time format
	if err != nil {
		runtimeError("Invalid time format: %s (use 'HH:MM:SS')", strObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewTime(t.Hour(), t.Minute(), t.Second()))
}

func timeFormatTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("time_format() expects 2 arguments (Time, format string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	timeObj, ok := args[0].Obj.(*runtime.ObjTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("time_format() first argument must be a Time")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("time_format() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatted := timeObj.Time.Format(formatObj.Chars)
	return runtime.ObjVal(runtime.NewObjString(formatted))
}

func timeAddTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 4 {
		runtimeError("time_add() expects 4 arguments (Time, hours, minutes, seconds)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	timeObj, ok := args[0].Obj.(*runtime.ObjTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("time_add() first argument must be a Time")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	for i := 1; i < 4; i++ {
		if args[i].Type != runtime.VAL_NUMBER {
			runtimeError("time_add() arguments 2-4 must be numbers")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
	}
	hours := int(args[1].Number)
	minutes := int(args[2].Number)
	seconds := int(args[3].Number)
	duration := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	newTime := timeObj.Time.Add(duration)
	return runtime.ObjVal(runtime.NewTime(newTime.Hour(), newTime.Minute(), newTime.Second()))
}

func timeSubtractTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 4 {
		runtimeError("time_subtract() expects 4 arguments (Time, hours, minutes, seconds)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	timeObj, ok := args[0].Obj.(*runtime.ObjTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("time_subtract() first argument must be a Time")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	for i := 1; i < 4; i++ {
		if args[i].Type != runtime.VAL_NUMBER {
			runtimeError("time_subtract() arguments 2-4 must be numbers")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
	}
	hours := int(args[1].Number)
	minutes := int(args[2].Number)
	seconds := int(args[3].Number)
	duration := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	newTime := timeObj.Time.Add(-duration)
	return runtime.ObjVal(runtime.NewTime(newTime.Hour(), newTime.Minute(), newTime.Second()))
}

func timeGetTimeZone(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("time_get_timezone() expects 1 argument (Time)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	timeObj, ok := args[0].Obj.(*runtime.ObjTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("time_get_timezone() argument must be a Time")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	zone, _ := timeObj.Time.Zone()
	return runtime.ObjVal(runtime.NewObjString(zone))
}

func timeConvertTimeZone(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("time_convert_timezone() expects 2 arguments (Time, timezone string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	timeObj, ok := args[0].Obj.(*runtime.ObjTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("time_convert_timezone() first argument must be a Time")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	tzObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("time_convert_timezone() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	loc, err := time.LoadLocation(tzObj.Chars)
	if err != nil {
		runtimeError("Invalid timezone: %s", tzObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	newTime := timeObj.Time.In(loc)
	return runtime.ObjVal(runtime.NewTime(newTime.Hour(), newTime.Minute(), newTime.Second()))
}

// ============================================================================
// Native Functions: DateTime
// ============================================================================

func dateTimeNew(argCount int, args []runtime.Value) runtime.Value {
	switch argCount {
	case 0:
		// Return current datetime
		now := time.Now()
		year, month, day := now.Date()
		hour, minute, second := now.Hour(), now.Minute(), now.Second()
		return runtime.ObjVal(runtime.NewDateTime(year, month, day, hour, minute, second))
	case 1:
		// Set year, default rest to minimal values
		if args[0].Type != runtime.VAL_NUMBER {
			runtimeError("Argument must be a number")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
		year := int(args[0].Number)
		return runtime.ObjVal(runtime.NewDateTime(year, time.January, 1, 0, 0, 0))
	case 6:
		// Set year, month, day, hour, minute, second
		for i := 0; i < 6; i++ {
			if args[i].Type != runtime.VAL_NUMBER {
				runtimeError("Arguments must be numbers")
				return runtime.Value{Type: runtime.VAL_NULL}
			}
		}
		year := int(args[0].Number)
		month := time.Month(args[1].Number) // Assumes month is 1-12
		day := int(args[2].Number)
		hour := int(args[3].Number)
		minute := int(args[4].Number)
		second := int(args[5].Number)
		return runtime.ObjVal(runtime.NewDateTime(year, month, day, hour, minute, second))
	default:
		runtimeError("DateTime requires 0, 1, or 6 arguments")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
}

func dateTimeNow(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		runtimeError("datetime_now() expects 0 arguments")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	now := time.Now()
	return runtime.ObjVal(runtime.NewDateTime(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()))
}

func dateTimeParseDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("datetime_parse() expects 1 argument (string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_parse() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("datetime_parse() expects a string argument")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	t, err := time.Parse("2006-01-02 15:04:05", strObj.Chars) // Standard datetime format
	if err != nil {
		runtimeError("Invalid datetime format: %s (use 'YYYY-MM-DD HH:MM:SS')", strObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewDateTime(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()))
}

func dateTimeFormatDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("datetime_format() expects 2 arguments (DateTime, format string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dtObj, ok := args[0].Obj.(*runtime.ObjDateTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_format() first argument must be a DateTime")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("datetime_format() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatted := dtObj.Time.Format(formatObj.Chars)
	return runtime.ObjVal(runtime.NewObjString(formatted))
}

func dateTimeAddDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 7 {
		runtimeError("datetime_add() expects 7 arguments (DateTime, years, months, days, hours, minutes, seconds)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dtObj, ok := args[0].Obj.(*runtime.ObjDateTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_add() first argument must be a DateTime")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	for i := 1; i < 7; i++ {
		if args[i].Type != runtime.VAL_NUMBER {
			runtimeError("datetime_add() arguments 2-7 must be numbers")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
	}
	years := int(args[1].Number)
	months := int(args[2].Number)
	days := int(args[3].Number)
	hours := int(args[4].Number)
	minutes := int(args[5].Number)
	seconds := int(args[6].Number)
	newTime := dtObj.Time.AddDate(years, months, days).Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second)
	return runtime.ObjVal(runtime.NewDateTime(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), newTime.Second()))
}

func dateTimeSubtractDateTime(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 7 {
		runtimeError("datetime_subtract() expects 7 arguments (DateTime, years, months, days, hours, minutes, seconds)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dtObj, ok := args[0].Obj.(*runtime.ObjDateTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_subtract() first argument must be a DateTime")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	for i := 1; i < 7; i++ {
		if args[i].Type != runtime.VAL_NUMBER {
			runtimeError("datetime_subtract() arguments 2-7 must be numbers")
			return runtime.Value{Type: runtime.VAL_NULL}
		}
	}
	years := int(args[1].Number)
	months := int(args[2].Number)
	days := int(args[3].Number)
	hours := int(args[4].Number)
	minutes := int(args[5].Number)
	seconds := int(args[6].Number)
	newTime := dtObj.Time.AddDate(-years, -months, -days).Add(-time.Duration(hours)*time.Hour - time.Duration(minutes)*time.Minute - time.Duration(seconds)*time.Second)
	return runtime.ObjVal(runtime.NewDateTime(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), newTime.Second()))
}

func dateTimeGetDateTimeComponent(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("datetime_get_component() expects 2 arguments (DateTime, component string)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dtObj, ok := args[0].Obj.(*runtime.ObjDateTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_get_component() first argument must be a DateTime")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	compObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("datetime_get_component() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	switch compObj.Chars {
	case "year":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dtObj.Time.Year())}
	case "month":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dtObj.Time.Month())}
	case "day":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dtObj.Time.Day())}
	case "hour":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dtObj.Time.Hour())}
	case "minute":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dtObj.Time.Minute())}
	case "second":
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(dtObj.Time.Second())}
	default:
		runtimeError("Invalid component '%s' for DateTime (use 'year', 'month', 'day', 'hour', 'minute', 'second')", compObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
}

func dateTimeSetDateTimeComponent(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 3 {
		runtimeError("datetime_set_component() expects 3 arguments (DateTime, component string, value)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dtObj, ok := args[0].Obj.(*runtime.ObjDateTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_set_component() first argument must be a DateTime")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	compObj, ok := args[1].Obj.(*runtime.ObjString)
	if !ok || args[1].Type != runtime.VAL_OBJ {
		runtimeError("datetime_set_component() second argument must be a string")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[2].Type != runtime.VAL_NUMBER {
		runtimeError("datetime_set_component() third argument must be a number")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	value := int(args[2].Number)
	switch compObj.Chars {
	case "year":
		dtObj.Time = time.Date(value, dtObj.Time.Month(), dtObj.Time.Day(), dtObj.Time.Hour(), dtObj.Time.Minute(), dtObj.Time.Second(), dtObj.Time.Nanosecond(), dtObj.Time.Location())
	case "month":
		dtObj.Time = time.Date(dtObj.Time.Year(), time.Month(value), dtObj.Time.Day(), dtObj.Time.Hour(), dtObj.Time.Minute(), dtObj.Time.Second(), dtObj.Time.Nanosecond(), dtObj.Time.Location())
	case "day":
		dtObj.Time = time.Date(dtObj.Time.Year(), dtObj.Time.Month(), value, dtObj.Time.Hour(), dtObj.Time.Minute(), dtObj.Time.Second(), dtObj.Time.Nanosecond(), dtObj.Time.Location())
	case "hour":
		dtObj.Time = time.Date(dtObj.Time.Year(), dtObj.Time.Month(), dtObj.Time.Day(), value, dtObj.Time.Minute(), dtObj.Time.Second(), dtObj.Time.Nanosecond(), dtObj.Time.Location())
	case "minute":
		dtObj.Time = time.Date(dtObj.Time.Year(), dtObj.Time.Month(), dtObj.Time.Day(), dtObj.Time.Hour(), value, dtObj.Time.Second(), dtObj.Time.Nanosecond(), dtObj.Time.Location())
	case "second":
		dtObj.Time = time.Date(dtObj.Time.Year(), dtObj.Time.Month(), dtObj.Time.Day(), dtObj.Time.Hour(), dtObj.Time.Minute(), value, dtObj.Time.Nanosecond(), dtObj.Time.Location())
	default:
		runtimeError("Invalid component '%s' for DateTime (use 'year', 'month', 'day', 'hour', 'minute', 'second')", compObj.Chars)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(dtObj)
}

func dateTimeAddDays(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("datetime_add_days() expects 2 arguments (DateTime, days)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dtObj, ok := args[0].Obj.(*runtime.ObjDateTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_add_days() first argument must be a DateTime")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_NUMBER {
		runtimeError("datetime_add_days() second argument must be a number")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	days := int(args[1].Number)
	newTime := dtObj.Time.AddDate(0, 0, days)
	return runtime.ObjVal(runtime.NewDateTime(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), newTime.Second()))
}

func dateTimeSubtractDays(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("datetime_subtract_days() expects 2 arguments (DateTime, days)")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	dtObj, ok := args[0].Obj.(*runtime.ObjDateTime)
	if !ok || args[0].Type != runtime.VAL_OBJ {
		runtimeError("datetime_subtract_days() first argument must be a DateTime")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[1].Type != runtime.VAL_NUMBER {
		runtimeError("datetime_subtract_days() second argument must be a number")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	days := int(args[1].Number)
	newTime := dtObj.Time.AddDate(0, 0, -days)
	return runtime.ObjVal(runtime.NewDateTime(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), newTime.Second()))
}

// ============================================================================
// Native Functions: Random Operations
// ============================================================================

func shuffleNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'shuffle' expects 1 argument (the array).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'shuffle' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	array, ok := args[0].Obj.(*runtime.ObjArray)
	if !ok {
		runtimeError("'shuffle' can only be used on arrays.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	rand.Shuffle(len(array.Elements), func(i, j int) {
		array.Elements[i], array.Elements[j] = array.Elements[j], array.Elements[i]
	})
	return args[0]
}

func randomBetweenNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'random_between' expects 2 arguments (min, max).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_NUMBER || args[1].Type != runtime.VAL_NUMBER {
		runtimeError("'random_between' expects two numbers.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	min := args[0].Number
	max := args[1].Number
	if min > max {
		runtimeError("min must be less than or equal to max.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if math.Floor(min) == min && math.Floor(max) == max {
		minInt := int(min)
		maxInt := int(max)
		randomInt := rand.Intn(maxInt-minInt+1) + minInt
		return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(randomInt)}
	}
	randomFloat := min + rand.Float64()*(max-min)
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: randomFloat}
}

func randomStringNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'random_string' expects 1 argument (size).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_NUMBER {
		runtimeError("'random_string' expects a number (size).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	size := int(args[0].Number)
	if size < 0 {
		runtimeError("Size must be non-negative.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	var sb strings.Builder
	for i := 0; i < size; i++ {
		index := rand.Intn(len(charset))
		sb.WriteByte(charset[index])
	}
	return runtime.ObjVal(runtime.NewObjString(sb.String()))
}

// ============================================================================
// Native Functions: Output Operations
// ============================================================================

func printNative(argCount int, args []runtime.Value) runtime.Value {
	for i := 0; i < argCount; i++ {
		strVal := toStr(1, args[i:i+1])
		if strVal.Type == runtime.VAL_OBJ {
			if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
				fmt.Print(unescapeString(strObj.Chars))
			} else {
				fmt.Print("error")
			}
		} else {
			fmt.Print("error")
		}
		if i < argCount-1 {
			fmt.Print(" ")
		}
	}
	return runtime.Value{Type: runtime.VAL_NULL}
}

func printlnNative(argCount int, args []runtime.Value) runtime.Value {
	for i := 0; i < argCount; i++ {
		strVal := toStr(1, args[i:i+1])
		if strVal.Type == runtime.VAL_OBJ {
			if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
				fmt.Print(unescapeString(strObj.Chars))
			} else {
				fmt.Print("error")
			}
		} else {
			fmt.Print("error")
		}
		if i < argCount-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
	return runtime.Value{Type: runtime.VAL_NULL}
}

func printfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount < 1 {
		runtimeError("'printf' expects at least 1 argument (format string).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatVal := args[0]
	if formatVal.Type != runtime.VAL_OBJ {
		runtimeError("'printf' first argument must be a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatObj, ok := formatVal.Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("'printf' first argument must be a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	format := unescapeString(formatObj.Chars) // Unescape the format string
	var printArgs []interface{}
	for _, arg := range args[1:] {
		switch arg.Type {
		case runtime.VAL_BOOL:
			printArgs = append(printArgs, arg.Bool)
		case runtime.VAL_NUMBER:
			if math.Mod(arg.Number, 1) == 0 {
				printArgs = append(printArgs, int(arg.Number))
			} else {
				printArgs = append(printArgs, arg.Number)
			}
		case runtime.VAL_OBJ:
			switch obj := arg.Obj.(type) {
			case *runtime.ObjString:
				printArgs = append(printArgs, unescapeString(obj.Chars)) // Unescape string arguments
			case *runtime.ObjArray:
				strVal := arrayToStringNative(1, []runtime.Value{arg})
				if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
					printArgs = append(printArgs, unescapeString(strObj.Chars))
				} else {
					printArgs = append(printArgs, "unknown array")
				}
			default:
				strVal := toStr(1, []runtime.Value{arg})
				if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
					printArgs = append(printArgs, unescapeString(strObj.Chars))
				} else {
					printArgs = append(printArgs, "unknown")
				}
			}
		case runtime.VAL_NULL:
			printArgs = append(printArgs, "null")
		default:
			printArgs = append(printArgs, "unknown")
		}
	}
	// Replace all format specifiers with %v
	adjustedFormat := strings.ReplaceAll(format, "%s", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%d", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%f", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%g", "%v")
	fmt.Printf(adjustedFormat, printArgs...)
	return runtime.Value{Type: runtime.VAL_NULL}
}

// ============================================================================
// Native Functions: Input Operations
// ============================================================================

func scanNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		runtimeError("'scan' expects 0 arguments.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		runtimeError("Error reading input: %v", err)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	line = strings.TrimSuffix(line, "\n")
	parts := strings.Fields(line)
	values := make([]runtime.Value, len(parts))
	for i, part := range parts {
		values[i] = runtime.ObjVal(runtime.NewObjString(part))
	}
	return runtime.ObjVal(runtime.NewArray(values))
}

func scanlnNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 0 {
		runtimeError("'scanln' expects 0 arguments.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		runtimeError("Error reading input: %v", err)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	line = strings.TrimSuffix(line, "\n")
	return runtime.ObjVal(runtime.NewObjString(line))
}

func scanfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'scanf' expects 1 argument (format string).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatVal := args[0]
	if formatVal.Type != runtime.VAL_OBJ {
		runtimeError("'scanf' expects a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if _, ok := formatVal.Obj.(*runtime.ObjString); !ok {
		runtimeError("'scanf' expects a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		runtimeError("Error reading input: %v", err)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	line = strings.TrimSuffix(line, "\n")
	return runtime.ObjVal(runtime.NewObjString(line))
}

// ============================================================================
// Native Functions: Format Operations
// ============================================================================

func sprintfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount < 1 {
		runtimeError("'sprintf' expects at least 1 argument (format string).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatVal := args[0]
	if formatVal.Type != runtime.VAL_OBJ {
		runtimeError("'sprintf' first argument must be a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatObj, ok := formatVal.Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("'sprintf' first argument must be a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	format := unescapeString(formatObj.Chars) // Unescape the format string
	var printArgs []interface{}
	for _, arg := range args[1:] {
		switch arg.Type {
		case runtime.VAL_BOOL:
			printArgs = append(printArgs, arg.Bool)
		case runtime.VAL_NUMBER:
			if math.Mod(arg.Number, 1) == 0 {
				printArgs = append(printArgs, int(arg.Number))
			} else {
				printArgs = append(printArgs, arg.Number)
			}
		case runtime.VAL_OBJ:
			switch obj := arg.Obj.(type) {
			case *runtime.ObjString:
				printArgs = append(printArgs, unescapeString(obj.Chars)) // Unescape string arguments
			case *runtime.ObjArray:
				strVal := arrayToStringNative(1, []runtime.Value{arg})
				if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
					printArgs = append(printArgs, unescapeString(strObj.Chars))
				} else {
					printArgs = append(printArgs, "unknown array")
				}
			default:
				strVal := toStr(1, []runtime.Value{arg})
				if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
					printArgs = append(printArgs, unescapeString(strObj.Chars))
				} else {
					printArgs = append(printArgs, "unknown")
				}
			}
		case runtime.VAL_NULL:
			printArgs = append(printArgs, "null")
		default:
			printArgs = append(printArgs, "unknown")
		}
	}
	// Replace all format specifiers with %v
	adjustedFormat := strings.ReplaceAll(format, "%s", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%d", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%f", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%g", "%v")
	formatted := fmt.Sprintf(adjustedFormat, printArgs...)
	return runtime.ObjVal(runtime.NewObjString(formatted))
}

func errorfNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount < 1 {
		runtimeError("'errorf' expects at least 1 argument (format string).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatVal := args[0]
	if formatVal.Type != runtime.VAL_OBJ {
		runtimeError("'errorf' first argument must be a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	formatObj, ok := formatVal.Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("'errorf' first argument must be a string (format).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	format := unescapeString(formatObj.Chars) // Unescape the format string
	var printArgs []interface{}
	for _, arg := range args[1:] {
		switch arg.Type {
		case runtime.VAL_BOOL:
			printArgs = append(printArgs, arg.Bool)
		case runtime.VAL_NUMBER:
			if math.Mod(arg.Number, 1) == 0 {
				printArgs = append(printArgs, int(arg.Number))
			} else {
				printArgs = append(printArgs, arg.Number)
			}
		case runtime.VAL_OBJ:
			switch obj := arg.Obj.(type) {
			case *runtime.ObjString:
				printArgs = append(printArgs, unescapeString(obj.Chars)) // Unescape string arguments
			case *runtime.ObjArray:
				strVal := arrayToStringNative(1, []runtime.Value{arg})
				if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
					printArgs = append(printArgs, unescapeString(strObj.Chars))
				} else {
					printArgs = append(printArgs, "unknown array")
				}
			default:
				strVal := toStr(1, []runtime.Value{arg})
				if strObj, ok := strVal.Obj.(*runtime.ObjString); ok {
					printArgs = append(printArgs, unescapeString(strObj.Chars))
				} else {
					printArgs = append(printArgs, "unknown")
				}
			}
		case runtime.VAL_NULL:
			printArgs = append(printArgs, "null")
		default:
			printArgs = append(printArgs, "unknown")
		}
	}
	// Replace all format specifiers with %v
	adjustedFormat := strings.ReplaceAll(format, "%s", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%d", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%f", "%v")
	adjustedFormat = strings.ReplaceAll(adjustedFormat, "%g", "%v")
	errMsg := fmt.Sprintf(adjustedFormat, printArgs...)
	return runtime.ObjVal(runtime.NewObjString(errMsg))
}

// ============================================================================
// Native Functions: File Operations
// ============================================================================

func readFileNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'read_file' expects 1 argument (file path).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	pathVal := args[0]
	if pathVal.Type != runtime.VAL_OBJ {
		runtimeError("'read_file' expects a string (file path).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	pathObj, ok := pathVal.Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("'read_file' expects a string (file path).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	path := pathObj.Chars
	content, err := os.ReadFile(path)
	if err != nil {
		runtimeError("Error reading file: %v", err)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.ObjVal(runtime.NewObjString(string(content)))
}

func writeFileNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 2 {
		runtimeError("'write_file' expects 2 arguments (file path, content).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	pathVal := args[0]
	contentVal := args[1]
	if pathVal.Type != runtime.VAL_OBJ {
		runtimeError("'write_file' first argument must be a string (file path).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	pathObj, ok := pathVal.Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("'write_file' first argument must be a string (file path).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if contentVal.Type != runtime.VAL_OBJ {
		runtimeError("'write_file' second argument must be a string (content).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	contentObj, ok := contentVal.Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("'write_file' second argument must be a string (content).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	path := pathObj.Chars
	content := contentObj.Chars
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		runtimeError("Error writing file: %v", err)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{Type: runtime.VAL_NULL}
}

// ============================================================================
// Native Functions: Utility Operations
// ============================================================================

func parseIntNative(argCount int, args []runtime.Value) runtime.Value {
	if argCount != 1 {
		runtimeError("'parse_int' expects 1 argument (string).")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	if args[0].Type != runtime.VAL_OBJ {
		runtimeError("'parse_int' expects a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	strObj, ok := args[0].Obj.(*runtime.ObjString)
	if !ok {
		runtimeError("'parse_int' expects a string.")
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	num, err := strconv.Atoi(strObj.Chars)
	if err != nil {
		runtimeError("Failed to parse '%s' as an integer: %v", strObj.Chars, err)
		return runtime.Value{Type: runtime.VAL_NULL}
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(num)}
}

// ============================================================================
// Native Functions: Others Operations
// ============================================================================

// clockNative returns the current time in seconds as a number.
func clockNative(argCount int, args []runtime.Value) runtime.Value {
	return runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(time.Now().UnixNano()) / 1e9,
	}
}
