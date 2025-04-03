package compiler

import (
	"github.com/cryptrunner49/goseedvm/internal/runtime"
	"github.com/cryptrunner49/goseedvm/internal/token"
)

// arrayLiteral parses an array literal and emits the corresponding bytecode.
// It collects the elements, enforces a maximum element count of 255, and then
// emits an OP_ARRAY opcode with the element count.
func arrayLiteral(canAssign bool) {
	elementCount := 0
	if !check(token.TOKEN_RIGHT_BRACKET) {
		for {
			expression()
			elementCount++
			if elementCount == 255 {
				reportError("Array literal cannot have more than 255 elements.")
			}
			if !match(token.TOKEN_COMMA) {
				break
			}
		}
	}
	consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after array elements.")

	emitBytes(byte(runtime.OP_ARRAY), byte(elementCount))
}

// subscript parses array subscript expressions, handling both element access and slice syntax.
// For slices, it expects an optional start expression, a colon, and an optional end expression.
// It emits either an OP_ARRAY_SLICE or an OP_ARRAY_GET/OP_ARRAY_SET opcode depending on the context.
func subscript(canAssign bool) {
	// Determine if there is a start index specified.
	hasStart := !check(token.TOKEN_COLON) && !check(token.TOKEN_RIGHT_BRACKET)
	if hasStart {
		expression()
	} else {
		// If no start index is provided, push null as default.
		emitConstant(runtime.Value{Type: runtime.VAL_NULL})
	}

	if match(token.TOKEN_COLON) {
		// Slice syntax detected; check for an end index.
		hasEnd := !check(token.TOKEN_RIGHT_BRACKET)
		if hasEnd {
			expression()
		} else {
			// If no end index is provided, push null as default.
			emitConstant(runtime.Value{Type: runtime.VAL_NULL})
		}
		consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after slice indices.")
		emitByte(byte(runtime.OP_ARRAY_SLICE))
	} else {
		// Regular array index access.
		consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after array index.")
		if canAssign && match(token.TOKEN_EQUAL) {
			// Assignment to an array element.
			expression()
			emitByte(byte(runtime.OP_ARRAY_SET))
		} else {
			// Retrieve an array element.
			emitByte(byte(runtime.OP_ARRAY_GET))
		}
	}
}
