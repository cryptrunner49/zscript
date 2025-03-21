package chunk

import (
	"errors"

	"github.com/cryptrunner49/gorex/internal/value"
)

type OpCode byte

const (
	OP_CONSTANT OpCode = iota
	OP_RETURN
)

type Chunk struct {
	code      []OpCode
	lines     []int
	count     int
	capacity  int
	constants value.ValueArray
}

// New creates a new Chunk with zero values
func New() *Chunk {
	c := &Chunk{}
	c.init()
	return c
}

func (c *Chunk) init() {
	c.code = nil
	c.lines = nil
	c.count = 0
	c.capacity = 0
	c.constants.Init()
}

// Write adds a byte and line number to the chunk, growing if necessary
func (c *Chunk) Write(opCode OpCode, line int) error {
	if c.count+1 > c.capacity {
		if err := c.grow(); err != nil {
			return err
		}
	}

	c.code[c.count] = opCode
	c.lines[c.count] = line
	c.count++
	return nil
}

// AddConstant adds a value to constants and returns its index
func (c *Chunk) AddConstant(val value.Value) int {
	c.constants.Write(val)
	return c.constants.Count() - 1
}

// Free resets the chunk to initial state
func (c *Chunk) Free() {
	c.constants.Free()
	c.init()
}

// Count returns the number of bytes in the chunk
func (c *Chunk) Count() int {
	return c.count
}

// Code returns a read-only slice of the bytes
func (c *Chunk) Code() []OpCode {
	return c.code[:c.count]
}

// Lines returns a read-only slice of the line numbers
func (c *Chunk) Lines() []int {
	return c.lines[:c.count]
}

// Constants returns the constants array
func (c *Chunk) Constants() *value.ValueArray {
	return &c.constants
}

func (c *Chunk) grow() error {
	newCapacity := growCapacity(c.capacity)
	if newCapacity < 0 {
		return errors.New("capacity overflow")
	}

	newCode := make([]OpCode, newCapacity)
	newLines := make([]int, newCapacity)
	copy(newCode, c.code)
	copy(newLines, c.lines)
	c.code = newCode
	c.lines = newLines
	c.capacity = newCapacity
	return nil
}

func growCapacity(capacity int) int {
	if capacity < 8 {
		return 8
	}
	return capacity * 2
}
