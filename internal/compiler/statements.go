package compiler

import (
	"fmt"

	"github.com/cryptrunner49/spy/internal/runtime"
	"github.com/cryptrunner49/spy/internal/token"
)

func expressionStatement() {
	expression()
	consumeOptionalSemicolon()
	emitByte(byte(runtime.OP_POP))
}

func ifStatement() {
	consume(token.TOKEN_LEFT_PAREN, "Expected '(' after 'if'.")
	expression()
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after condition.")
	consume(token.TOKEN_COLON, "Expected ':' after if condition.")
	thenJump := emitJump(byte(runtime.OP_JUMP_IF_FALSE))
	emitByte(byte(runtime.OP_POP))
	block()
	elseJump := emitJump(byte(runtime.OP_JUMP))
	patchJump(thenJump)
	emitByte(byte(runtime.OP_POP))
	if match(token.TOKEN_ELSE) {
		consume(token.TOKEN_COLON, "Expected ':' after else.")
		block()
	}
	patchJump(elseJump)
}

func whileStatement() {
	beginScope()
	consume(token.TOKEN_LEFT_PAREN, "Expected '(' after 'while'.")
	loopStart := currentChunk().Count()
	expression()
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after condition.")
	consume(token.TOKEN_COLON, "Expected ':' after while condition.")
	exitJump := emitJump(byte(runtime.OP_JUMP_IF_FALSE))
	emitByte(byte(runtime.OP_POP))
	current.loops = append(current.loops, Loop{start: loopStart})
	block()
	emitLoop(loopStart)
	currentLoop := current.loops[len(current.loops)-1]
	// Patch continue jumps
	for _, operandPos := range currentLoop.continuePatches {
		opAddress := operandPos - 1
		currentIPAfterOperand := opAddress + 3
		offset := currentIPAfterOperand - loopStart
		if offset < 0 {
			reportError("Invalid continue jump offset.")
			return
		}
		currentChunk().Code()[operandPos] = byte(offset >> 8)
		currentChunk().Code()[operandPos+1] = byte(offset)
	}
	patchJump(exitJump)
	emitByte(byte(runtime.OP_POP))
	// Patch break jumps
	currentLoop.exitAddress = currentChunk().Count()
	for _, operandPos := range currentLoop.exitPatches {
		offset := currentLoop.exitAddress - operandPos - 2
		if offset < 0 {
			reportError("Invalid break jump offset.")
			return
		}
		currentChunk().Code()[operandPos] = byte(offset >> 8)
		currentChunk().Code()[operandPos+1] = byte(offset)
	}
	current.loops = current.loops[:len(current.loops)-1]
	endScope()
}

func forStatement() {
	beginScope()
	consume(token.TOKEN_LEFT_PAREN, "Expected '(' after 'for'.")

	if match(token.TOKEN_SEMICOLON) {
		// No initializer
	} else if match(token.TOKEN_VAR) {
		varDeclaration()
	} else {
		expressionStatement()
	}

	loopStart := currentChunk().Count()
	exitJump := -1

	if !match(token.TOKEN_SEMICOLON) {
		expression()
		consumeOptionalSemicolon()
		exitJump = emitJump(byte(runtime.OP_JUMP_IF_FALSE))
		emitByte(byte(runtime.OP_POP)) // Pop condition result
	}

	current.loops = append(current.loops, Loop{
		jumpType:        JUMP_FOR,
		start:           loopStart,
		exitPatches:     make([]int, 0),
		continuePatches: make([]int, 0),
		hasIncrement:    false,
	})
	currentLoop := &current.loops[len(current.loops)-1]

	bodyJump := -1
	incrementStart := -1
	if !match(token.TOKEN_RIGHT_PAREN) {
		bodyJump = emitJump(byte(runtime.OP_JUMP))
		incrementStart = currentChunk().Count()

		expression() // Increment part
		emitByte(byte(runtime.OP_POP))
		consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after for clauses.")
		consume(token.TOKEN_COLON, "Expected ':' after iter clauses.")

		emitLoop(loopStart)
		loopStart = incrementStart
		patchJump(bodyJump)

		currentLoop.hasIncrement = true
		currentLoop.incrementStart = incrementStart
	}

	block() // Loop body

	emitLoop(loopStart)

	if exitJump != -1 {
		patchJump(exitJump)
		emitByte(byte(runtime.OP_POP)) // Pop condition result
	}

	currentLoop.exitAddress = currentChunk().Count()

	for _, operandPos := range currentLoop.exitPatches {
		opAddress := operandPos - 1
		currentIPAfterOperand := opAddress + 3
		offset := currentLoop.exitAddress - currentIPAfterOperand
		high := byte(offset >> 8)
		low := byte(offset)
		currentChunk().Code()[operandPos] = high
		currentChunk().Code()[operandPos+1] = low
	}

	for _, operandPos := range currentLoop.continuePatches {
		opAddress := operandPos - 1
		currentIPAfterOperand := opAddress + 3
		target := currentLoop.start
		if currentLoop.hasIncrement {
			target = currentLoop.incrementStart
		}
		offset := currentIPAfterOperand - target
		high := byte(offset >> 8)
		low := byte(offset)
		currentChunk().Code()[operandPos] = high
		currentChunk().Code()[operandPos+1] = low
	}

	current.loops = current.loops[:len(current.loops)-1]
	endScope()
}

