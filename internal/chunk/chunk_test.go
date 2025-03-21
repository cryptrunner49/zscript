package chunk

import (
	"testing"
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
}

func TestWrite(t *testing.T) {
	c := New()
	if err := c.Write(OP_RETURN); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if c.count != 1 {
		t.Errorf("Expected count 1, got %d", c.count)
	}
	if c.Code()[0] != OP_RETURN {
		t.Errorf("Expected OP_RETURN, got %d", c.Code()[0])
	}
}

func TestGrow(t *testing.T) {
	c := New()
	for i := 0; i < 10; i++ {
		if err := c.Write(OP_RETURN); err != nil {
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
	c.Write(OP_RETURN)
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
}
