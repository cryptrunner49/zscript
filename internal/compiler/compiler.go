package compiler

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/common"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/lexer"
	"github.com/cryptrunner49/gorex/internal/token"
	"github.com/cryptrunner49/gorex/internal/value"
)

type Parser struct {
	current   token.Token
	previous  token.Token
	hadError  bool
	panicMode bool
}

var parser Parser
var compilingChunk *chunk.Chunk

type Precedence int

const (
	PREC_NONE Precedence = iota
	PREC_ASSIGNMENT
	PREC_OR
	PREC_AND
	PREC_EQUALITY
	PREC_COMPARISON
	PREC_TERM
	PREC_FACTOR
	PREC_UNARY
	PREC_CALL
	PREC_PRIMARY
)

type ParseFn func()

type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence Precedence
}

var rules []ParseRule

func init() {
	// Initialize rules slice with enough capacity
	rules = make([]ParseRule, token.TOKEN_EOF+1)
	// Populate rules dynamically
	rules[token.TOKEN_LEFT_PAREN] = ParseRule{grouping, nil, PREC_NONE}
	rules[token.TOKEN_RIGHT_PAREN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_LEFT_BRACE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_RIGHT_BRACE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_COMMA] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_DOT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_MINUS] = ParseRule{unary, binary, PREC_TERM}
	rules[token.TOKEN_PLUS] = ParseRule{nil, binary, PREC_TERM}
	rules[token.TOKEN_SEMICOLON] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_SLASH] = ParseRule{nil, binary, PREC_FACTOR}
	rules[token.TOKEN_STAR] = ParseRule{nil, binary, PREC_FACTOR}
	rules[token.TOKEN_BANG] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_BANG_EQUAL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_EQUAL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_EQUAL_EQUAL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_GREATER] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_GREATER_EQUAL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_LESS] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_LESS_EQUAL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_IDENTIFIER] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_STRING] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_NUMBER] = ParseRule{number, nil, PREC_NONE}
	rules[token.TOKEN_AND] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_CLASS] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_ELSE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_FALSE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_FOR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_FN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_IF] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_NULL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_OR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_PRINT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_RETURN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_SUPER] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_THIS] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_TRUE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_VAR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_WHILE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_ERROR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_EOF] = ParseRule{nil, nil, PREC_NONE}
}

func currentChunk() *chunk.Chunk {
	return compilingChunk
}

func errorAt(t token.Token, message string) {
	if parser.panicMode {
		return
	}
	parser.panicMode = true
	fmt.Fprintf(os.Stderr, "[line %d] Error", t.Line)
	if t.Type == token.TOKEN_EOF {
		fmt.Fprintf(os.Stderr, " at end")
	} else if t.Type == token.TOKEN_ERROR {
		// Nothing
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

func advance() {
	parser.previous = parser.current
	for {
		parser.current = lexer.ScanToken()
		if parser.current.Type != token.TOKEN_ERROR {
			break
		}
		errorAtCurrent(parser.current.Start)
	}
}

func consume(typ token.TokenType, message string) {
	if parser.current.Type == typ {
		advance()
		return
	}
	errorAtCurrent(message)
}

func emitByte(byte uint8) {
	currentChunk().Write(byte, parser.previous.Line)
}

func emitBytes(byte1, byte2 uint8) {
	emitByte(byte1)
	emitByte(byte2)
}

func emitReturn() {
	emitByte(uint8(chunk.OP_RETURN))
}

func endCompiler() {
	emitReturn()
	/*
		if common.DebugPrintCode && !parser.hadError {
			debug.Disassemble(currentChunk(), "code")
		}
	*/
	if common.DebugPrintCode {
		debug.Disassemble(currentChunk(), "code")
	}
}

func expression() {
	parsePrecedence(PREC_ASSIGNMENT)
}

func parsePrecedence(precedence Precedence) {
	advance()
	prefixRule := getRule(parser.previous.Type).Prefix
	if prefixRule == nil {
		error("Expect expression")
		return
	}
	prefixRule()

	for precedence <= getRule(parser.current.Type).Precedence {
		advance()
		infixRule := getRule(parser.previous.Type).Infix
		infixRule()
	}
}

func grouping() {
	expression()
	consume(token.TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func makeConstant(val value.Value) uint8 {
	constant := currentChunk().AddConstant(val)
	if constant > 255 {
		error("Too many constants in one chunk.")
		return 0
	}
	return uint8(constant)
}

func emitConstant(val value.Value) {
	emitBytes(uint8(chunk.OP_CONSTANT), makeConstant(val))
}

func number() {
	val, _ := strconv.ParseFloat(parser.previous.Start, 64)
	emitConstant(value.Value(val))
}

func unary() {
	operatorType := parser.previous.Type
	parsePrecedence(PREC_UNARY)
	switch operatorType {
	case token.TOKEN_MINUS:
		emitByte(uint8(chunk.OP_NEGATE))
	}
}

func binary() {
	operatorType := parser.previous.Type
	rule := getRule(operatorType)
	parsePrecedence(Precedence(rule.Precedence + 1))
	switch operatorType {
	case token.TOKEN_PLUS:
		emitByte(uint8(chunk.OP_ADD))
	case token.TOKEN_MINUS:
		emitByte(uint8(chunk.OP_SUBTRACT))
	case token.TOKEN_STAR:
		emitByte(uint8(chunk.OP_MULTIPLY))
	case token.TOKEN_SLASH:
		emitByte(uint8(chunk.OP_DIVIDE))
	}
}

func getRule(typ token.TokenType) ParseRule {
	return rules[typ]
}

func Compile(source string, ch *chunk.Chunk) bool {
	lexer.InitLexer(source)
	compilingChunk = ch
	parser.hadError = false
	parser.panicMode = false
	advance()
	expression()
	consume(token.TOKEN_EOF, "Expect end of expression.")
	endCompiler()
	return !parser.hadError
}
