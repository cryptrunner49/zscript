package debug

import (
	"fmt"

	"github.com/cryptrunner49/spy/internal/runtime"
)

func Disassemble(ch *runtime.Chunk, name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < ch.Count(); {
		offset = DisassembleInstruction(ch, offset)
	}
}

func DisassembleInstruction(ch *runtime.Chunk, offset int) int {
	fmt.Printf("%04d ", offset)
	if offset > 0 && ch.Lines()[offset] == ch.Lines()[offset-1] {
		fmt.Print("   | ")
	} else {
		fmt.Printf("%4d ", ch.Lines()[offset])
	}

	instruction := ch.Code()[offset]
	switch instruction {
	case uint8(runtime.OP_CONSTANT):
		return constantInstruction("OP_CONSTANT", ch, offset)
	case uint8(runtime.OP_NULL):
		return simpleInstruction("OP_NULL", offset)
	case uint8(runtime.OP_TRUE):
		return simpleInstruction("OP_TRUE", offset)
	case uint8(runtime.OP_FALSE):
		return simpleInstruction("OP_FALSE", offset)
	case uint8(runtime.OP_POP):
		return simpleInstruction("OP_POP", offset)
	case uint8(runtime.OP_SET_LOCAL):
		return byteInstruction("OP_SET_LOCAL", ch, offset)
	case uint8(runtime.OP_GET_LOCAL):
		return byteInstruction("OP_GET_LOCAL", ch, offset)
	case uint8(runtime.OP_DEFINE_GLOBAL):
		return constantInstruction("OP_DEFINE_GLOBAL", ch, offset)
	case uint8(runtime.OP_SET_GLOBAL):
		return constantInstruction("OP_SET_GLOBAL", ch, offset)
	case uint8(runtime.OP_GET_GLOBAL):
		return constantInstruction("OP_GET_GLOBAL", ch, offset)
	case uint8(runtime.OP_GET_UPVALUE):
		return byteInstruction("OP_GET_UPVALUE", ch, offset)
	case uint8(runtime.OP_SET_UPVALUE):
		return byteInstruction("OP_SET_UPVALUE", ch, offset)
	case uint8(runtime.OP_GET_PROPERTY):
		return constantInstruction("OP_GET_PROPERTY", ch, offset)
	case uint8(runtime.OP_SET_PROPERTY):
		return constantInstruction("OP_SET_PROPERTY", ch, offset)
	case uint8(runtime.OP_EQUAL):
		return simpleInstruction("OP_EQUAL", offset)
	case uint8(runtime.OP_GREATER):
		return simpleInstruction("OP_GREATER", offset)
	case uint8(runtime.OP_LESS):
		return simpleInstruction("OP_LESS", offset)
	case uint8(runtime.OP_ADD):
		return simpleInstruction("OP_ADD", offset)
	case uint8(runtime.OP_SUBTRACT):
		return simpleInstruction("OP_SUBTRACT", offset)
	case uint8(runtime.OP_MULTIPLY):
		return simpleInstruction("OP_MULTIPLY", offset)
	case uint8(runtime.OP_DIVIDE):
		return simpleInstruction("OP_DIVIDE", offset)
	case uint8(runtime.OP_NOT):
		return simpleInstruction("OP_NOT", offset)
	case uint8(runtime.OP_NEGATE):
		return simpleInstruction("OP_NEGATE", offset)
	case uint8(runtime.OP_CALL):
		return byteInstruction("OP_CALL", ch, offset)
	case uint8(runtime.OP_CLOSURE):
		offset++
		constant := ch.Code()[offset]
		offset++
		fmt.Printf("%-16s %4d ", "OP_CLOSURE", constant)
		runtime.PrintValue(ch.Constants().Values()[constant])
		fmt.Println()
		function := ch.Constants().Values()[constant].Obj.(*runtime.ObjFunction)
		for j := 0; j < function.UpvalueCount; j++ {
			isLocal := ch.Code()[offset]
			offset++
			index := ch.Code()[offset]
			offset++
			var upvalueType string
			if isLocal != 0 {
				upvalueType = "local"
			} else {
				upvalueType = "upvalue"
			}
			fmt.Printf("%04d      | %s %d\n", offset-2, upvalueType, index)
		}
		return offset
	case uint8(runtime.OP_CLOSE_UPVALUE):
		return simpleInstruction("OP_CLOSE_UPVALUE", offset)
	case uint8(runtime.OP_RETURN):
		return simpleInstruction("OP_RETURN", offset)
	case uint8(runtime.OP_JUMP):
		return jumpInstruction("OP_JUMP", 1, ch, offset)
	case uint8(runtime.OP_JUMP_IF_FALSE):
		return jumpInstruction("OP_JUMP_IF_FALSE", 1, ch, offset)
	case uint8(runtime.OP_JUMP_IF_TRUE):
		return jumpInstruction("OP_JUMP_IF_TRUE", 1, ch, offset)
	case uint8(runtime.OP_LOOP):
		return jumpInstruction("OP_LOOP", -1, ch, offset)
	case uint8(runtime.OP_BREAK):
		return jumpInstruction("OP_BREAK", 1, ch, offset)
	case uint8(runtime.OP_CONTINUE):
		return jumpInstruction("OP_CONTINUE", 1, ch, offset)
	case uint8(runtime.OP_STRUCT):
		return structInstruction(ch, offset)
	case uint8(runtime.OP_INSTANCE):
		return byteInstruction("OP_INSTANCE", ch, offset)
	case uint8(runtime.OP_GET_VALUE):
		return simpleInstruction("OP_GET_VALUE", offset)
	case uint8(runtime.OP_SET_VALUE):
		return simpleInstruction("OP_SET_VALUE", offset)
	case uint8(runtime.OP_ARRAY):
		return byteInstruction("OP_ARRAY", ch, offset)
	case uint8(runtime.OP_ARRAY_LEN):
		return simpleInstruction("OP_ARRAY_LEN", offset)
	case uint8(runtime.OP_ARRAY_SLICE):
		return simpleInstruction("OP_ARRAY_SLICE", offset)
	case uint8(runtime.OP_MAP):
		pairCount := int(ch.Code()[offset+1])
		fmt.Printf("%-16s %d pairs\n", "OP_MAP", pairCount)
		return offset + 2
	case uint8(runtime.OP_MODULE):
		return constantInstruction("OP_MODULE", ch, offset)
	case uint8(runtime.OP_IMPORT):
		return constantInstruction("OP_IMPORT", ch, offset)
	case uint8(runtime.OP_USE):
		return constantInstruction("OP_USE", ch, offset)
	case uint8(runtime.OP_DEFINE_EXTERN):
		offset++
		returnTypeIdx := int(ch.Code()[offset])
		fmt.Printf("%-16s return type: %d '", "OP_DEFINE_EXTERN", returnTypeIdx)
		runtime.PrintValue(ch.Constants().Values()[returnTypeIdx])
		fmt.Println("'")
		offset++
		paramCount := int(ch.Code()[offset])
		fmt.Printf("          param count: %d\n", paramCount)
		offset++
		for i := 0; i < paramCount; i++ {
			paramTypeIdx := int(ch.Code()[offset])
			fmt.Printf("          param %d: %d '", i, paramTypeIdx)
			runtime.PrintValue(ch.Constants().Values()[paramTypeIdx])
			fmt.Println("'")
			offset++
		}
		funcNameIdx := int(ch.Code()[offset])
		fmt.Printf("          function name: %d '", funcNameIdx)
		runtime.PrintValue(ch.Constants().Values()[funcNameIdx])
		fmt.Println("'")
		return offset + 1
	case uint8(runtime.OP_MATCH):
		return matchInstruction(ch, offset)
	case uint8(runtime.OP_DUP):
		return simpleInstruction("OP_DUP", offset)
	case uint8(runtime.OP_EXPONENTIAL):
		return simpleInstruction("OP_EXPONENTIAL", offset)
	case uint8(runtime.OP_FLOOR):
		return simpleInstruction("OP_FLOOR", offset)
	case uint8(runtime.OP_PERCENT):
		return simpleInstruction("OP_PERCENT", offset)
	default:
		fmt.Printf("Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

func simpleInstruction(name string, offset int) int {
	fmt.Println(name)
	return offset + 1
}

func constantInstruction(name string, ch *runtime.Chunk, offset int) int {
	constant := ch.Code()[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	runtime.PrintValue(ch.Constants().Values()[constant])
	fmt.Println("'")
	return offset + 2
}

func byteInstruction(name string, ch *runtime.Chunk, offset int) int {
	slot := ch.Code()[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot)
	return offset + 2
}

func jumpInstruction(name string, sign int, ch *runtime.Chunk, offset int) int {
	jump := int(ch.Code()[offset+1])<<8 | int(ch.Code()[offset+2])
	fmt.Printf("%-16s %4d -> %d\n", name, offset, offset+3+sign*jump)
	return offset + 3
}

func structInstruction(ch *runtime.Chunk, offset int) int {
	// Read the struct name constant.
	constant := ch.Code()[offset+1]
	fmt.Printf("%-16s %4d '", "OP_STRUCT", constant)
	runtime.PrintValue(ch.Constants().Values()[constant])
	fmt.Println("'")
	// Read the field count.
	fieldCount := int(ch.Code()[offset+2])
	fmt.Printf("          field count: %d\n", fieldCount)
	// Advance past opcode, struct name, and field count.
	offset += 3
	// For each field, print the field name and its default value.
	for i := 0; i < fieldCount; i++ {
		// Field name constant.
		nameConstant := ch.Code()[offset]
		fmt.Printf("%04d      | field name constant %d: '", offset, nameConstant)
		runtime.PrintValue(ch.Constants().Values()[nameConstant])
		fmt.Println("'")
		offset++
		// Field default value constant.
		defConstant := ch.Code()[offset]
		fmt.Printf("%04d      | field default constant %d: '", offset, defConstant)
		runtime.PrintValue(ch.Constants().Values()[defConstant])
		fmt.Println("'")
		offset++
	}
	return offset
}

// New function to disassemble OP_MATCH
func matchInstruction(ch *runtime.Chunk, offset int) int {
	fmt.Printf("%-16s", "OP_MATCH")
	offset++ // Skip opcode
	numCases := ch.Code()[offset]
	fmt.Printf("%d cases, default -> ", numCases)
	offset++
	defaultOffset := (int(ch.Code()[offset]) << 8) | int(ch.Code()[offset+1])
	fmt.Printf("%04d", offset+defaultOffset)
	offset += 2
	for i := 0; i < int(numCases); i++ {
		if offset+2 >= len(ch.Code()) {
			fmt.Println("\n      | <truncated>")
			return offset
		}
		constant := ch.Code()[offset]
		fmt.Printf("\n%04d      | case constant %d: '", offset, constant)
		if constant < uint8(len(ch.Constants().Values())) {
			runtime.PrintValue(ch.Constants().Values()[constant])
		} else {
			fmt.Print("<invalid>")
		}
		fmt.Print("'")
		offset++
		caseOffset := (int(ch.Code()[offset]) << 8) | int(ch.Code()[offset+1])
		fmt.Printf(" offset -> %04d", offset+caseOffset-1)
		offset += 2
	}
	fmt.Println()
	return offset
}
