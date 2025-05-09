package compiler

import (
	"fmt"
	"os"

	"github.com/cryptrunner49/zscript/internal/lexer"
	"github.com/cryptrunner49/zscript/internal/runtime"
	"github.com/cryptrunner49/zscript/internal/token"
)

// errorAt reports a compilation error at the specified token, printing the error message with the
// token's line number and context (e.g., token text or "end of file"). If in panic mode, it suppresses
// further error reporting to avoid cascading errors.
func errorAt(t token.Token, message string) {
	if parser.panicMode {
		return
	}
	parser.panicMode = true
	fmt.Fprintf(os.Stderr, "[line %d] Error", t.Line)
	if t.Type == token.TOKEN_EOF {
		fmt.Fprintf(os.Stderr, " at end of file")
	} else if t.Type == token.TOKEN_ERROR {
		// Lexer error already reported, no further action required.
	} else {
		fmt.Fprintf(os.Stderr, " at '%s'", t.Start)
	}
	fmt.Fprintf(os.Stderr, ": %s\n", message)
	parser.hadError = true
}

// error reports an error using the previous token.
func reportError(message string) {
	errorAt(parser.previous, message)
}

// errorAtCurrent reports an error at the current token.
func errorAtCurrent(message string) {
	errorAt(parser.current, message)
}

// currentChunk retrieves the current chunk of bytecode being compiled.
func currentChunk() *runtime.Chunk {
	return &current.function.Chunk
}

// advance moves to the next token, skipping over any lexer errors and reporting them.
func advance() {
	parser.previous = parser.current
	for {
		parser.current = lexer.ScanToken()
		if parser.current.Type != token.TOKEN_ERROR {
			break
		}
		errorAtCurrent(fmt.Sprintf("Invalid token '%s' encountered.", parser.current.Start))
	}
}

// consume expects the current token to be of a specific type and advances, or reports an error.
func consume(typ token.TokenType, message string) {
	if parser.current.Type == typ {
		advance()
		return
	}
	errorAtCurrent(message)
}

// check returns true if the current token is of the expected type.
func check(typ token.TokenType) bool {
	return parser.current.Type == typ
}

// match checks for a token type match and advances if a match is found.
func match(typ token.TokenType) bool {
	if !check(typ) {
		return false
	}
	advance()
	return true
}

// argumentList compiles the list of arguments in a function call and returns the count.
func argumentList() uint8 {
	var argCount uint8 = 0
	if !check(token.TOKEN_RIGHT_PAREN) {
		for {
			expression()
			if argCount == 255 {
				reportError("Function call cannot have more than 255 arguments.")
			}
			argCount++
			if !match(token.TOKEN_COMMA) {
				break
			}
		}
	}
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' to close argument list (e.g., 'func(a, b)').")
	return argCount
}

// synchronize discards tokens until it reaches a statement boundary, helping recover from errors.
func synchronize() {
	parser.panicMode = false
	for parser.current.Type != token.TOKEN_EOF {
		if parser.previous.Type == token.TOKEN_SEMICOLON {
			return
		}
		switch parser.current.Type {
		case token.TOKEN_CLASS, token.TOKEN_FUNC, token.TOKEN_VAR, token.TOKEN_FOR,
			token.TOKEN_IF, token.TOKEN_WHILE, token.TOKEN_RETURN:
			return
		}
		advance()
	}
}

// identifierConstant creates a constant for an identifier (variable name) and returns its index.
func identifierConstant(name token.Token) uint8 {
	return makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(name.Start)})
}

// identifiersEqual checks if two identifier tokens are equal based on their string content.
func identifiersEqual(a, b token.Token) bool {
	return a.Start == b.Start
}

// consumeOptionalSemicolon tries to match a semicolon.
func consumeOptionalSemicolon() {
	if match(token.TOKEN_SEMICOLON) {
		return
	} else {
		return
	}
}
