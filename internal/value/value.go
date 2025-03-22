package value

import "fmt"

type Value float64

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

func (va *ValueArray) Write(value Value) {
	if va.count+1 > va.capacity {
		newCapacity := growCapacity(va.capacity)
		newValues := make([]Value, newCapacity)
		copy(newValues, va.values)
		va.values = newValues
		va.capacity = newCapacity
	}
	va.values[va.count] = value
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

func PrintValue(value Value) {
	fmt.Printf("%g", value)
}

func growCapacity(capacity int) int {
	if capacity < 8 {
		return 8
	}
	return capacity * 2
}
