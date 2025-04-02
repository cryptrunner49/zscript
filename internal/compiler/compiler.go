package compiler

import (
	"fmt"
	"strconv"

	"github.com/cryptrunner49/goseedvm/internal/common"
	"github.com/cryptrunner49/goseedvm/internal/debug"
	"github.com/cryptrunner49/goseedvm/internal/lexer"
	"github.com/cryptrunner49/goseedvm/internal/runtime"
	"github.com/cryptrunner49/goseedvm/internal/token"
)

// FunctionType is an enum used to differentiate between function declarations and top-level scripts.
type FunctionType int

const (
	TYPE_FUNCTION FunctionType = iota // Regular function definition.
	TYPE_SCRIPT                       // Top-level script execution.
)

// Parser holds the current and previous tokens and error flags for parsing.
type Parser struct {
	current   token.Token // The current token being processed.
	previous  token.Token // The previous token processed.
	hadError  bool        // Flag indicating if a parsing error has occurred.
	panicMode bool        // Flag to suppress cascading errors after an error is encountered.
}

// Local represents a local variable with its name, scope depth, and if it was captured by a closure.
type Local struct {
	name       token.Token // Token representing the variable's name.
	depth      int         // Scope depth where the variable was declared.
	isCaptured bool        // Indicates if the variable is captured by an enclosing function.
}

// Upvalue holds information about a variable captured by a closure.
type Upvalue struct {
	index   uint8 // Index of the variable in the parent's local variables.
	isLocal bool  // Indicates if the captured variable was a local variable.
}

// LoopType defines different kinds of loops.
type LoopType int

const (
	LOOP_WHILE LoopType = iota // While loop.
	LOOP_FOR                   // For loop.
)

// Loop is used to manage loop state during compilation, including jump patching.
type Loop struct {
	loopType        LoopType // Type of loop (while or for).
	start           int      // Bytecode index where the loop begins.
	exitPatches     []int    // List of jump offsets to patch for loop exit.
	continuePatches []int    // List of jump offsets to patch for continue statements.
	exitAddress     int      // Address to jump to when exiting the loop.
	hasIncrement    bool     // Flag indicating if the loop has an increment expression.
	incrementStart  int      // Bytecode index where the increment expression starts.
}

// Compiler holds the current state of the compilation process.
type Compiler struct {
	enclosing    *Compiler            // Reference to the parent compiler for nested functions.
	function     *runtime.ObjFunction // The function object currently being compiled.
	functionType FunctionType         // Type of function (regular or script).
	locals       [256]Local           // Fixed array of local variables.
	localCount   int                  // Current count of local variables.
	upvalues     [256]Upvalue         // Fixed array of upvalues for closures.
	scopeDepth   int                  // Current depth of local scope nesting.
	loops        []Loop               // Stack of active loops for break/continue handling.
}

var parser Parser     // Global parser state.
var current *Compiler // Pointer to the current compiler instance.

// Precedence defines operator precedence levels.
type Precedence int

const (
	PREC_NONE       Precedence = iota // No precedence.
	PREC_ASSIGNMENT                   // Assignment operators.
	PREC_OR                           // Logical OR.
	PREC_AND                          // Logical AND.
	PREC_EQUALITY                     // Equality operators.
	PREC_COMPARISON                   // Comparison operators.
	PREC_TERM                         // Term operators (addition, subtraction).
	PREC_FACTOR                       // Factor operators (multiplication, division).
	PREC_UNARY                        // Unary operators.
	PREC_CALL                         // Call and subscript operators.
	PREC_PRIMARY                      // Primary expressions.
)

// ParseFn represents a pointer to a parsing function.
type ParseFn func(bool)

// ParseRule defines the rules for parsing a token type: its prefix, infix parsing functions, and precedence.
type ParseRule struct {
	Prefix     ParseFn    // Function to call when token is used in a prefix context.
	Infix      ParseFn    // Function to call when token is used in an infix context.
	Precedence Precedence // Operator precedence level for the token.
}

