package debug

import (
	"fmt"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/value"
)

func Disassemble(ch *chunk.Chunk, name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < ch.Count(); {
		offset = disassembleInstruction(ch, offset)
	}
}

func disassembleInstruction(ch *chunk.Chunk, offset int) int {
	fmt.Printf("%04d ", offset)

	if offset > 0 && ch.Lines()[offset] == ch.Lines()[offset-1] {
		fmt.Print("   | ")
	} else {
		fmt.Printf("%4d ", ch.Lines()[offset])
	}

	instruction := ch.Code()[offset]
	switch instruction {
	case chunk.OP_CONSTANT:
		return constantInstruction("OP_CONSTANT", ch, offset)
	case chunk.OP_RETURN:
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
