package compiler

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cryptrunner49/gorex/internal/chunk"
	"github.com/cryptrunner49/gorex/internal/common"
	"github.com/cryptrunner49/gorex/internal/debug"
	"github.com/cryptrunner49/gorex/internal/lexer"
	"github.com/cryptrunner49/gorex/internal/object"
	"github.com/cryptrunner49/gorex/internal/token"
	"github.com/cryptrunner49/gorex/internal/value"
)

type Parser struct {
	current   token.Token
	previous  token.Token
	hadError  bool
	panicMode bool
}

type Local struct {
	name  token.Token
	depth int
}

type Compiler struct {
	locals     [256]Local
	localCount int
	scopeDepth int
}

var parser Parser
var current *Compiler
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

type ParseFn func(bool)

type ParseRule struct {
	Prefix     ParseFn
	Infix      ParseFn
	Precedence Precedence
}

var rules []ParseRule

func init() {
	rules = make([]ParseRule, token.TOKEN_EOF+1)
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
	rules[token.TOKEN_BANG] = ParseRule{unary, nil, PREC_NONE}
	rules[token.TOKEN_BANG_EQUAL] = ParseRule{nil, binary, PREC_EQUALITY}
	rules[token.TOKEN_EQUAL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_EQUAL_EQUAL] = ParseRule{nil, binary, PREC_EQUALITY}
	rules[token.TOKEN_GREATER] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_GREATER_EQUAL] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_LESS] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_LESS_EQUAL] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_IDENTIFIER] = ParseRule{variable, nil, PREC_NONE}
	rules[token.TOKEN_STRING] = ParseRule{stringLiteral, nil, PREC_NONE}
	rules[token.TOKEN_NUMBER] = ParseRule{number, nil, PREC_NONE}
	rules[token.TOKEN_AND] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_CLASS] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_ELSE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_FALSE] = ParseRule{literal, nil, PREC_NONE}
	rules[token.TOKEN_FOR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_FN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_IF] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_NULL] = ParseRule{literal, nil, PREC_NONE}
	rules[token.TOKEN_OR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_PRINT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_RETURN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_SUPER] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_THIS] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_TRUE] = ParseRule{literal, nil, PREC_NONE}
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
		// Nothing.
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

func emitByte(b byte) {
	currentChunk().Write(b, parser.previous.Line)
}

func emitBytes(b1, b2 byte) {
	emitByte(b1)
	emitByte(b2)
}

func emitReturn() {
	emitByte(byte(chunk.OP_RETURN))
}

func endCompiler() {
	emitReturn()
	if common.DebugPrintCode && !parser.hadError {
		debug.Disassemble(currentChunk(), "code")
	}
}

func beginScope() {
	current.scopeDepth++
}

func endScope() {
	current.scopeDepth--
	for current.localCount > 0 && current.locals[current.localCount-1].depth > current.scopeDepth {
		emitByte(byte(chunk.OP_POP))
		current.localCount--
	}
}

func expression() {
	parsePrecedence(PREC_ASSIGNMENT)
}

func statement() {
	if match(token.TOKEN_PRINT) {
		printStatement()
	} else if match(token.TOKEN_LEFT_BRACE) {
		beginScope()
		block()
		endScope()
	} else {
		expressionStatement()
	}
}