// breakStatement compiles a break statement, jumping to the end of the innermost loop or match.
func breakStatement() {
	if len(current.loops) == 0 {
		reportError("Cannot use 'break' outside of a loop or match statement.")
		return
	}
	currentLoop := &current.loops[len(current.loops)-1]
	emitByte(byte(runtime.OP_BREAK))
	operandPos := currentChunk().Count()
	emitByte(0xFF)
	emitByte(0xFF)
	currentLoop.exitPatches = append(currentLoop.exitPatches, operandPos)
	consumeOptionalSemicolon()
}

// continueStatement compiles a continue statement, applicable only to loops.
func continueStatement() {
	if len(current.loops) == 0 {
		reportError("Cannot use 'continue' outside of a loop.")
		return
	}
	currentLoop := &current.loops[len(current.loops)-1]
	if currentLoop.jumpType == JUMP_MATCH {
		reportError("Cannot use 'continue' inside a match statement.")
		return
	}
	// Existing continue logic remains unchanged
	emitByte(byte(runtime.OP_CONTINUE))
	jumpPos := currentChunk().Count()
	emitByte(0xFF)
	emitByte(0xFF)
	currentLoop.continuePatches = append(currentLoop.continuePatches, jumpPos)
	consumeOptionalSemicolon()
}

func returnStatement() {
	if current.functionType == TYPE_SCRIPT {
		reportError("Cannot use 'return' outside a function at top-level code.")
	}
	if match(token.TOKEN_SEMICOLON) {
		emitReturn()
	} else {
		expression()
		consumeOptionalSemicolon()
		emitByte(byte(runtime.OP_RETURN))
	}
}

// declareTemporary reserves a temporary local variable with a dummy name.
// It returns the slot number of the temporary local.
func declareTemporary() uint8 {
	dummy := token.Token{Start: "", Length: 0, Line: parser.previous.Line}
	addLocal(dummy)
	markInitialized()
	return uint8(current.localCount - 1)
}

func iterVarDeclaration() {
	// Consume an identifier for the iterator variable.
	consume(token.TOKEN_IDENTIFIER, "Expected iterator variable name.")
	// Declare the variable in the current scope.
	declareVariable()
	// Mark it as initialized.
	markInitialized()
}

