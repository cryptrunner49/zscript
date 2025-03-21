package value

import (
	"testing"
)

func TestInit(t *testing.T) {
	var va ValueArray
	va.Init()

	if va.count != 0 {
		t.Errorf("Expected count 0, got %d", va.count)
	}
	if va.capacity != 0 {
		t.Errorf("Expected capacity 0, got %d", va.capacity)
	}
	if va.values != nil {
		t.Errorf("Expected nil values, got %v", va.values)
	}
}

func TestWrite(t *testing.T) {
	var va ValueArray
	va.Init()
	va.Write(Value(1.2))

	if va.count != 1 {
		t.Errorf("Expected count 1, got %d", va.count)
	}
	if va.capacity < 1 {
		t.Errorf("Expected capacity >= 1, got %d", va.capacity)
	}
	if va.Values()[0] != 1.2 {
		t.Errorf("Expected value 1.2, got %f", va.Values()[0])
	}
}

func TestGrow(t *testing.T) {
	var va ValueArray
	va.Init()
	for i := 0; i < 10; i++ {
		va.Write(Value(float64(i)))
	}

	if va.count != 10 {
		t.Errorf("Expected count 10, got %d", va.count)
	}
	if va.capacity < 10 {
		t.Errorf("Expected capacity >= 10, got %d", va.capacity)
	}
	for i := 0; i < 10; i++ {
		if va.Values()[i] != Value(float64(i)) {
			t.Errorf("Expected value %d at index %d, got %f", i, i, va.Values()[i])
		}
	}
}

func TestFree(t *testing.T) {
	var va ValueArray
	va.Init()
	va.Write(Value(1.2))
	va.Free()

	if va.count != 0 {
		t.Errorf("Expected count 0, got %d", va.count)
	}
	if va.capacity != 0 {
		t.Errorf("Expected capacity 0, got %d", va.capacity)
	}
	if va.values != nil {
		t.Errorf("Expected nil values, got %v", va.values)
	}
}

func TestPrintValue(t *testing.T) {
	// This test would ideally capture stdout, but for simplicity we'll just call it
	// to ensure it doesn't panic. A full test would use a bytes.Buffer to capture output.
	PrintValue(Value(1.2)) // Should print "1.2" without panicking
}
