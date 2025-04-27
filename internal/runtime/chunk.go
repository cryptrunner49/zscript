package runtime

import (
	"errors"
)

type Chunk struct {
	code      []uint8
	lines     []int
	count     int
	capacity  int
	constants ValueArray
}

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

func (c *Chunk) Write(byte uint8, line int) error {
	if c.count+1 > c.capacity {
		if err := c.grow(); err != nil {
			return err
		}
	}
	c.code[c.count] = byte
	c.lines[c.count] = line
	c.count++
	return nil
}

func (c *Chunk) AddConstant(val Value) int {
	c.constants.Write(val)
	return c.constants.Count() - 1
}

func (c *Chunk) Free() {
	c.constants.Free()
	c.init()
}

func (c *Chunk) Count() int {
	return c.count
}

func (c *Chunk) Code() []uint8 {
	return c.code[:c.count]
}

func (c *Chunk) Lines() []int {
	return c.lines[:c.count]
}

func (c *Chunk) Constants() *ValueArray {
	return &c.constants
}

func (c *Chunk) grow() error {
	newCapacity := growCapacity(c.capacity)
	if newCapacity < 0 {
		return errors.New("capacity overflow")
	}
	newCode := make([]uint8, newCapacity)
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
