package vm

import (
	"math"
	"strings"

	"github.com/cryptrunner49/spy/internal/runtime"
)

// Helper function for numeric addition
func addNumbers(a, b runtime.Value) runtime.Value {
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number + b.Number}
}

// Helper function for string concatenation
func addStrings(a, b runtime.Value) runtime.Value {
	s1 := toStr(1, []runtime.Value{a}).Obj.(*runtime.ObjString).Chars
	s2 := toStr(1, []runtime.Value{b}).Obj.(*runtime.ObjString).Chars
	return runtime.ObjVal(runtime.NewObjString(s1 + s2))
}

// Helper function for array addition
func addArrays(arr1, arr2 *runtime.ObjArray) runtime.Value {
	len1 := len(arr1.Elements)
	len2 := len(arr2.Elements)
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}
	result := make([]runtime.Value, maxLen)
	minLen := len1
	if len2 < minLen {
		minLen = len2
	}
	for i := 0; i < minLen; i++ {
		v1 := arr1.Elements[i]
		v2 := arr2.Elements[i]
		if v1.Type == runtime.VAL_NUMBER && v2.Type == runtime.VAL_NUMBER {
			result[i] = addNumbers(v1, v2)
		} else {
			result[i] = addStrings(v1, v2)
		}
	}
	if len1 > len2 {
		for i := minLen; i < len1; i++ {
			result[i] = arr1.Elements[i]
		}
	} else {
		for i := minLen; i < len2; i++ {
			result[i] = arr2.Elements[i]
		}
	}
	return runtime.ObjVal(runtime.NewArray(result))
}

// Helper function for map addition
func addMaps(map1, map2 *runtime.ObjMap) runtime.Value {
	result := runtime.NewMap()
	for key, value := range map1.Entries {
		result.Entries[key] = value
	}
	for key, value := range map2.Entries {
		result.Entries[key] = value
	}
	return runtime.ObjVal(result)
}

// Helper function for struct instance addition
func addInstances(inst1, inst2 *runtime.ObjInstance) (runtime.Value, InterpretResult) {
	// Check if they are instances of the same struct
	if inst1.Structure != inst2.Structure {
		return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Cannot add instances of different structs ('%s' and '%s').", inst1.Structure.Name.Chars, inst2.Structure.Name.Chars)
	}

	result := runtime.NewInstance(inst1.Structure)
	for fieldName, val1 := range inst1.Fields {
		val2, exists := inst2.Fields[fieldName]
		if !exists {
			continue // Skip if field doesn't exist in second instance (shouldn't happen with same struct)
		}
		var fieldResult runtime.Value
		switch val1.Type {
		case runtime.VAL_NUMBER:
			if val2.Type == runtime.VAL_NUMBER {
				fieldResult = addNumbers(val1, val2)
			} else {
				return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for addition in struct: expected number, got %s.", typeName(val2))
			}
		case runtime.VAL_OBJ:
			if val2.Type != runtime.VAL_OBJ {
				return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for addition in struct: expected object, got %s.", typeName(val2))
			}
			switch v1 := val1.Obj.(type) {
			case *runtime.ObjString:
				if v2, ok := val2.Obj.(*runtime.ObjString); ok {
					fieldResult = addStrings(val1, runtime.ObjVal(v2))
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for addition in struct: expected string.")
				}
			case *runtime.ObjArray:
				if v2, ok := val2.Obj.(*runtime.ObjArray); ok {
					fieldResult = addArrays(v1, v2)
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for addition in struct: expected array.")
				}
			case *runtime.ObjMap:
				if v2, ok := val2.Obj.(*runtime.ObjMap); ok {
					fieldResult = addMaps(v1, v2)
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for addition in struct: expected map.")
				}
			case *runtime.ObjInstance:
				if v2, ok := val2.Obj.(*runtime.ObjInstance); ok {
					var err InterpretResult
					fieldResult, err = addInstances(v1, v2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for addition in struct: expected struct instance.")
				}
			default:
				return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Unsupported field type for addition in struct.")
			}
		default:
			return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for addition in struct.")
		}
		result.Fields[fieldName] = fieldResult
	}
	return runtime.ObjVal(result), INTERPRET_OK
}

// Helper function for numeric subtraction
func subtractNumbers(a, b runtime.Value) runtime.Value {
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number - b.Number}
}

