package compiler

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/cryptrunner49/goseedvm/internal/runtime"
	"github.com/cryptrunner49/goseedvm/internal/token"
)

func fnDeclaration() {
	global := parseVariable("Expected a function name after 'fn' (e.g., 'fn myFunc()').")
	markInitialized()
	function(TYPE_FUNCTION)
	defineVariable(global)
}

func varDeclaration() {
	global := parseVariable("Expected a variable name after 'var' (e.g., 'var x').")
	if match(token.TOKEN_EQUAL) {
		expression()
	} else {
		emitByte(byte(runtime.OP_NULL))
	}
	consume(token.TOKEN_SEMICOLON, "Expected ';' after variable declaration (e.g., 'var x = 5;').")
	defineVariable(global)
}

func structDeclaration() {
	consume(token.TOKEN_IDENTIFIER, "Expected a struct name after 'struct' (e.g., 'struct Point').")
	nameConstant := identifierConstant(parser.previous)
	declareVariable()

	if match(token.TOKEN_LEFT_BRACE) {
		fieldCount := 0
		fieldNames := make([]*runtime.ObjString, 0)
		fieldDefaults := make([]runtime.Value, 0)

		if !check(token.TOKEN_RIGHT_BRACE) {
			for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
				consume(token.TOKEN_IDENTIFIER, "Expected a field name in struct (e.g., 'x' in 'x = 0').")
				fieldName := runtime.NewObjString(parser.previous.Start)
				fieldNames = append(fieldNames, fieldName)

				var defaultValue runtime.Value
				if match(token.TOKEN_EQUAL) {
					if match(token.TOKEN_NUMBER) {
						val, _ := strconv.ParseFloat(parser.previous.Start, 64)
						defaultValue = runtime.Value{Type: runtime.VAL_NUMBER, Number: val}
					} else if match(token.TOKEN_STRING) {
						text := parser.previous.Start
						str := text[1 : len(text)-1]
						defaultValue = runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(str)}
					} else if match(token.TOKEN_TRUE) {
						defaultValue = runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
					} else if match(token.TOKEN_FALSE) {
						defaultValue = runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
					} else if match(token.TOKEN_NULL) {
						defaultValue = runtime.Value{Type: runtime.VAL_NULL}
					} else {
						reportError("Expected a literal value (number, string, true, false, null) for field default.")
						defaultValue = runtime.Value{Type: runtime.VAL_NULL}
					}
				} else {
					defaultValue = runtime.Value{Type: runtime.VAL_NULL}
				}
				fieldDefaults = append(fieldDefaults, defaultValue)
				fieldCount++

				if !match(token.TOKEN_COMMA) && !check(token.TOKEN_RIGHT_BRACE) {
					consume(token.TOKEN_SEMICOLON, "Expected ',' between fields or '}' to end struct.")
				}
			}
		}

		consume(token.TOKEN_RIGHT_BRACE, "Expected '}' to close struct body (unmatched '{').")
		emitBytes(byte(runtime.OP_STRUCT), nameConstant)
		emitByte(byte(fieldCount))
		for i := 0; i < fieldCount; i++ {
			nameConst := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: fieldNames[i]})
			defaultConst := makeConstant(fieldDefaults[i])
			emitByte(nameConst)
			emitByte(defaultConst)
		}
	} else {
		consume(token.TOKEN_SEMICOLON, "Expected '{' to define fields or ';' for an empty struct.")
		emitBytes(byte(runtime.OP_STRUCT), nameConstant)
		emitByte(0)
	}

	defineVariable(nameConstant)
}

// captureTop emits an opcode to capture the value at the top of the stack as a constant.
// (It does not actually pop it; the VM’s OP_COPY_TOP will later copy the top value.)
func captureTop() uint8 {
	// Here we simulate capturing the value on top by emitting a dummy opcode.
	// In our design, we assume that the value produced by OP_MODULE is in the constant pool.
	// For example, we might emit an OP_COPY_TOP opcode (which our VM would implement to
	// copy the top-of-stack value into a constant slot).
	// For simplicity, we simply add a constant with a dummy value (the value will be overwritten at runtime).
	// In a real implementation you would design an opcode for this purpose.
	return makeConstant(runtime.Value{Type: runtime.VAL_NULL})
}

func compileModuleFunction() runtime.Value {
	var fnCompiler Compiler
	// Initialize a new compiler for this function.
	initCompiler(&fnCompiler, TYPE_FUNCTION, current.scriptDir)
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
	consume(token.TOKEN_RIGHT_PAREN, "Expected ')' to close parameter list.")
	consume(token.TOKEN_LEFT_BRACE, "Expected '{' to start function body.")
	block()
	// Finish the function.
	fnObj := endCompiler()
	// Emit the closure opcode and capture its constant.
	closureConst := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: fnObj})
	emitBytes(byte(runtime.OP_CLOSURE), closureConst)
	for i := 0; i < fnObj.UpvalueCount; i++ {
		isLocal := fnCompiler.upvalues[i].isLocal
		index := fnCompiler.upvalues[i].index
		var byteToEmit byte
		if isLocal {
			byteToEmit = 1
		} else {
			byteToEmit = 0
		}
		emitByte(byteToEmit)
		emitByte(index)
	}
	// Here we simulate retrieving the closure value.
	// (In a full implementation you would fetch the closure from the constant pool,
	// but for our purposes we create a new closure object from fnObj.)
	return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewClosure(fnObj)}
}