var rules []ParseRule // Table mapping token types to their parsing rules.

// Initialize the parsing rules for each token type.
func init() {
	rules = make([]ParseRule, token.TOKEN_EOF+1)
	rules[token.TOKEN_LEFT_PAREN] = ParseRule{grouping, call, PREC_CALL}
	rules[token.TOKEN_RIGHT_PAREN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_LEFT_BRACE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_RIGHT_BRACE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_LEFT_BRACKET] = ParseRule{arrayLiteral, subscript, PREC_CALL}
	rules[token.TOKEN_RIGHT_BRACKET] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_COMMA] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_DOT] = ParseRule{nil, dot, PREC_CALL}
	rules[token.TOKEN_MINUS] = ParseRule{unary, binary, PREC_TERM}
	rules[token.TOKEN_PLUS] = ParseRule{nil, binary, PREC_TERM}
	rules[token.TOKEN_SEMICOLON] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_SLASH] = ParseRule{nil, binary, PREC_FACTOR}
	rules[token.TOKEN_PERCENT] = ParseRule{nil, binary, PREC_FACTOR}
	rules[token.TOKEN_STAR] = ParseRule{nil, binary, PREC_FACTOR}
	rules[token.TOKEN_PIPE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_QUESTION] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_AT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_HASH] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_DOLLAR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_COLON] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_BANG] = ParseRule{unary, nil, PREC_NONE}
	rules[token.TOKEN_BANG_EQUAL] = ParseRule{nil, binary, PREC_EQUALITY}
	rules[token.TOKEN_EQUAL] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_EQUAL_EQUAL] = ParseRule{nil, binary, PREC_EQUALITY}
	rules[token.TOKEN_GREATER] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_GREATER_EQUAL] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_LESS] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_LESS_EQUAL] = ParseRule{nil, binary, PREC_COMPARISON}
	rules[token.TOKEN_IDENTIFIER] = ParseRule{variable, nil, PREC_NONE}
	rules[token.TOKEN_CHAR] = ParseRule{charLiteral, nil, PREC_NONE}
	rules[token.TOKEN_STRING] = ParseRule{stringLiteral, nil, PREC_NONE}
	rules[token.TOKEN_NUMBER] = ParseRule{number, nil, PREC_NONE}
	rules[token.TOKEN_AND] = ParseRule{nil, and, PREC_AND}
	rules[token.TOKEN_CLASS] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_ELSE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_FALSE] = ParseRule{literal, nil, PREC_NONE}
	rules[token.TOKEN_FOR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_FN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_IF] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_NULL] = ParseRule{literal, nil, PREC_NONE}
	rules[token.TOKEN_OR] = ParseRule{nil, or, PREC_OR}
	rules[token.TOKEN_PRINT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_RETURN] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_SUPER] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_STRUCT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_THIS] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_TRUE] = ParseRule{literal, nil, PREC_NONE}
	rules[token.TOKEN_VAR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_WHILE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_ITER] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_BREAK] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_CONTINUE] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_MATCH] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_WITH] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_THROUGH] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_RANDOM] = ParseRule{random, nil, PREC_NONE}
	rules[token.TOKEN_IMPORT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_EXPORT] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_ERROR] = ParseRule{nil, nil, PREC_NONE}
	rules[token.TOKEN_EOF] = ParseRule{nil, nil, PREC_NONE}
}

// dot handles property access on objects (e.g., object.field).
func dot(canAssign bool) {
	consume(token.TOKEN_IDENTIFIER, "Expected a property name after '.' (e.g., 'object.field').")
	name := identifierConstant(parser.previous)
	if canAssign && match(token.TOKEN_EQUAL) {
		expression()
		emitBytes(byte(runtime.OP_SET_PROPERTY), name)
	} else {
		emitBytes(byte(runtime.OP_GET_PROPERTY), name)
	}
}