func block() {
	for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
		declaration()
	}
	consume(token.TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func declaration() {
	if match(token.TOKEN_VAR) {
		varDeclaration()
	} else {
		statement()
	}
	if parser.panicMode {
		synchronize()
	}
}

func varDeclaration() {
	global := parseVariable("Expect variable name.")
	if match(token.TOKEN_EQUAL) {
		expression()
	} else {
		emitByte(byte(chunk.OP_NULL))
	}
	consume(token.TOKEN_SEMICOLON, "Expect ';' after variable declaration.")
	defineVariable(global)
}

func printStatement() {
	expression()
	consume(token.TOKEN_SEMICOLON, "Expect ';' after value.")
	emitByte(byte(chunk.OP_PRINT))
}

func expressionStatement() {
	expression()
	consume(token.TOKEN_SEMICOLON, "Expect ';' after expression.")
	emitByte(byte(chunk.OP_POP))
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

func parsePrecedence(precedence Precedence) {
	advance()
	prefixRule := getRule(parser.previous.Type).Prefix
	if prefixRule == nil {
		error("Expect expression")
		return
	}
	canAssign := precedence <= PREC_ASSIGNMENT
	prefixRule(canAssign)

	for precedence <= getRule(parser.current.Type).Precedence {
		advance()
		infixRule := getRule(parser.previous.Type).Infix
		infixRule(canAssign)
	}

	if canAssign && match(token.TOKEN_EQUAL) {
		error("Invalid assignment target.")
	}
}

func identifierConstant(name token.Token) uint8 {
	return makeConstant(value.Value{Type: value.VAL_OBJ, Obj: object.NewObjString(name.Start)})
}

func identifiersEqual(a, b token.Token) bool {
	return a.Start == b.Start
}

func addLocal(name token.Token) {
	if current.localCount == 256 {
		error("Too many local variables in function.")
		return
	}
	local := &current.locals[current.localCount]
	local.name = name
	local.depth = -1 // Uninitialized
	current.localCount++
}

func declareVariable() {
	if current.scopeDepth == 0 {
		return
	}
	name := parser.previous
	for i := current.localCount - 1; i >= 0; i-- {
		local := current.locals[i]
		if local.depth != -1 && local.depth < current.scopeDepth {
			break
		}
		if identifiersEqual(name, local.name) {
			error("Already variable with this name in this scope.")
		}
	}
	addLocal(name)
}

func parseVariable(errorMessage string) uint8 {
	consume(token.TOKEN_IDENTIFIER, errorMessage)
	declareVariable()
	if current.scopeDepth > 0 {
		return 0
	}
	return identifierConstant(parser.previous)
}

func markInitialized() {
	current.locals[current.localCount-1].depth = current.scopeDepth
}

func defineVariable(global uint8) {
	if current.scopeDepth > 0 {
		markInitialized()
		return
	}
	emitBytes(byte(chunk.OP_DEFINE_GLOBAL), global)
}

func grouping(canAssign bool) {
	expression()
	consume(token.TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func stringLiteral(canAssign bool) {
	text := parser.previous.Start
	if len(text) < 2 {
		error("Invalid string literal.")
		return
	}
	str := text[1 : len(text)-1]
	emitConstant(value.Value{Type: value.VAL_OBJ, Obj: object.NewObjString(str)})
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
	emitBytes(byte(chunk.OP_CONSTANT), makeConstant(val))
}

func number(canAssign bool) {
	val, _ := strconv.ParseFloat(parser.previous.Start, 64)
	emitConstant(value.Value{Type: value.VAL_NUMBER, Number: val})
}

func unary(canAssign bool) {
	operatorType := parser.previous.Type
	parsePrecedence(PREC_UNARY)
	switch operatorType {
	case token.TOKEN_MINUS:
		emitByte(byte(chunk.OP_NEGATE))
	case token.TOKEN_BANG:
		emitByte(byte(chunk.OP_NOT))
	}
}

func binary(canAssign bool) {
	operatorType := parser.previous.Type
	rule := getRule(operatorType)
	parsePrecedence(Precedence(rule.Precedence + 1))
	switch operatorType {
	case token.TOKEN_PLUS:
		emitByte(byte(chunk.OP_ADD))
	case token.TOKEN_MINUS:
		emitByte(byte(chunk.OP_SUBTRACT))
	case token.TOKEN_STAR:
		emitByte(byte(chunk.OP_MULTIPLY))
	case token.TOKEN_SLASH:
		emitByte(byte(chunk.OP_DIVIDE))
	case token.TOKEN_BANG_EQUAL:
		emitBytes(byte(chunk.OP_EQUAL), byte(chunk.OP_NOT))
	case token.TOKEN_EQUAL_EQUAL:
		emitByte(byte(chunk.OP_EQUAL))
	case token.TOKEN_GREATER:
		emitByte(byte(chunk.OP_GREATER))
	case token.TOKEN_GREATER_EQUAL:
		emitBytes(byte(chunk.OP_LESS), byte(chunk.OP_NOT))
	case token.TOKEN_LESS:
		emitByte(byte(chunk.OP_LESS))
	case token.TOKEN_LESS_EQUAL:
		emitBytes(byte(chunk.OP_GREATER), byte(chunk.OP_NOT))
	}
}

func literal(canAssign bool) {
	switch parser.previous.Type {
	case token.TOKEN_FALSE:
		emitByte(byte(chunk.OP_FALSE))
	case token.TOKEN_NULL:
		emitByte(byte(chunk.OP_NULL))
	case token.TOKEN_TRUE:
		emitByte(byte(chunk.OP_TRUE))
	}
}

func resolveLocal(name token.Token) int {
	for i := current.localCount - 1; i >= 0; i-- {
		local := current.locals[i]
		if identifiersEqual(name, local.name) {
			if local.depth == -1 {
				error("Can't read local variable in its own initializer.")
			}
			return i
		}
	}
	return -1
}

func namedVariable(name token.Token, canAssign bool) {
	var getOp, setOp uint8
	arg := resolveLocal(name)
	if arg != -1 {
		getOp = byte(chunk.OP_GET_LOCAL)
		setOp = byte(chunk.OP_SET_LOCAL)
	} else {
		arg = int(identifierConstant(name))
		getOp = byte(chunk.OP_GET_GLOBAL)
		setOp = byte(chunk.OP_SET_GLOBAL)
	}
	if canAssign && match(token.TOKEN_EQUAL) {
		expression()
		emitBytes(setOp, uint8(arg))
	} else {
		emitBytes(getOp, uint8(arg))
	}
}

func variable(canAssign bool) {
	namedVariable(parser.previous, canAssign)
}

func getRule(typ token.TokenType) ParseRule {
	return rules[typ]
}

func initCompiler(compiler *Compiler) {
	compiler.localCount = 0
	compiler.scopeDepth = 0
	current = compiler
}

func Compile(source string, ch *chunk.Chunk) bool {
	lexer.InitLexer(source)
	var compiler Compiler
	initCompiler(&compiler)
	compilingChunk = ch
	parser.hadError = false
	parser.panicMode = false
	advance()
	for !match(token.TOKEN_EOF) {
		declaration()
	}
	endCompiler()
	return !parser.hadError
}