// Helper function for string cropping
func subtractStrings(a, b runtime.Value) runtime.Value {
	s1 := toStr(1, []runtime.Value{a}).Obj.(*runtime.ObjString).Chars
	s2 := toStr(1, []runtime.Value{b}).Obj.(*runtime.ObjString).Chars
	idx := strings.Index(s1, s2)
	if idx >= 0 {
		return runtime.ObjVal(runtime.NewObjString(s1[:idx] + s1[idx+len(s2):]))
	}
	return runtime.ObjVal(runtime.NewObjString(s1))
}

// Helper function for array subtraction
func subtractArrays(arr1, arr2 *runtime.ObjArray) runtime.Value {
	len1 := len(arr1.Elements)
	len2 := len(arr2.Elements)
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}
	result := make([]runtime.Value, maxLen)
	minLen := len1
	if len2 < minLen {
		minLen = len2
	}
	for i := 0; i < minLen; i++ {
		v1 := arr1.Elements[i]
		v2 := arr2.Elements[i]
		if v1.Type == runtime.VAL_NUMBER && v2.Type == runtime.VAL_NUMBER {
			result[i] = subtractNumbers(v1, v2)
		} else {
			result[i] = subtractStrings(v1, v2)
		}
	}
	if len1 > len2 {
		for i := minLen; i < len1; i++ {
			result[i] = arr1.Elements[i]
		}
	} else {
		for i := minLen; i < len2; i++ {
			result[i] = arr2.Elements[i]
		}
	}
	return runtime.ObjVal(runtime.NewArray(result))
}

// Helper function for map subtraction
func subtractMaps(map1, map2 *runtime.ObjMap) runtime.Value {
	result := runtime.NewMap()
	for key, value := range map1.Entries {
		result.Entries[key] = value
	}
	for key := range map2.Entries {
		delete(result.Entries, key)
	}
	return runtime.ObjVal(result)
}

// Helper function for struct instance subtraction
func subtractInstances(inst1, inst2 *runtime.ObjInstance) (runtime.Value, InterpretResult) {
	// Check if they are instances of the same struct
	if inst1.Structure != inst2.Structure {
		return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Cannot subtract instances of different structs ('%s' and '%s').", inst1.Structure.Name.Chars, inst2.Structure.Name.Chars)
	}

	result := runtime.NewInstance(inst1.Structure)
	for fieldName, val1 := range inst1.Fields {
		val2, exists := inst2.Fields[fieldName]
		if !exists {
			continue // Skip if field doesn't exist in second instance (shouldn't happen with same struct)
		}
		var fieldResult runtime.Value
		switch val1.Type {
		case runtime.VAL_NUMBER:
			if val2.Type == runtime.VAL_NUMBER {
				fieldResult = subtractNumbers(val1, val2)
			} else {
				return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for subtraction in struct: expected number, got %s.", typeName(val2))
			}
		case runtime.VAL_OBJ:
			if val2.Type != runtime.VAL_OBJ {
				return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for subtraction in struct: expected object, got %s.", typeName(val2))
			}
			switch v1 := val1.Obj.(type) {
			case *runtime.ObjString:
				if v2, ok := val2.Obj.(*runtime.ObjString); ok {
					fieldResult = subtractStrings(val1, runtime.ObjVal(v2))
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for subtraction in struct: expected string.")
				}
			case *runtime.ObjArray:
				if v2, ok := val2.Obj.(*runtime.ObjArray); ok {
					fieldResult = subtractArrays(v1, v2)
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for subtraction in struct: expected array.")
				}
			case *runtime.ObjMap:
				if v2, ok := val2.Obj.(*runtime.ObjMap); ok {
					fieldResult = subtractMaps(v1, v2)
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for subtraction in struct: expected map.")
				}
			case *runtime.ObjInstance:
				if v2, ok := val2.Obj.(*runtime.ObjInstance); ok {
					var err InterpretResult
					fieldResult, err = subtractInstances(v1, v2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for subtraction in struct: expected struct instance.")
				}
			default:
				return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Unsupported field type for subtraction in struct.")
			}
		default:
			return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Incompatible field types for subtraction in struct.")
		}
		result.Fields[fieldName] = fieldResult
	}
	return runtime.ObjVal(result), INTERPRET_OK
}

// Helper function for numeric multiplication
func multiplyNumbers(a, b runtime.Value) runtime.Value {
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number * b.Number}
}