// emitByte writes a single byte into the current chunk with the current line number.
func emitByte(b byte) {
	currentChunk().Write(b, parser.previous.Line)
}

// emitBytes writes two consecutive bytes to the current chunk.
func emitBytes(b1, b2 byte) {
	emitByte(b1)
	emitByte(b2)
}

// emitReturn writes the return opcode to the chunk, ending the function.
func emitReturn() {
	emitByte(byte(runtime.OP_NULL))
	emitByte(byte(runtime.OP_RETURN))
}

// endCompiler finishes the current function, emits a return, and optionally disassembles the code for debugging.
func endCompiler() *runtime.ObjFunction {
	emitReturn()
	function := current.function
	if common.DebugPrintCode && !parser.hadError {
		name := "<script>"
		if function.Name != nil {
			name = function.Name.Chars
		}
		debug.Disassemble(currentChunk(), name)
	}
	current = current.enclosing
	return function
}

// block compiles a block statement by repeatedly compiling declarations until a closing brace is found.
func block() {
	for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
		declaration()
	}
	consume(token.TOKEN_RIGHT_BRACE, "Expected '}' to close block (unmatched '{').")
}

// beginScope increases the scope depth, starting a new local variable scope.
func beginScope() {
	current.scopeDepth++
}

// endScope decreases the scope depth and removes local variables declared in that scope.
func endScope() {
	current.scopeDepth--
	for current.localCount > 0 && current.locals[current.localCount-1].depth > current.scopeDepth {
		if current.locals[current.localCount-1].isCaptured {
			emitByte(byte(runtime.OP_CLOSE_UPVALUE))
		} else {
			emitByte(byte(runtime.OP_POP))
		}
		current.localCount--
	}
}

// expression compiles an expression using the assignment precedence level.
func expression() {
	parsePrecedence(PREC_ASSIGNMENT)
}

// statement compiles a statement, dispatching to the appropriate function based on the token.
func statement() {
	if match(token.TOKEN_PRINT) {
		printStatement()
	} else if match(token.TOKEN_IF) {
		ifStatement()
	} else if match(token.TOKEN_WHILE) {
		whileStatement()
	} else if match(token.TOKEN_FOR) {
		forStatement()
	} else if match(token.TOKEN_ITER) {
		iterStatement()
	} else if match(token.TOKEN_BREAK) {
		breakStatement()
	} else if match(token.TOKEN_CONTINUE) {
		continueStatement()
	} else if match(token.TOKEN_RETURN) {
		returnStatement()
	} else if match(token.TOKEN_MATCH) {
		matchStatement()
	} else if match(token.TOKEN_LEFT_BRACE) {
		beginScope()
		block()
		endScope()
	} else {
		expressionStatement()
	}
}

// declaration compiles declarations like variables, functions, and structs.
func declaration() {
	if match(token.TOKEN_STRUCT) {
		structDeclaration()
	} else if match(token.TOKEN_FN) {
		fnDeclaration()
	} else if match(token.TOKEN_VAR) {
		varDeclaration()
	} else if match(token.TOKEN_IMPORT) {
		importDeclaration()
	} else if match(token.TOKEN_EXPORT) {
		exportDeclaration()
	} else {
		statement()
	}
	if parser.panicMode {
		synchronize()
	}
}

// call compiles a function call by parsing the argument list and emitting the call opcode.
func call(canAssign bool) {
	argCount := argumentList()
	emitBytes(byte(runtime.OP_CALL), argCount)
}

// parsePrecedence compiles an expression based on a minimum precedence, handling operators accordingly.
func parsePrecedence(precedence Precedence) {
	advance()
	prefixRule := getRule(parser.previous.Type).Prefix
	if prefixRule == nil {
		error("Expected an expression but found no valid starting token.")
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
		error("Invalid assignment target; only variables or properties can be assigned.")
	}
}

