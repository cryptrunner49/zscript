package compiler

import (
	"github.com/cryptrunner49/goseedvm/internal/runtime"
	"github.com/cryptrunner49/goseedvm/internal/token"
)

func printStatement() {
	expression()
	consume(token.TOKEN_SEMICOLON, "Expected ';' after value in print statement (e.g., 'print x;').")
	emitByte(byte(runtime.OP_PRINT))
}

func expressionStatement() {
	expression()
	consume(token.TOKEN_SEMICOLON, "Expected ';' after expression (e.g., 'x + 1;').")
	emitByte(byte(runtime.OP_POP))
}

func ifStatement() {
	consume(token.TOKEN_LEFT_PAREN, "Expected '(' after 'if' to start condition.")
	expression()
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after if condition (e.g., 'if (x > 0)').")
	thenJump := emitJump(byte(runtime.OP_JUMP_IF_FALSE))
	emitByte(byte(runtime.OP_POP))
	statement()
	elseJump := emitJump(byte(runtime.OP_JUMP))
	patchJump(thenJump)
	emitByte(byte(runtime.OP_POP))
	if match(token.TOKEN_ELSE) {
		statement()
	}
	patchJump(elseJump)
}

func whileStatement() {
	beginScope()
	consume(token.TOKEN_LEFT_PAREN, "Expected '(' after 'while'.")

	// Set loopStart before the condition
	loopStart := currentChunk().Count() // Will be 0017

	expression() // Emits condition (0017–0021)
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after condition.")

	exitJump := emitJump(byte(runtime.OP_JUMP_IF_FALSE)) // 0022
	emitByte(byte(runtime.OP_POP))                       // 0025

	// Track loop for continue/break
	current.loops = append(current.loops, Loop{
		loopType:        LOOP_WHILE,
		start:           loopStart,
		exitPatches:     make([]int, 0),
		continuePatches: make([]int, 0),
	})
	currentLoop := &current.loops[len(current.loops)-1]

	statement() // Body (0026–0056)

	// Jump back to condition
	emitLoop(loopStart) // Jumps to 0017

	// Patch continue jumps to loopStart
	for _, operandPos := range currentLoop.continuePatches {
		opAddress := operandPos - 1
		currentIPAfterOperand := opAddress + 3
		offset := currentIPAfterOperand - loopStart
		if offset < 0 || offset > 65535 {
			error("Continue jump offset out of range.")
		}
		high := byte(offset >> 8)
		low := byte(offset)
		currentChunk().Code()[operandPos] = high
		currentChunk().Code()[operandPos+1] = low
	}

	// Patch exit jump
	patchJump(exitJump)
	emitByte(byte(runtime.OP_POP))

	// Patch break jumps
	currentLoop.exitAddress = currentChunk().Count()
	for _, patchPos := range currentLoop.exitPatches {
		patchJump(patchPos)
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
		consume(token.TOKEN_SEMICOLON, "Expected ';' after loop condition.")
		exitJump = emitJump(byte(runtime.OP_JUMP_IF_FALSE))
		emitByte(byte(runtime.OP_POP)) // Pop condition result
	}

	current.loops = append(current.loops, Loop{
		loopType:        LOOP_FOR,
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

		emitLoop(loopStart)
		loopStart = incrementStart
		patchJump(bodyJump)

		currentLoop.hasIncrement = true
		currentLoop.incrementStart = incrementStart
	}

	statement() // Loop body

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

func breakStatement() {
	if len(current.loops) == 0 {
		error("Cannot use 'break' outside of a loop.")
		return
	}
	currentLoop := &current.loops[len(current.loops)-1]
	emitByte(byte(runtime.OP_BREAK))
	operandPos := currentChunk().Count()
	emitByte(0xFF)
	emitByte(0xFF)
	currentLoop.exitPatches = append(currentLoop.exitPatches, operandPos)
	consume(token.TOKEN_SEMICOLON, "Expected ';' after 'break'.")
}

func continueStatement() {
	if len(current.loops) == 0 {
		error("Cannot use 'continue' outside of a loop.")
		return
	}
	currentLoop := &current.loops[len(current.loops)-1]

	// Emit continue instruction
	emitByte(byte(runtime.OP_CONTINUE))

	// Calculate jump offset (will be patched later)
	jumpPos := currentChunk().Count()
	emitByte(0xFF) // placeholder for jump offset
	emitByte(0xFF)

	// Record this continue for later patching
	currentLoop.continuePatches = append(currentLoop.continuePatches, jumpPos)

	consume(token.TOKEN_SEMICOLON, "Expected ';' after 'continue'.")
}

func returnStatement() {
	if current.functionType == TYPE_SCRIPT {
		error("Cannot use 'return' outside a function at top-level code.")
	}
	if match(token.TOKEN_SEMICOLON) {
		emitReturn()
	} else {
		expression()
		consume(token.TOKEN_SEMICOLON, "Expected ';' after return value (e.g., 'return 42;').")
		emitByte(byte(runtime.OP_RETURN))
	}
}

func matchStatement() {

}