func modDeclarationField() (*runtime.ObjString, runtime.Value) {
	// Parse the nested module's name.
	consume(token.TOKEN_IDENTIFIER, "Expected module name in nested module declaration.")
	nestedName := runtime.NewObjString(parser.previous.Start)
	// Do not call declareVariable here.
	consume(token.TOKEN_LEFT_BRACE, "Expected '{' to begin nested module body.")

	// Prepare slices for the nested module's fields.
	nestedFieldNames := make([]*runtime.ObjString, 0)
	nestedFieldDefaults := make([]runtime.Value, 0)

	// Parse declarations inside the nested module body.
	for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
		if match(token.TOKEN_VAR) {
			consume(token.TOKEN_IDENTIFIER, "Expected variable name in nested module.")
			fName := runtime.NewObjString(parser.previous.Start)
			var defVal runtime.Value
			if match(token.TOKEN_EQUAL) {
				if match(token.TOKEN_NUMBER) {
					val, _ := strconv.ParseFloat(parser.previous.Start, 64)
					defVal = runtime.Value{Type: runtime.VAL_NUMBER, Number: val}
				} else if match(token.TOKEN_STRING) {
					text := parser.previous.Start
					str := text[1 : len(text)-1]
					defVal = runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(str)}
				} else if match(token.TOKEN_TRUE) {
					defVal = runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
				} else if match(token.TOKEN_FALSE) {
					defVal = runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
				} else if match(token.TOKEN_NULL) {
					defVal = runtime.Value{Type: runtime.VAL_NULL}
				} else {
					reportError("Expected a literal value for nested module variable initializer.")
					defVal = runtime.Value{Type: runtime.VAL_NULL}
				}
			} else {
				defVal = runtime.Value{Type: runtime.VAL_NULL}
			}
			consume(token.TOKEN_SEMICOLON, "Expected ';' after variable declaration in nested module.")
			nestedFieldNames = append(nestedFieldNames, fName)
			nestedFieldDefaults = append(nestedFieldDefaults, defVal)
		} else if match(token.TOKEN_FN) {
			consume(token.TOKEN_IDENTIFIER, "Expected function name in nested module.")
			fName := runtime.NewObjString(parser.previous.Start)
			markInitialized()
			fnCVal := compileModuleFunction()
			// Optionally consume a semicolon.
			match(token.TOKEN_SEMICOLON)
			nestedFieldNames = append(nestedFieldNames, fName)
			nestedFieldDefaults = append(nestedFieldDefaults, fnCVal)
		} else if match(token.TOKEN_MOD) {
			// Recursively compile further nested modules.
			nName, nVal := modDeclarationField()
			match(token.TOKEN_SEMICOLON)
			nestedFieldNames = append(nestedFieldNames, nName)
			nestedFieldDefaults = append(nestedFieldDefaults, nVal)
		} else {
			reportError("Expected 'var', 'fn', or 'mod' in nested module body.")
			synchronize()
		}
	}
	consume(token.TOKEN_RIGHT_BRACE, "Expected '}' to close nested module body.")

	// Create the nested module object now.
	objModule := runtime.NewModule(nestedName)
	for i := 0; i < len(nestedFieldNames); i++ {
		objModule.Fields[nestedFieldNames[i]] = nestedFieldDefaults[i]
	}
	// Add the module object to the constant pool.
	// (This is a compile-time constant representing the nested module.)
	_ = makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: objModule})
	// Return the nested module's name and its module value.
	return nestedName, runtime.Value{Type: runtime.VAL_OBJ, Obj: objModule}
}