// addLocal adds a new local variable to the current compiler state.
func addLocal(name token.Token) {
	if current.localCount == 256 {
		error("Too many local variables in this scope (max 256).")
		return
	}
	local := &current.locals[current.localCount]
	local.name = name
	local.depth = -1 // Uninitialized.
	local.isCaptured = false
	current.localCount++
}

// declareVariable handles variable declarations and checks for redeclaration in the same scope.
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
			error(fmt.Sprintf("Variable '%s' is already declared in this scope.", name.Start))
		}
	}
	addLocal(name)
}

// parseVariable parses an identifier token for variable declarations.
func parseVariable(errorMessage string) uint8 {
	consume(token.TOKEN_IDENTIFIER, errorMessage)
	declareVariable()
	if current.scopeDepth > 0 {
		return 0
	}
	return identifierConstant(parser.previous)
}

// markInitialized marks the most recently added local variable as initialized.
func markInitialized() {
	if current.scopeDepth == 0 {
		return
	}
	current.locals[current.localCount-1].depth = current.scopeDepth
}

// function compiles a function declaration, including parameter parsing and function body.
func function(funcType FunctionType) {
	var compiler Compiler
	initCompiler(&compiler, funcType)
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
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' to close parameter list (e.g., 'fn foo()').")
	consume(token.TOKEN_LEFT_BRACE, "Expected '{' to start function body.")
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

// defineVariable either marks a variable as initialized in the local scope or emits a global definition.
func defineVariable(global uint8) {
	if current.scopeDepth > 0 {
		markInitialized()
		return
	}
	emitBytes(byte(runtime.OP_DEFINE_GLOBAL), global)
}

// grouping compiles a grouped expression enclosed in parentheses.
func grouping(canAssign bool) {
	expression()
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' to close grouped expression (unmatched '(').")
}

// stringLiteral compiles a string literal by removing the enclosing quotes and emitting a constant.
func stringLiteral(canAssign bool) {
	text := parser.previous.Start
	if len(text) < 2 {
		error("Invalid string literal; must be enclosed in quotes (e.g., \"hello\").")
		return
	}
	str := text[1 : len(text)-1]
	emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(str)})
}

// charLiteral compiles a character literal by removing the enclosing quotes.
func charLiteral(canAssign bool) {
	text := parser.previous.Start
	if len(text) < 2 {
		error("Invalid char literal; must be enclosed in quotes (e.g., 'a').")
		return
	}
	str := text[1 : len(text)-1]
	emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(str)})
}

// makeConstant adds a constant value to the current chunk and returns its index.
func makeConstant(val runtime.Value) uint8 {
	constant := currentChunk().AddConstant(val)
	if constant > 255 {
		error("Too many constants in this chunk (max 256). Consider splitting the code.")
		return 0
	}
	return uint8(constant)
}

// emitConstant writes the constant opcode along with the index of the constant.
func emitConstant(val runtime.Value) {
	emitBytes(byte(runtime.OP_CONSTANT), makeConstant(val))
}

// number compiles a numeric literal by parsing it and emitting the constant.
func number(canAssign bool) {
	val, err := strconv.ParseFloat(parser.previous.Start, 64)
	if err != nil {
		error(fmt.Sprintf("Invalid number literal '%s'; must be a valid number.", parser.previous.Start))
		return
	}
	emitConstant(runtime.Value{Type: runtime.VAL_NUMBER, Number: val})
}

// unary compiles a unary operator expression.
func unary(canAssign bool) {
	operatorType := parser.previous.Type
	parsePrecedence(PREC_UNARY)
	switch operatorType {
	case token.TOKEN_MINUS:
		emitByte(byte(runtime.OP_NEGATE))
	case token.TOKEN_BANG:
		emitByte(byte(runtime.OP_NOT))
	}
}