func iterStatement() {
	// Start a new scope for the iterator variables to ensure proper cleanup.
	beginScope()

	// Parse the iterator syntax: expect '(' after 'iter'.
	consume(token.TOKEN_LEFT_PAREN, "Expected '(' after 'iter'.")

	// Ensure 'var' keyword follows '(' to declare the iterator variable.
	if !match(token.TOKEN_VAR) {
		reportError("Expected 'var' after '(' in iter statement.")
	}

	// Declare the iterator variable (e.g., 'item') and get its slot in the local scope.
	iterVarDeclaration()
	iterVarSlot := uint8(current.localCount - 1) // Slot for 'item' (typically slot 1).

	// Expect 'in' to separate the variable from the iterable expression.
	consume(token.TOKEN_IN, "Expected 'in' after iterator variable.")

	// Compile the iterable expression (e.g., [1, 2, 3]), leaving it on the stack.
	expression()

	// Expect ')' to close the iterator declaration.
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after condition.")
	consume(token.TOKEN_COLON, "Expected ':' after while condition.")

	// Store the iterable in a temporary local variable to persist across iterations.
	arraySlot := declareTemporary()                  // Slot for the array (typically slot 2).
	emitBytes(byte(runtime.OP_SET_LOCAL), arraySlot) // Assign iterable to local slot.
	emitByte(byte(runtime.OP_POP))                   // Remove iterable from stack.

	// Generate a unique global name for the iterator 'it' (e.g., "__iter_5").
	uniqueIteratorName := fmt.Sprintf("__iter_%d", parser.previous.Line)
	iteratorNameToken := token.Token{Start: uniqueIteratorName, Length: len(uniqueIteratorName), Line: parser.previous.Line}
	iteratorConstant := identifierConstant(iteratorNameToken)

	// Initialize the iterator by calling array_iter(iterable) and store in global 'it'.
	constantIndex := identifierConstant(token.Token{Start: "array_iter", Length: len("array_iter"), Line: parser.previous.Line})
	emitBytes(byte(runtime.OP_GET_GLOBAL), constantIndex)       // Push array_iter function.
	emitBytes(byte(runtime.OP_GET_LOCAL), arraySlot)            // Push the iterable.
	emitBytes(byte(runtime.OP_CALL), 1)                         // Call array_iter, returns iterator.
	emitBytes(byte(runtime.OP_DEFINE_GLOBAL), iteratorConstant) // Define global 'it'.

	// Mark the start of the iteration loop.
	loopStart := currentChunk().Count()

	// Condition: Check if the iterator is done using iter_done(it).
	constantIndex = identifierConstant(token.Token{Start: "iter_done", Length: len("iter_done"), Line: parser.previous.Line})
	emitBytes(byte(runtime.OP_GET_GLOBAL), constantIndex)    // Push iter_done function.
	emitBytes(byte(runtime.OP_GET_GLOBAL), iteratorConstant) // Push global 'it'.
	emitBytes(byte(runtime.OP_CALL), 1)                      // Call iter_done, returns bool.
	exitJump := emitJump(byte(runtime.OP_JUMP_IF_TRUE))      // Jump to end if true (done).
	emitByte(byte(runtime.OP_POP))                           // Pop false result.

	// Get the current value from the iterator using iter_value(it).
	constantIndex = identifierConstant(token.Token{Start: "iter_value", Length: len("iter_value"), Line: parser.previous.Line})
	emitBytes(byte(runtime.OP_GET_GLOBAL), constantIndex)    // Push iter_value function.
	emitBytes(byte(runtime.OP_GET_GLOBAL), iteratorConstant) // Push global 'it'.
	emitBytes(byte(runtime.OP_CALL), 1)                      // Call iter_value, returns value.
	emitBytes(byte(runtime.OP_SET_LOCAL), iterVarSlot)       // Assign value to 'item'.
	// Removed emitByte(byte(runtime.OP_POP)) here to keep the value on the stack.

	// Compile the loop body (e.g., { print item; }).
	block()

	// Advance the iterator to the next element using iter_next(it).
	constantIndex = identifierConstant(token.Token{Start: "iter_next", Length: len("iter_next"), Line: parser.previous.Line})
	emitBytes(byte(runtime.OP_GET_GLOBAL), constantIndex)    // Push iter_next function.
	emitBytes(byte(runtime.OP_GET_GLOBAL), iteratorConstant) // Push global 'it'.
	emitBytes(byte(runtime.OP_CALL), 1)                      // Call iter_next, returns null.
	emitByte(byte(runtime.OP_POP))                           // Pop null result.

	// Loop back to the condition check.
	emitLoop(loopStart)

	// Patch the exit jump to point here when iter_done returns true.
	patchJump(exitJump)

	// Cleanup: Adjust locals for scope exit.
	current.localCount -= 2 // Account for 'item' and 'array'.
	endScope()              // Close scope; no additional pops needed.
}

func matchStatement() {
	// TODO
	fmt.Println("###### TODO ######")
}

func passStatement() {
	consumeOptionalSemicolon()
}
