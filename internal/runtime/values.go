package runtime

import (
	"fmt"
)

type ValueType int

const (
	VAL_BOOL ValueType = iota
	VAL_NULL
	VAL_NUMBER
	VAL_OBJ
)

type Value struct {
	Type   ValueType
	Bool   bool
	Number float64
	Obj    interface{}
}

type ValueArray struct {
	values   []Value
	count    int
	capacity int
}

func (va *ValueArray) Init() {
	va.values = nil
	va.count = 0
	va.capacity = 0
}

func (va *ValueArray) Write(val Value) {
	if va.count+1 > va.capacity {
		newCapacity := growCapacity(va.capacity)
		newValues := make([]Value, newCapacity)
		copy(newValues, va.values)
		va.values = newValues
		va.capacity = newCapacity
	}
	if va.count < len(va.values) {
		va.values[va.count] = val
	} else {
		va.values = append(va.values, val)
	}
	va.count++
}

func (va *ValueArray) Free() {
	va.values = nil
	va.count = 0
	va.capacity = 0
}

func (va *ValueArray) Count() int {
	return va.count
}

func (va *ValueArray) Values() []Value {
	return va.values[:va.count]
}

func PrintValue(v Value) {
	switch v.Type {
	case VAL_BOOL:
		fmt.Print(v.Bool)
	case VAL_NULL:
		fmt.Print("null")
	case VAL_NUMBER:
		fmt.Printf("%g", v.Number)
	case VAL_OBJ:
		PrintObject(v.Obj)
	}
}

func Equal(a, b Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case VAL_BOOL:
		return a.Bool == b.Bool
	case VAL_NULL:
		return true
	case VAL_NUMBER:
		return a.Number == b.Number
	case VAL_OBJ:
		aStr, okA := a.Obj.(*ObjString)
		bStr, okB := b.Obj.(*ObjString)
		if okA && okB {
			return aStr.Chars == bStr.Chars
		}
		return false
	default:
		return false
	}
}