// binary compiles a binary operator expression.
func binary(canAssign bool) {
	operatorType := parser.previous.Type
	rule := getRule(operatorType)
	parsePrecedence(Precedence(rule.Precedence + 1))
	switch operatorType {
	case token.TOKEN_PLUS:
		emitByte(byte(runtime.OP_ADD))
	case token.TOKEN_MINUS:
		emitByte(byte(runtime.OP_SUBTRACT))
	case token.TOKEN_STAR:
		emitByte(byte(runtime.OP_MULTIPLY))
	case token.TOKEN_SLASH:
		emitByte(byte(runtime.OP_DIVIDE))
	case token.TOKEN_PERCENT:
		emitByte(byte(runtime.OP_MOD))
	case token.TOKEN_BANG_EQUAL:
		emitBytes(byte(runtime.OP_EQUAL), byte(runtime.OP_NOT))
	case token.TOKEN_EQUAL_EQUAL:
		emitByte(byte(runtime.OP_EQUAL))
	case token.TOKEN_GREATER:
		emitByte(byte(runtime.OP_GREATER))
	case token.TOKEN_GREATER_EQUAL:
		emitBytes(byte(runtime.OP_LESS), byte(runtime.OP_NOT))
	case token.TOKEN_LESS:
		emitByte(byte(runtime.OP_LESS))
	case token.TOKEN_LESS_EQUAL:
		emitBytes(byte(runtime.OP_GREATER), byte(runtime.OP_NOT))
	}
}

// literal compiles literal tokens like false, null, or true.
func literal(canAssign bool) {
	switch parser.previous.Type {
	case token.TOKEN_FALSE:
		emitByte(byte(runtime.OP_FALSE))
	case token.TOKEN_NULL:
		emitByte(byte(runtime.OP_NULL))
	case token.TOKEN_TRUE:
		emitByte(byte(runtime.OP_TRUE))
	}
}

// random compiles the random operator, currently a stub.
func random(canAssign bool) {

}

// resolveLocal searches for a local variable by name in the given compiler and returns its index or -1 if not found.
func resolveLocal(comp *Compiler, name token.Token) int {
	for i := comp.localCount - 1; i >= 0; i-- {
		local := comp.locals[i]
		if identifiersEqual(name, local.name) {
			if local.depth == -1 {
				error(fmt.Sprintf("Cannot use variable '%s' in its own initializer.", name.Start))
			}
			return i
		}
	}
	return -1
}

// addUpvalue adds an upvalue to the compiler's list, avoiding duplicates.
func addUpvalue(compiler *Compiler, index uint8, isLocal bool) int {
	upvalueCount := compiler.function.UpvalueCount
	for i := 0; i < upvalueCount; i++ {
		upvalue := compiler.upvalues[i]
		if upvalue.index == index && upvalue.isLocal == isLocal {
			return i
		}
	}
	if upvalueCount == 256 {
		error("Too many upvalues in this function (max 256).")
		return 0
	}
	compiler.upvalues[upvalueCount] = Upvalue{index: index, isLocal: isLocal}
	compiler.function.UpvalueCount++
	return upvalueCount
}

// resolveUpvalue recursively resolves a variable from enclosing scopes and marks it as captured.
func resolveUpvalue(compiler *Compiler, name token.Token) int {
	if compiler.enclosing == nil {
		return -1
	}
	local := resolveLocal(compiler.enclosing, name)
	if local != -1 {
		compiler.enclosing.locals[local].isCaptured = true
		return addUpvalue(compiler, uint8(local), true)
	}
	upvalue := resolveUpvalue(compiler.enclosing, name)
	if upvalue != -1 {
		return addUpvalue(compiler, uint8(upvalue), false)
	}
	return -1
}

