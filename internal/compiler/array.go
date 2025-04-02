package compiler

import (
	"github.com/cryptrunner49/goseedvm/internal/runtime"
	"github.com/cryptrunner49/goseedvm/internal/token"
)

// Add array literal parsing
func arrayLiteral(canAssign bool) {
	elementCount := 0
	if !check(token.TOKEN_RIGHT_BRACKET) {
		for {
			expression()
			elementCount++
			if elementCount == 255 {
				error("Array literal cannot have more than 255 elements.")
			}
			if !match(token.TOKEN_COMMA) {
				break
			}
		}
	}
	consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after array elements.")

	emitBytes(byte(runtime.OP_ARRAY), byte(elementCount))
}

// Add array subscript access
func subscript(canAssign bool) {
	expression()
	consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after array index.")

	if canAssign && match(token.TOKEN_EQUAL) {
		expression()
		emitByte(byte(runtime.OP_ARRAY_SET))
	} else {
		emitByte(byte(runtime.OP_ARRAY_GET))
	}
}
