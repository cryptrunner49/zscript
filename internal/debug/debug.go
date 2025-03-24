package debug

import (
	"fmt"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/value"
)

func Disassemble(ch *chunk.Chunk, name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < ch.Count(); {
		offset = DisassembleInstruction(ch, offset)
	}
}

func DisassembleInstruction(ch *chunk.Chunk, offset int) int {
	fmt.Printf("%04d ", offset)
	if offset > 0 && ch.Lines()[offset] == ch.Lines()[offset-1] {
		fmt.Print("   | ")
	} else {
		fmt.Printf("%4d ", ch.Lines()[offset])
	}

	instruction := ch.Code()[offset]
	switch instruction {
	case uint8(chunk.OP_CONSTANT):
		return constantInstruction("OP_CONSTANT", ch, offset)
	case uint8(chunk.OP_NULL):
		return simpleInstruction("OP_NULL", offset)
	case uint8(chunk.OP_TRUE):
		return simpleInstruction("OP_TRUE", offset)
	case uint8(chunk.OP_FALSE):
		return simpleInstruction("OP_FALSE", offset)
	case uint8(chunk.OP_POP):
		return simpleInstruction("OP_POP", offset)
	case uint8(chunk.OP_SET_LOCAL):
		return byteInstruction("OP_SET_LOCAL", ch, offset)
	case uint8(chunk.OP_GET_LOCAL):
		return byteInstruction("OP_GET_LOCAL", ch, offset)
	case uint8(chunk.OP_DEFINE_GLOBAL):
		return constantInstruction("OP_DEFINE_GLOBAL", ch, offset)
	case uint8(chunk.OP_SET_GLOBAL):
		return constantInstruction("OP_SET_GLOBAL", ch, offset)
	case uint8(chunk.OP_GET_GLOBAL):
		return constantInstruction("OP_GET_GLOBAL", ch, offset)
	case uint8(chunk.OP_EQUAL):
		return simpleInstruction("OP_EQUAL", offset)
	case uint8(chunk.OP_GREATER):
		return simpleInstruction("OP_GREATER", offset)
	case uint8(chunk.OP_LESS):
		return simpleInstruction("OP_LESS", offset)
	case uint8(chunk.OP_ADD):
		return simpleInstruction("OP_ADD", offset)
	case uint8(chunk.OP_SUBTRACT):
		return simpleInstruction("OP_SUBTRACT", offset)
	case uint8(chunk.OP_MULTIPLY):
		return simpleInstruction("OP_MULTIPLY", offset)
	case uint8(chunk.OP_DIVIDE):
		return simpleInstruction("OP_DIVIDE", offset)
	case uint8(chunk.OP_NOT):
		return simpleInstruction("OP_NOT", offset)
	case uint8(chunk.OP_NEGATE):
		return simpleInstruction("OP_NEGATE", offset)
	case uint8(chunk.OP_PRINT):
		return simpleInstruction("OP_PRINT", offset)
	case uint8(chunk.OP_RETURN):
		return simpleInstruction("OP_RETURN", offset)
	default:
		fmt.Printf("Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

func simpleInstruction(name string, offset int) int {
	fmt.Println(name)
	return offset + 1
}

func constantInstruction(name string, ch *chunk.Chunk, offset int) int {
	constant := ch.Code()[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	value.PrintValue(ch.Constants().Values()[constant])
	fmt.Println("'")
	return offset + 2
}

func byteInstruction(name string, ch *chunk.Chunk, offset int) int {
	slot := ch.Code()[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot)
	return offset + 2
}