// namedVariable compiles a variable access or assignment, handling locals, upvalues, or globals.
func namedVariable(name token.Token, canAssign bool) {
	var getOp, setOp uint8
	var arg int
	if localArg := resolveLocal(current, name); localArg != -1 {
		arg = localArg
		getOp = byte(runtime.OP_GET_LOCAL)
		setOp = byte(runtime.OP_SET_LOCAL)
	} else if upvalueArg := resolveUpvalue(current, name); upvalueArg != -1 {
		arg = upvalueArg
		getOp = byte(runtime.OP_GET_UPVALUE)
		setOp = byte(runtime.OP_SET_UPVALUE)
	} else {
		arg = int(identifierConstant(name))
		getOp = byte(runtime.OP_GET_GLOBAL)
		setOp = byte(runtime.OP_SET_GLOBAL)
	}
	if canAssign && match(token.TOKEN_EQUAL) {
		expression()
		emitBytes(setOp, uint8(arg))
	} else {
		emitBytes(getOp, uint8(arg))
	}
}

// variable is the entry point for parsing a variable expression.
func variable(canAssign bool) {
	namedVariable(parser.previous, canAssign)
}

// getRule retrieves the parsing rule for a given token type.
func getRule(typ token.TokenType) ParseRule {
	return rules[typ]
}

// initCompiler initializes a new compiler for a function or script and sets up the first local variable.
func initCompiler(compiler *Compiler, funcType FunctionType) {
	compiler.enclosing = current
	compiler.function = nil
	compiler.functionType = funcType
	compiler.localCount = 0
	compiler.scopeDepth = 0
	compiler.function = runtime.NewFunction()
	current = compiler
	if funcType != TYPE_SCRIPT {
		current.function.Name = runtime.CopyString(parser.previous.Start)
	}
	current.localCount++
	local := &current.locals[current.localCount-1]
	local.depth = 0
	local.isCaptured = false
	local.name.Start = ""
	local.name.Length = 0
}

// Compile is the entry point for compiling source code into a function object.
// It initializes the lexer, sets up the compiler state, and processes all declarations.
func Compile(source string) *runtime.ObjFunction {
	lexer.InitLexer(source)
	var compiler Compiler
	initCompiler(&compiler, TYPE_SCRIPT)
	parser.hadError = false
	parser.panicMode = false
	advance()
	for !match(token.TOKEN_EOF) {
		declaration()
	}
	function := endCompiler()
	if !parser.hadError {
		return function
	} else {
		return nil
	}
}

// emitJump writes a jump instruction with a placeholder for the jump offset.
// Returns the offset index where the placeholder was written.
func emitJump(instruction byte) int {
	emitByte(instruction)
	emitByte(0xff)
	emitByte(0xff)
	return currentChunk().Count() - 2
}

// patchJump updates a previously emitted jump instruction with the correct jump offset.
func patchJump(offset int) {
	jump := currentChunk().Count() - offset - 2
	if jump > 65535 {
		error("Jump distance too large (max 65535 bytes). Simplify the code block.")
	}
	currentChunk().Code()[offset] = byte((jump >> 8) & 0xff)
	currentChunk().Code()[offset+1] = byte(jump & 0xff)
}

// and compiles a logical AND operator by emitting short-circuit jump logic.
func and(canAssign bool) {
	endJump := emitJump(byte(runtime.OP_JUMP_IF_FALSE))
	emitByte(byte(runtime.OP_POP))
	parsePrecedence(PREC_AND)
	patchJump(endJump)
}

// or compiles a logical OR operator by emitting appropriate jump instructions.
func or(canAssign bool) {
	elseJump := emitJump(byte(runtime.OP_JUMP_IF_FALSE))
	endJump := emitJump(byte(runtime.OP_JUMP))
	patchJump(elseJump)
	emitByte(byte(runtime.OP_POP))
	parsePrecedence(PREC_OR)
	patchJump(endJump)
}

// emitLoop writes a loop instruction that jumps back to the beginning of the loop.
func emitLoop(loopStart int) {
	emitByte(byte(runtime.OP_LOOP))
	offset := currentChunk().Count() - loopStart + 2
	if offset > 65535 {
		error("Loop body too large (max 65535 bytes). Reduce loop size.")
	}
	emitByte(byte((offset >> 8) & 0xff))
	emitByte(byte(offset & 0xff))
}