// Helper function for array multiplication
func multiplyArrays(arr1, arr2 *runtime.ObjArray) (runtime.Value, InterpretResult) {
	len1 := len(arr1.Elements)
	len2 := len(arr2.Elements)
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}
	result := make([]runtime.Value, maxLen)
	minLen := len1
	if len2 < minLen {
		minLen = len2
	}
	for i := 0; i < minLen; i++ {
		v1 := arr1.Elements[i]
		v2 := arr2.Elements[i]
		if v1.Type != runtime.VAL_NUMBER || v2.Type != runtime.VAL_NUMBER {
			return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Array elements for '*' must be numbers (found %s and %s at index %d).", typeName(v1), typeName(v2), i)
		}
		result[i] = multiplyNumbers(v1, v2)
	}
	if len1 > len2 {
		for i := minLen; i < len1; i++ {
			result[i] = arr1.Elements[i]
		}
	} else {
		for i := minLen; i < len2; i++ {
			result[i] = arr2.Elements[i]
		}
	}
	return runtime.ObjVal(runtime.NewArray(result)), INTERPRET_OK
}

// Helper function for struct instance multiplication
func multiplyInstances(inst1, inst2 *runtime.ObjInstance) (runtime.Value, InterpretResult) {
	if inst1.Structure != inst2.Structure {
		return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Cannot multiply instances of different structs ('%s' and '%s').", inst1.Structure.Name.Chars, inst2.Structure.Name.Chars)
	}
	result := runtime.NewInstance(inst1.Structure)
	for fieldName, val1 := range inst1.Fields {
		val2, exists := inst2.Fields[fieldName]
		if !exists {
			continue
		}
		var fieldResult runtime.Value
		switch val1.Type {
		case runtime.VAL_NUMBER:
			if val2.Type == runtime.VAL_NUMBER {
				fieldResult = multiplyNumbers(val1, val2)
			} else {
				fieldResult = val1
			}
		case runtime.VAL_OBJ:
			if arr1, ok := val1.Obj.(*runtime.ObjArray); ok {
				if arr2, ok := val2.Obj.(*runtime.ObjArray); ok {
					var err InterpretResult
					fieldResult, err = multiplyArrays(arr1, arr2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					fieldResult = val1
				}
			} else if inst1, ok := val1.Obj.(*runtime.ObjInstance); ok {
				if inst2, ok := val2.Obj.(*runtime.ObjInstance); ok {
					var err InterpretResult
					fieldResult, err = multiplyInstances(inst1, inst2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					fieldResult = val1
				}
			} else {
				fieldResult = val1
			}
		default:
			fieldResult = val1
		}
		result.Fields[fieldName] = fieldResult
	}
	return runtime.ObjVal(result), INTERPRET_OK
}

// Helper function for numeric division
func divideNumbers(a, b runtime.Value) runtime.Value {
	if b.Number == 0 {
		runtimeError("Division by zero in struct field operation.")
		return a
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: a.Number / b.Number}
}

// Helper function for array division
func divideArrays(arr1, arr2 *runtime.ObjArray) (runtime.Value, InterpretResult) {
	len1 := len(arr1.Elements)
	len2 := len(arr2.Elements)
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}
	result := make([]runtime.Value, maxLen)
	minLen := len1
	if len2 < minLen {
		minLen = len2
	}
	for i := 0; i < minLen; i++ {
		v1 := arr1.Elements[i]
		v2 := arr2.Elements[i]
		if v1.Type != runtime.VAL_NUMBER || v2.Type != runtime.VAL_NUMBER {
			return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Array elements for '/' must be numbers (found %s and %s at index %d).", typeName(v1), typeName(v2), i)
		}
		result[i] = divideNumbers(v1, v2)
	}
	if len1 > len2 {
		for i := minLen; i < len1; i++ {
			result[i] = arr1.Elements[i]
		}
	} else {
		for i := minLen; i < len2; i++ {
			result[i] = arr2.Elements[i]
		}
	}
	return runtime.ObjVal(runtime.NewArray(result)), INTERPRET_OK
}

// Helper function for struct instance division
func divideInstances(inst1, inst2 *runtime.ObjInstance) (runtime.Value, InterpretResult) {
	if inst1.Structure != inst2.Structure {
		return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Cannot divide instances of different structs ('%s' and '%s').", inst1.Structure.Name.Chars, inst2.Structure.Name.Chars)
	}
	result := runtime.NewInstance(inst1.Structure)
	for fieldName, val1 := range inst1.Fields {
		val2, exists := inst2.Fields[fieldName]
		if !exists {
			continue
		}
		var fieldResult runtime.Value
		switch val1.Type {
		case runtime.VAL_NUMBER:
			if val2.Type == runtime.VAL_NUMBER {
				fieldResult = divideNumbers(val1, val2)
			} else {
				fieldResult = val1
			}
		case runtime.VAL_OBJ:
			if arr1, ok := val1.Obj.(*runtime.ObjArray); ok {
				if arr2, ok := val2.Obj.(*runtime.ObjArray); ok {
					var err InterpretResult
					fieldResult, err = divideArrays(arr1, arr2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					fieldResult = val1
				}
			} else if inst1, ok := val1.Obj.(*runtime.ObjInstance); ok {
				if inst2, ok := val2.Obj.(*runtime.ObjInstance); ok {
					var err InterpretResult
					fieldResult, err = divideInstances(inst1, inst2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					fieldResult = val1
				}
			} else {
				fieldResult = val1
			}
		default:
			fieldResult = val1
		}
		result.Fields[fieldName] = fieldResult
	}
	return runtime.ObjVal(result), INTERPRET_OK
}

// Helper function for numeric modulo
func modNumbers(a, b runtime.Value) runtime.Value {
	if b.Number == 0 {
		runtimeError("Modulo by zero in struct field operation.")
		return a
	}
	return runtime.Value{Type: runtime.VAL_NUMBER, Number: math.Mod(a.Number, b.Number)}
}

// Helper function for array modulo
func modArrays(arr1, arr2 *runtime.ObjArray) (runtime.Value, InterpretResult) {
	len1 := len(arr1.Elements)
	len2 := len(arr2.Elements)
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}
	result := make([]runtime.Value, maxLen)
	minLen := len1
	if len2 < minLen {
		minLen = len2
	}
	for i := 0; i < minLen; i++ {
		v1 := arr1.Elements[i]
		v2 := arr2.Elements[i]
		if v1.Type != runtime.VAL_NUMBER || v2.Type != runtime.VAL_NUMBER {
			return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Array elements for '%%' must be numbers (found %s and %s at index %d).", typeName(v1), typeName(v2), i)
		}
		result[i] = modNumbers(v1, v2)
	}
	if len1 > len2 {
		for i := minLen; i < len1; i++ {
			result[i] = arr1.Elements[i]
		}
	} else {
		for i := minLen; i < len2; i++ {
			result[i] = arr2.Elements[i]
		}
	}
	return runtime.ObjVal(runtime.NewArray(result)), INTERPRET_OK
}

// Helper function for struct instance modulo
func modInstances(inst1, inst2 *runtime.ObjInstance) (runtime.Value, InterpretResult) {
	if inst1.Structure != inst2.Structure {
		return runtime.Value{Type: runtime.VAL_NULL}, runtimeError("Cannot compute modulo of instances of different structs ('%s' and '%s').", inst1.Structure.Name.Chars, inst2.Structure.Name.Chars)
	}
	result := runtime.NewInstance(inst1.Structure)
	for fieldName, val1 := range inst1.Fields {
		val2, exists := inst2.Fields[fieldName]
		if !exists {
			continue
		}
		var fieldResult runtime.Value
		switch val1.Type {
		case runtime.VAL_NUMBER:
			if val2.Type == runtime.VAL_NUMBER {
				fieldResult = modNumbers(val1, val2)
			} else {
				fieldResult = val1
			}
		case runtime.VAL_OBJ:
			if arr1, ok := val1.Obj.(*runtime.ObjArray); ok {
				if arr2, ok := val2.Obj.(*runtime.ObjArray); ok {
					var err InterpretResult
					fieldResult, err = modArrays(arr1, arr2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					fieldResult = val1
				}
			} else if inst1, ok := val1.Obj.(*runtime.ObjInstance); ok {
				if inst2, ok := val2.Obj.(*runtime.ObjInstance); ok {
					var err InterpretResult
					fieldResult, err = modInstances(inst1, inst2)
					if err != INTERPRET_OK {
						return runtime.Value{Type: runtime.VAL_NULL}, err
					}
				} else {
					fieldResult = val1
				}
			} else {
				fieldResult = val1
			}
		default:
			fieldResult = val1
		}
		result.Fields[fieldName] = fieldResult
	}
	return runtime.ObjVal(result), INTERPRET_OK
}
