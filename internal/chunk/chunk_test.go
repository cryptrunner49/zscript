package chunk

import (
	"testing"

	"github.com/cryptrunner49/gorex/internal/value"
)

func TestNew(t *testing.T) {
	c := New()
	if c.count != 0 {
		t.Errorf("Expected count 0, got %d", c.count)
	}
	if c.capacity != 0 {
		t.Errorf("Expected capacity 0, got %d", c.capacity)
	}
	if c.code != nil {
		t.Errorf("Expected nil code, got %v", c.code)
	}
	if c.lines != nil {
		t.Errorf("Expected nil lines, got %v", c.lines)
	}
	if c.constants.Count() != 0 {
		t.Errorf("Expected constants count 0, got %d", c.constants.Count())
	}
}

func TestWrite(t *testing.T) {
	c := New()
	if err := c.Write(uint8(OP_RETURN), 123); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if c.count != 1 {
		t.Errorf("Expected count 1, got %d", c.count)
	}
	if c.Code()[0] != uint8(OP_RETURN) {
		t.Errorf("Expected OP_RETURN, got %d", c.Code()[0])
	}
	if c.Lines()[0] != 123 {
		t.Errorf("Expected line 123, got %d", c.Lines()[0])
	}
}

func TestAddConstant(t *testing.T) {
	c := New()
	idx := c.AddConstant(value.Value(1.2))
	if idx != 0 {
		t.Errorf("Expected index 0, got %d", idx)
	}
	if c.constants.Count() != 1 {
		t.Errorf("Expected constants count 1, got %d", c.constants.Count())
	}
	if c.constants.Values()[0] != 1.2 {
		t.Errorf("Expected constant 1.2, got %f", c.constants.Values()[0])
	}
}

func TestGrow(t *testing.T) {
	c := New()
	for i := 0; i < 10; i++ {
		if err := c.Write(uint8(OP_RETURN), 123); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}
	if c.capacity < 10 {
		t.Errorf("Expected capacity >= 10, got %d", c.capacity)
	}
	if c.count != 10 {
		t.Errorf("Expected count 10, got %d", c.count)
	}
}

func TestFree(t *testing.T) {
	c := New()
	if err := c.Write(uint8(OP_RETURN), 123); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	c.AddConstant(value.Value(1.2))
	c.Free()
	if c.count != 0 {
		t.Errorf("Expected count 0, got %d", c.count)
	}
	if c.capacity != 0 {
		t.Errorf("Expected capacity 0, got %d", c.capacity)
	}
	if c.code != nil {
		t.Errorf("Expected nil code, got %v", c.code)
	}
	if c.lines != nil {
		t.Errorf("Expected nil lines, got %v", c.lines)
	}
	if c.constants.Count() != 0 {
		t.Errorf("Expected constants count 0, got %d", c.constants.Count())
	}
}
