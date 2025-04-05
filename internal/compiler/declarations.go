package compiler

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/cryptrunner49/goseedvm/internal/runtime"
	"github.com/cryptrunner49/goseedvm/internal/token"
)

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
			reportError(fmt.Sprintf("Variable '%s' is already declared in this scope.", name.Start))
		}
	}
	addLocal(name)
}

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
	//_ = makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: objModule})
	// Return the nested module's name and its module value.
	return nestedName, runtime.Value{Type: runtime.VAL_OBJ, Obj: objModule}
}

func defDeclaration() {
	// Parse the module path (e.g., position, position.x, position.x.z).
	var modulePathParts []string
	consume(token.TOKEN_IDENTIFIER, "Expected an identifier after 'def' (e.g., 'def position' or 'def position.x').")
	modulePathParts = append(modulePathParts, parser.previous.Start)
	for match(token.TOKEN_DOT) {
		consume(token.TOKEN_IDENTIFIER, "Expected identifier after '.' in path (e.g., 'def position.x').")
		modulePathParts = append(modulePathParts, parser.previous.Start)
	}

	// Require 'as' keyword.
	if !match(token.TOKEN_AS) {
		reportError("Expected 'as' after module path in 'def' declaration (e.g., 'def position.x as pos;').")
		return
	}

	// Parse the alias name.
	consume(token.TOKEN_IDENTIFIER, "Expected alias name after 'as' (e.g., 'def position.x as pos;').")
	aliasConstant := identifierConstant(parser.previous)

	// Resolve the module path.
	emitBytes(byte(runtime.OP_GET_GLOBAL), identifierConstant(token.Token{Start: modulePathParts[0]}))
	for i := 1; i < len(modulePathParts); i++ {
		emitBytes(byte(runtime.OP_GET_PROPERTY), identifierConstant(token.Token{Start: modulePathParts[i]}))
	}

	// Define the alias in the current scope.
	defineVariable(aliasConstant)

	// Require semicolon to terminate the declaration.
	consume(token.TOKEN_SEMICOLON, "Expected ';' after 'def' alias declaration (e.g., 'def position.x as pos;').")
}

func modDeclaration() {
	// Parse the module path (e.g., Geometry.Shapes).
	var modulePathParts []string
	consume(token.TOKEN_IDENTIFIER, "Expected module name after 'mod'.")
	modulePathParts = append(modulePathParts, parser.previous.Start)
	for match(token.TOKEN_DOT) {
		consume(token.TOKEN_IDENTIFIER, "Expected identifier after '.' in module path.")
		modulePathParts = append(modulePathParts, parser.previous.Start)
	}

	// Check for alias syntax: "as <alias_name>".
	if match(token.TOKEN_AS) {
		consume(token.TOKEN_IDENTIFIER, "Expected alias name after 'as'.")
		aliasConstant := identifierConstant(parser.previous)

		// Resolve the module path.
		emitBytes(byte(runtime.OP_GET_GLOBAL), identifierConstant(token.Token{Start: modulePathParts[0]}))
		for i := 1; i < len(modulePathParts); i++ {
			emitBytes(byte(runtime.OP_GET_PROPERTY), identifierConstant(token.Token{Start: modulePathParts[i]}))
		}

		// Define the alias in the current scope.
		defineVariable(aliasConstant)
		consume(token.TOKEN_SEMICOLON, "Expected ';' after module alias declaration.")
		return
	}

	// If no alias, proceed with module definition
	if !check(token.TOKEN_LEFT_BRACE) {
		reportError("Expected '{' to define module body or 'as' for aliasing after module name.")
		return
	}

	moduleName := modulePathParts[0]
	nameConstant := identifierConstant(token.Token{Start: moduleName})
	declareVariable() // Reserve the module name in the current scope.

	consume(token.TOKEN_LEFT_BRACE, "Expected '{' to define module body.")
	fieldNames := make([]*runtime.ObjString, 0)
	fieldDefaults := make([]runtime.Value, 0)

	// Parse module body
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
			match(token.TOKEN_SEMICOLON)
			fieldNames = append(fieldNames, fName)
			fieldDefaults = append(fieldDefaults, fnCVal)
		} else if match(token.TOKEN_MOD) {
			// Nested module
			nestedName, nestedVal := modDeclarationField()
			fieldNames = append(fieldNames, nestedName)
			fieldDefaults = append(fieldDefaults, nestedVal)
			match(token.TOKEN_SEMICOLON)
		} else {
			reportError("Expected 'var', 'fn', or 'mod' declaration in module body.")
			synchronize()
		}
	}
	consume(token.TOKEN_RIGHT_BRACE, "Expected '}' to close module body.")

	// Emit module creation
	emitBytes(byte(runtime.OP_MODULE), nameConstant)
	emitByte(byte(len(fieldNames)))
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
	// Parse library name: use "mylib"
	consume(token.TOKEN_STRING, "Expected a string literal after 'use' (e.g., 'use \"mylib\";').")
	libName := parser.previous.Start[1 : len(parser.previous.Start)-1] // Remove quotes
	libPathConstant := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(libName)})

	// Parse opening brace: {
	consume(token.TOKEN_LEFT_BRACE, "Expected '{' after library name in 'use' statement.")

	// Emit OP_USE with library name
	emitBytes(byte(runtime.OP_USE), libPathConstant)

	// Parse function declarations until '}'
	for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
		// Parse return type (e.g., "int", "bool", "size_t")
		consume(token.TOKEN_IDENTIFIER, "Expected return type before function name.")
		returnType := parser.previous.Start
		returnTypeConstant := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(returnType)})

		// Parse function name
		consume(token.TOKEN_IDENTIFIER, "Expected function name after return type.")
		funcName := parser.previous.Start
		funcNameConstant := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(funcName)})

		// Parse parameters: (int, double, etc.)
		consume(token.TOKEN_LEFT_PAREN, "Expected '(' after function name.")
		var paramTypes []string
		if !check(token.TOKEN_RIGHT_PAREN) {
			consume(token.TOKEN_IDENTIFIER, "Expected parameter type.")
			paramTypes = append(paramTypes, parser.previous.Start)
			for match(token.TOKEN_COMMA) {
				consume(token.TOKEN_IDENTIFIER, "Expected parameter type after ','.")
				paramTypes = append(paramTypes, parser.previous.Start)
			}
		}
		consume(token.TOKEN_RIGHT_PAREN, "Expected ')' after parameters.")

		// Emit OP_DEFINE_C_FUNC with function details
		emitByte(byte(runtime.OP_DEFINE_EXTERN))
		emitByte(returnTypeConstant)
		emitByte(byte(len(paramTypes)))
		for _, pt := range paramTypes {
			paramTypeConstant := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(pt)})
			emitByte(paramTypeConstant)
		}
		emitByte(funcNameConstant)

		// Expect semicolon after each function declaration
		consume(token.TOKEN_SEMICOLON, "Expected ';' after function declaration.")
	}

	// Parse closing brace: }
	consume(token.TOKEN_RIGHT_BRACE, "Expected '}' after function declarations.")

	// Expect semicolon after use statement
	consume(token.TOKEN_SEMICOLON, "Expected ';' after 'use' statement.")
}
