package compiler

import (
	"fmt"
	"os"

	"github.com/cryptrunner49/goseedvm/internal/lexer"
	"github.com/cryptrunner49/goseedvm/internal/runtime"
	"github.com/cryptrunner49/goseedvm/internal/token"
)

func errorAt(t token.Token, message string) {
	if parser.panicMode {
		return
	}
	parser.panicMode = true
	fmt.Fprintf(os.Stderr, "[line %d] Error", t.Line)
	if t.Type == token.TOKEN_EOF {
		fmt.Fprintf(os.Stderr, " at end of file")
	} else if t.Type == token.TOKEN_ERROR {
		// Lexer error already reported
	} else {
		fmt.Fprintf(os.Stderr, " at '%s'", t.Start)
	}
	fmt.Fprintf(os.Stderr, ": %s\n", message)
	parser.hadError = true
}

func error(message string) {
	errorAt(parser.previous, message)
}

func errorAtCurrent(message string) {
	errorAt(parser.current, message)
}

func currentChunk() *runtime.Chunk {
	return &current.function.Chunk
}

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

func consume(typ token.TokenType, message string) {
	if parser.current.Type == typ {
		advance()
		return
	}
	errorAtCurrent(message)
}

func check(typ token.TokenType) bool {
	return parser.current.Type == typ
}

func match(typ token.TokenType) bool {
	if !check(typ) {
		return false
	}
	advance()
	return true
}

func argumentList() uint8 {
	var argCount uint8 = 0
	if !check(token.TOKEN_RIGHT_PAREN) {
		for {
			expression()
			if argCount == 255 {
				error("Function call cannot have more than 255 arguments.")
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

func synchronize() {
	parser.panicMode = false
	for parser.current.Type != token.TOKEN_EOF {
		if parser.previous.Type == token.TOKEN_SEMICOLON {
			return
		}
		switch parser.current.Type {
		case token.TOKEN_CLASS, token.TOKEN_FN, token.TOKEN_VAR, token.TOKEN_FOR,
			token.TOKEN_IF, token.TOKEN_WHILE, token.TOKEN_PRINT, token.TOKEN_RETURN:
			return
		}
		advance()
	}
}

func identifierConstant(name token.Token) uint8 {
	return makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(name.Start)})
}

func identifiersEqual(a, b token.Token) bool {
	return a.Start == b.Start
}
