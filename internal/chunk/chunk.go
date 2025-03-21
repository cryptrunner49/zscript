package chunk

import "errors"

type OpCode uint8

const (
	OP_RETURN OpCode = iota
)

type Chunk struct {
	code     []OpCode
	count    int
	capacity int
}

// New creates a new Chunk with zero values
func New() *Chunk {
	return &Chunk{}
}

// Write adds an opcode to the chunk, growing if necessary
func (c *Chunk) Write(op OpCode) error {
	if c.count >= c.capacity {
		if err := c.grow(); err != nil {
			return err
		}
	}

	c.code[c.count] = op
	c.count++
	return nil
}

// Free resets the chunk to initial state
func (c *Chunk) Free() {
	c.code = nil
	c.count = 0
	c.capacity = 0
}

// Count returns the number of opcodes in the chunk
func (c *Chunk) Count() int {
	return c.count
}

// Code returns a read-only slice of the opcodes
func (c *Chunk) Code() []OpCode {
	return c.code[:c.count]
}

func (c *Chunk) grow() error {
	newCapacity := growCapacity(c.capacity)
	if newCapacity < 0 {
		return errors.New("capacity overflow")
	}

	newCode := make([]OpCode, newCapacity)
	copy(newCode, c.code)
	c.code = newCode
	c.capacity = newCapacity
	return nil
}

func growCapacity(capacity int) int {
	if capacity < 8 {
		return 8
	}
	return capacity * 2
}
