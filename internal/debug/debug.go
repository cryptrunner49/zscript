package debug

import (
	"fmt"

	"github.com/cryptrunner49/gorex/internal/chunk"
)

func Disassemble(ch *chunk.Chunk, name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < ch.Count(); {
		offset = disassembleInstruction(ch, offset)
	}
}

func disassembleInstruction(ch *chunk.Chunk, offset int) int {
	fmt.Printf("%04d ", offset)

	if offset >= ch.Count() {
		fmt.Println("Invalid offset")
		return offset + 1
	}

	instruction := ch.Code()[offset]
	switch instruction {
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