func modDeclaration() {
	// Parse the module name.
	consume(token.TOKEN_IDENTIFIER, "Expected a module name after 'mod'.")
	nameConstant := identifierConstant(parser.previous)
	declareVariable() // Reserve the module name in the current scope.

	// The module must have a body.
	consume(token.TOKEN_LEFT_BRACE, "Expected '{' to define module body.")

	// Prepare slices to record field names and their default (initializer) values.
	fieldNames := make([]*runtime.ObjString, 0)
	fieldDefaults := make([]runtime.Value, 0)

	// Parse declarations (fields) inside the module.
	// A module field can be either a variable declaration, a function declaration, or a nested module.
	for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
		if match(token.TOKEN_VAR) {
			consume(token.TOKEN_IDENTIFIER, "Expected variable name in module declaration.")
			fName := runtime.NewObjString(parser.previous.Start)
			var defVal runtime.Value
			if match(token.TOKEN_EQUAL) {
				if match(token.TOKEN_NUMBER) {
					val, _ := strconv.ParseFloat(parser.previous.Start, 64)
					defVal = runtime.Value{Type: runtime.VAL_NUMBER, Number: val}
				} else if match(token.TOKEN_STRING) {
					text := parser.previous.Start
					str := text[1 : len(text)-1]
					defVal = runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(str)}
				} else if match(token.TOKEN_TRUE) {
					defVal = runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
				} else if match(token.TOKEN_FALSE) {
					defVal = runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
				} else if match(token.TOKEN_NULL) {
					defVal = runtime.Value{Type: runtime.VAL_NULL}
				} else {
					reportError("Expected a literal value for variable initializer in module.")
					defVal = runtime.Value{Type: runtime.VAL_NULL}
				}
			} else {
				defVal = runtime.Value{Type: runtime.VAL_NULL}
			}
			consume(token.TOKEN_SEMICOLON, "Expected ';' after variable declaration in module.")
			fieldNames = append(fieldNames, fName)
			fieldDefaults = append(fieldDefaults, defVal)
		} else if match(token.TOKEN_FN) {
			consume(token.TOKEN_IDENTIFIER, "Expected function name in module declaration.")
			fName := runtime.NewObjString(parser.previous.Start)
			markInitialized()
			fnCVal := compileModuleFunction()
			// Optionally consume a semicolon.
			match(token.TOKEN_SEMICOLON)
			fieldNames = append(fieldNames, fName)
			fieldDefaults = append(fieldDefaults, fnCVal)
		} else if match(token.TOKEN_MOD) {
			// Nested module: "mod d { ... }"
			// Compile the nested module immediately.
			nestedName, nestedVal := modDeclarationField()
			fieldNames = append(fieldNames, nestedName)
			fieldDefaults = append(fieldDefaults, nestedVal)
			// Optionally consume a semicolon.
			match(token.TOKEN_SEMICOLON)
			/*
				} else if match(token.TOKEN_MOD) {
					// Nested module.
					nestedFieldName, nestedModuleConst := modDeclarationNested()
					// Optionally consume a semicolon.
					match(token.TOKEN_SEMICOLON)
					fieldNames = append(fieldNames, nestedFieldName)
					// Use the constant index returned by captureTop() as the field default.
					fieldDefaults = append(fieldDefaults, runtime.Value{Type: runtime.VAL_OBJ, Obj: nil}) // placeholder
					// We then emit an extra opcode to load the nested module.
					// For simplicity, we assume that the constant index in nestedModuleConst
					// will be used at runtime to load the nested module.
					// (In a complete solution, you'd design an opcode (e.g. OP_LOAD_NESTED)
					// that fetches the already–compiled module object.)
					// Here we simulate that by replacing the placeholder with a special marker.
					fieldDefaults[len(fieldDefaults)-1] = runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(fmt.Sprintf("nested:%d", nestedModuleConst))}
			*/
		} else {
			reportError("Expected 'var', 'fn', or 'mod' declaration in module body.")
			synchronize()
		}
	}
	consume(token.TOKEN_RIGHT_BRACE, "Expected '}' to close module body.")

	// Emit the module creation opcode.
	emitBytes(byte(runtime.OP_MODULE), nameConstant)
	emitByte(byte(len(fieldNames))) // Number of fields

	for i := 0; i < len(fieldNames); i++ {
		nameConst := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: fieldNames[i]})
		defConst := makeConstant(fieldDefaults[i])
		emitByte(nameConst)
		emitByte(defConst)
	}

	defineVariable(nameConstant)
}

func importDeclaration() {
	if match(token.TOKEN_STRING) {
		filename := parser.previous.Start[1 : len(parser.previous.Start)-1]
		absPath, errs := filepath.Abs(filepath.Join(current.scriptDir, filename))
		if errs != nil {
			reportError(fmt.Sprintf("Cannot resolve absolute path for '%s': %v", filename, errs))
			return
		}
		pathConstant := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(absPath)})
		emitBytes(byte(runtime.OP_IMPORT), pathConstant)
		consume(token.TOKEN_SEMICOLON, "Expected ';' after import statement.")
	} else {
		var path []string
		consume(token.TOKEN_IDENTIFIER, "Expected identifier after 'import'.")
		path = append(path, parser.previous.Start)
		for match(token.TOKEN_DOT) {
			consume(token.TOKEN_IDENTIFIER, "Expected identifier after '.'.")
			path = append(path, parser.previous.Start)
		}
		consume(token.TOKEN_AS, "Expected 'as' after module path.")
		consume(token.TOKEN_IDENTIFIER, "Expected alias name after 'as'.")
		aliasConstant := identifierConstant(parser.previous)
		emitBytes(byte(runtime.OP_GET_GLOBAL), identifierConstant(token.Token{Start: path[0]}))
		for _, part := range path[1:] {
			emitBytes(byte(runtime.OP_GET_PROPERTY), identifierConstant(token.Token{Start: part}))
		}
		defineVariable(aliasConstant)
		consume(token.TOKEN_SEMICOLON, "Expected ';' after import alias.")
	}
}

func useDeclaration() {
	// TODO
	reportError("The 'use' keyword is not yet implemented.")
}
