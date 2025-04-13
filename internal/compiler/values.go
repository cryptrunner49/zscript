package compiler

import (
	"github.com/cryptrunner49/spy/internal/runtime"
	"github.com/cryptrunner49/spy/internal/token"
)

// function compiles a function declaration, including parameter parsing and function body.
func function(funcType FunctionType) {
	var compiler Compiler
	initCompiler(&compiler, funcType, current.scriptDir) // Regular function: no module context
	beginScope()
	consume(token.TOKEN_LEFT_PAREN, "Expected '(' after function name to start parameter list.")
	if !check(token.TOKEN_RIGHT_PAREN) {
		for {
			current.function.Arity++
			if current.function.Arity > 255 {
				errorAtCurrent("Function cannot have more than 255 parameters.")
			}
			paramConstant := parseVariable("Expected a parameter name (e.g., 'x' in 'fn foo(x)').")
			defineVariable(paramConstant)
			if !match(token.TOKEN_COMMA) {
				break
			}
		}
	}
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after parameters.")
	consume(token.TOKEN_COLON, "Expected ':' after function parameters.")
	block()
	function := endCompiler()
	emitBytes(byte(runtime.OP_CLOSURE), makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: function}))
	for i := 0; i < function.UpvalueCount; i++ {
		isLocal := compiler.upvalues[i].isLocal
		index := compiler.upvalues[i].index
		var byteToEmit byte
		if isLocal {
			byteToEmit = 1
		} else {
			byteToEmit = 0
		}
		emitByte(byteToEmit)
		emitByte(index)
	}
}

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
// subscript handles array/map access and slices. Key changes:
// 1. Keep OP_ARRAY_SLICE for slice operations
// 2. Use generic OP_GET_INDEX/OP_SET_INDEX for element access
func subscript(canAssign bool) {
	// Handle slice start (optional)
	hasStart := !check(token.TOKEN_COLON) && !check(token.TOKEN_RIGHT_BRACKET)
	if hasStart {
		expression()
	} else {
		emitConstant(runtime.Value{Type: runtime.VAL_NULL}) // Default start
	}

	if match(token.TOKEN_COLON) {
		// Handle slice end (optional)
		hasEnd := !check(token.TOKEN_RIGHT_BRACKET)
		if hasEnd {
			expression()
		} else {
			emitConstant(runtime.Value{Type: runtime.VAL_NULL}) // Default end
		}
		consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after slice")
		emitByte(byte(runtime.OP_ARRAY_SLICE)) // Array-specific slice
	} else {
		// Regular element access - use generic index ops
		consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after index")
		if canAssign && match(token.TOKEN_EQUAL) {
			expression()
			emitByte(byte(runtime.OP_SET_VALUE)) // Works for arrays AND maps
		} else {
			emitByte(byte(runtime.OP_GET_VALUE)) // Works for arrays AND maps
		}
	}
}

func mapLiteral(canAssign bool) {
	pairs := 0
	for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
		// Parse key
		if match(token.TOKEN_STRING) {
			// Key is a string literal
			key := parser.previous.Start[1 : len(parser.previous.Start)-1]
			emitConstant(runtime.ObjVal(runtime.NewObjString(key)))
		} else if match(token.TOKEN_IDENTIFIER) {
			// Key is an identifier (treated as string)
			key := parser.previous.Start
			emitConstant(runtime.ObjVal(runtime.NewObjString(key)))
		} else {
			reportError("Map key must be a string or identifier")
			return
		}
		consume(token.TOKEN_COLON, "Expected ':' after map key")
		// Parse value
		expression()
		pairs++
		if !match(token.TOKEN_COMMA) {
			break
		}
	}
	consume(token.TOKEN_RIGHT_BRACE, "Expected '}' after map literal")
	emitBytes(byte(runtime.OP_MAP), byte(pairs))
}

// instance emits the OP_INSTANCE opcode with the number of arguments.
func instance(canAssign bool) {
	force := false
	if parser.previous.Type == token.TOKEN_BANG { // '!'' was just consumed
		force = true
	}

	// Parse arguments if '{' follows '!'
	if force && match(token.TOKEN_LEFT_BRACE) {
		// Continue with argument list
	} else if !force {
		// Already consumed '{' above
	} else {
		reportError("Expected '{' after '!' in instance initializer (e.g., 'Struct!{x = 1}').")
		return
	}

	argCount := instanceArgumentList()

	// Emit the force flag as a constant (true if '!' was used, false otherwise)
	emitConstant(runtime.Value{Type: runtime.VAL_BOOL, Bool: force})
	emitBytes(byte(runtime.OP_INSTANCE), argCount)
}

// instanceArgumentList parses key-value pairs for instance initialization (e.g., {x = 1, y = 2}).
// Returns the number of key-value pairs (argCount).
func instanceArgumentList() uint8 {
	var argCount uint8 = 0
	if !check(token.TOKEN_RIGHT_BRACE) {
		for {
			// Expect an identifier (field name)
			consume(token.TOKEN_IDENTIFIER, "Expected field name in instance initializer (e.g., 'x = value').")
			fieldName := parser.previous
			fieldNameConstant := identifierConstant(fieldName)
			emitBytes(byte(runtime.OP_CONSTANT), fieldNameConstant) // Emit field name as a string constant

			// Expect '=' followed by the value
			consume(token.TOKEN_EQUAL, "Expected '=' after field name in instance initializer.")
			expression() // Emit the value expression

			if argCount == 255 {
				reportError("Instance creation cannot have more than 255 field initializers.")
			}
			argCount++ // Increment for each key-value pair

			if !match(token.TOKEN_COMMA) {
				break
			}
		}
	}
	consume(token.TOKEN_RIGHT_BRACE, "Expected '}' to close instance initializer list (e.g., 'Point{x = 1, y = 2}').")
	return argCount
}
