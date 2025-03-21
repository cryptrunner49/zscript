package value

import "fmt"

type Value float64

type ValueArray struct {
	values   []Value
	count    int
	capacity int
}

// Init initializes a ValueArray with zero values
func (va *ValueArray) Init() {
	va.values = nil
	va.count = 0
	va.capacity = 0
}

// Write adds a value to the array, growing if necessary
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

// Free resets the array to initial state
func (va *ValueArray) Free() {
	va.values = nil
	va.count = 0
	va.capacity = 0
}

// Count returns the number of values in the array
func (va *ValueArray) Count() int {
	return va.count
}

// Values returns a read-only slice of the values
func (va *ValueArray) Values() []Value {
	return va.values[:va.count]
}

// PrintValue prints a value (matches C's printValue)
func PrintValue(value Value) {
	fmt.Printf("%g", value)
}

func growCapacity(capacity int) int {
	if capacity < 8 {
		return 8
	}
	return capacity * 2
}
