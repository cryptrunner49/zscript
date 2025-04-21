package compiler

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/cryptrunner49/spy/internal/runtime"
	"github.com/cryptrunner49/spy/internal/token"
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
	consumeOptionalSemicolon()
	defineVariable(global)
}

func structDeclaration() {
	consume(token.TOKEN_IDENTIFIER, "Expected a struct name after 'struct' (e.g., 'struct Point').")
	nameConstant := identifierConstant(parser.previous)
	declareVariable()

	// If no ':' follows, it's an empty struct
	if !match(token.TOKEN_COLON) {
		consumeOptionalSemicolon()
		emitBytes(byte(runtime.OP_STRUCT), nameConstant)
		emitByte(0) // No fields
		defineVariable(nameConstant)
		return
	}

	if !match(token.TOKEN_INDENT) {
		reportError("Expected indented block after ':' (in struct declaration).")
		return
	}

	fieldCount := 0
	fieldNames := make([]*runtime.ObjString, 0)
	fieldDefaults := make([]runtime.Value, 0)

	for !check(token.TOKEN_DEDENT) && !check(token.TOKEN_EOF) {
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
			} else if match(token.TOKEN_LEFT_BRACKET) {
				// Parse array literal and collect elements
				elements := make([]runtime.Value, 0)
				if !check(token.TOKEN_RIGHT_BRACKET) {
					for {
						if match(token.TOKEN_NUMBER) {
							val, _ := strconv.ParseFloat(parser.previous.Start, 64)
							elements = append(elements, runtime.Value{Type: runtime.VAL_NUMBER, Number: val})
							emitConstant(runtime.Value{Type: runtime.VAL_NUMBER, Number: val})
						} else if match(token.TOKEN_STRING) {
							text := parser.previous.Start
							str := text[1 : len(text)-1]
							objStr := runtime.NewObjString(str)
							elements = append(elements, runtime.Value{Type: runtime.VAL_OBJ, Obj: objStr})
							emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: objStr})
						} else if match(token.TOKEN_TRUE) {
							elements = append(elements, runtime.Value{Type: runtime.VAL_BOOL, Bool: true})
							emitConstant(runtime.Value{Type: runtime.VAL_BOOL, Bool: true})
						} else if match(token.TOKEN_FALSE) {
							elements = append(elements, runtime.Value{Type: runtime.VAL_BOOL, Bool: false})
							emitConstant(runtime.Value{Type: runtime.VAL_BOOL, Bool: false})
						} else if match(token.TOKEN_NULL) {
							elements = append(elements, runtime.Value{Type: runtime.VAL_NULL})
							emitConstant(runtime.Value{Type: runtime.VAL_NULL})
						} else {
							reportError("Array elements must be literals (number, string, true, false, null).")
							elements = append(elements, runtime.Value{Type: runtime.VAL_NULL})
							expression() // Consume invalid expression
						}
						if !match(token.TOKEN_COMMA) {
							break
						}
					}
				}
				consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after array elements.")
				// Create ObjArray and emit OP_ARRAY
				objArray := runtime.NewArray(elements)
				defaultValue = runtime.Value{Type: runtime.VAL_OBJ, Obj: objArray}
				emitBytes(byte(runtime.OP_ARRAY), byte(len(elements)))
			} else if match(token.TOKEN_LEFT_BRACE) {
				// Parse map literal and collect key-value pairs
				pairs := make(map[*runtime.ObjString]runtime.Value)
				for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
					var key *runtime.ObjString
					if match(token.TOKEN_STRING) {
						key = runtime.NewObjString(parser.previous.Start[1 : len(parser.previous.Start)-1])
						emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: key})
					} else if match(token.TOKEN_IDENTIFIER) {
						key = runtime.NewObjString(parser.previous.Start)
						emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: key})
					} else {
						reportError("Map key must be a string or identifier.")
						break
					}
					consume(token.TOKEN_COLON, "Expected ':' after map key.")
					var value runtime.Value
					if match(token.TOKEN_NUMBER) {
						val, _ := strconv.ParseFloat(parser.previous.Start, 64)
						value = runtime.Value{Type: runtime.VAL_NUMBER, Number: val}
						emitConstant(value)
					} else if match(token.TOKEN_STRING) {
						text := parser.previous.Start
						str := text[1 : len(text)-1]
						objStr := runtime.NewObjString(str)
						value = runtime.Value{Type: runtime.VAL_OBJ, Obj: objStr}
						emitConstant(value)
					} else if match(token.TOKEN_TRUE) {
						value = runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
						emitConstant(value)
					} else if match(token.TOKEN_FALSE) {
						value = runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
						emitConstant(value)
					} else if match(token.TOKEN_NULL) {
						value = runtime.Value{Type: runtime.VAL_NULL}
						emitConstant(value)
					} else {
						reportError("Map values must be literals (number, string, true, false, null).")
						value = runtime.Value{Type: runtime.VAL_NULL}
						expression() // Consume invalid expression
					}
					pairs[key] = value
					if !match(token.TOKEN_COMMA) {
						break
					}
				}
				consume(token.TOKEN_RIGHT_BRACE, "Expected '}' after map literal.")
				// Create ObjMap and emit OP_MAP
				objMap := runtime.NewMap()
				for k, v := range pairs {
					objMap.Entries[k] = v
				}
				defaultValue = runtime.Value{Type: runtime.VAL_OBJ, Obj: objMap}
				emitBytes(byte(runtime.OP_MAP), byte(len(pairs)))
			} else {
				reportError("Expected a literal value (number, string, true, false, null, array, or map) for field default.")
				defaultValue = runtime.Value{Type: runtime.VAL_NULL}
			}
		} else {
			defaultValue = runtime.Value{Type: runtime.VAL_NULL}
		}
		fieldDefaults = append(fieldDefaults, defaultValue)
		fieldCount++

		consumeOptionalSemicolon()
	}

	consume(token.TOKEN_DEDENT, "Expected dedent after struct block.")
	emitBytes(byte(runtime.OP_STRUCT), nameConstant)
	emitByte(byte(fieldCount))
	for i := 0; i < fieldCount; i++ {
		nameConst := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: fieldNames[i]})
		defaultConst := makeConstant(fieldDefaults[i])
		emitByte(nameConst)
		emitByte(defaultConst)
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
	consume(token.TOKEN_COLON, "Expected ':' after function parameters.")
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

	return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewClosure(fnObj)}
}

func modDeclarationField() (*runtime.ObjString, runtime.Value) {
	// Parse the nested module's name.
	consume(token.TOKEN_IDENTIFIER, "Expected module name in nested module declaration.")
	nestedName := runtime.NewObjString(parser.previous.Start)

	consume(token.TOKEN_COLON, "Expected ':' after nested module name.")
	consume(token.TOKEN_INDENT, "Expected indented block after ':'.")

	// Prepare slices for the nested module's fields.
	nestedFieldNames := make([]*runtime.ObjString, 0)
	nestedFieldDefaults := make([]runtime.Value, 0)

	// Parse declarations inside the nested module body.
	for !check(token.TOKEN_DEDENT) && !check(token.TOKEN_EOF) {
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
			consumeOptionalSemicolon()
			nestedFieldNames = append(nestedFieldNames, fName)
			nestedFieldDefaults = append(nestedFieldDefaults, defVal)
		} else if match(token.TOKEN_FUNC) {
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
	consume(token.TOKEN_DEDENT, "Expected dedent after nested module block.")

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

	// Optional semicolon to terminate the declaration.
	consumeOptionalSemicolon()
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
		consumeOptionalSemicolon()
		return
	}

	moduleName := modulePathParts[0]
	nameConstant := identifierConstant(token.Token{Start: moduleName})
	declareVariable() // Reserve the module name in the current scope.

	consume(token.TOKEN_COLON, "Expected ':' after module name.")
	consume(token.TOKEN_INDENT, "Expected indented block after ':'.")
	fieldNames := make([]*runtime.ObjString, 0)
	fieldDefaults := make([]runtime.Value, 0)

	// Parse module body
	for !check(token.TOKEN_DEDENT) && !check(token.TOKEN_EOF) {
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
				} else if match(token.TOKEN_LEFT_BRACKET) {
					// Parse array literal and collect elements
					elements := make([]runtime.Value, 0)
					if !check(token.TOKEN_RIGHT_BRACKET) {
						for {
							if match(token.TOKEN_NUMBER) {
								val, _ := strconv.ParseFloat(parser.previous.Start, 64)
								elements = append(elements, runtime.Value{Type: runtime.VAL_NUMBER, Number: val})
								emitConstant(runtime.Value{Type: runtime.VAL_NUMBER, Number: val})
							} else if match(token.TOKEN_STRING) {
								text := parser.previous.Start
								str := text[1 : len(text)-1]
								objStr := runtime.NewObjString(str)
								elements = append(elements, runtime.Value{Type: runtime.VAL_OBJ, Obj: objStr})
								emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: objStr})
							} else if match(token.TOKEN_TRUE) {
								elements = append(elements, runtime.Value{Type: runtime.VAL_BOOL, Bool: true})
								emitConstant(runtime.Value{Type: runtime.VAL_BOOL, Bool: true})
							} else if match(token.TOKEN_FALSE) {
								elements = append(elements, runtime.Value{Type: runtime.VAL_BOOL, Bool: false})
								emitConstant(runtime.Value{Type: runtime.VAL_BOOL, Bool: false})
							} else if match(token.TOKEN_NULL) {
								elements = append(elements, runtime.Value{Type: runtime.VAL_NULL})
								emitConstant(runtime.Value{Type: runtime.VAL_NULL})
							} else {
								reportError("Array elements must be literals (number, string, true, false, null).")
								elements = append(elements, runtime.Value{Type: runtime.VAL_NULL})
								expression() // Consume invalid expression
							}
							if !match(token.TOKEN_COMMA) {
								break
							}
						}
					}
					consume(token.TOKEN_RIGHT_BRACKET, "Expected ']' after array elements.")
					// Create ObjArray and emit OP_ARRAY
					objArray := runtime.NewArray(elements)
					defVal = runtime.Value{Type: runtime.VAL_OBJ, Obj: objArray}
					emitBytes(byte(runtime.OP_ARRAY), byte(len(elements)))
				} else if match(token.TOKEN_LEFT_BRACE) {
					// Parse map literal and collect key-value pairs
					pairs := make(map[*runtime.ObjString]runtime.Value)
					for !check(token.TOKEN_RIGHT_BRACE) && !check(token.TOKEN_EOF) {
						var key *runtime.ObjString
						if match(token.TOKEN_STRING) {
							key = runtime.NewObjString(parser.previous.Start[1 : len(parser.previous.Start)-1])
							emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: key})
						} else if match(token.TOKEN_IDENTIFIER) {
							key = runtime.NewObjString(parser.previous.Start)
							emitConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: key})
						} else {
							reportError("Map key must be a string or identifier.")
							break
						}
						consume(token.TOKEN_COLON, "Expected ':' after map key.")
						var value runtime.Value
						if match(token.TOKEN_NUMBER) {
							val, _ := strconv.ParseFloat(parser.previous.Start, 64)
							value = runtime.Value{Type: runtime.VAL_NUMBER, Number: val}
							emitConstant(value)
						} else if match(token.TOKEN_STRING) {
							text := parser.previous.Start
							str := text[1 : len(text)-1]
							objStr := runtime.NewObjString(str)
							value = runtime.Value{Type: runtime.VAL_OBJ, Obj: objStr}
							emitConstant(value)
						} else if match(token.TOKEN_TRUE) {
							value = runtime.Value{Type: runtime.VAL_BOOL, Bool: true}
							emitConstant(value)
						} else if match(token.TOKEN_FALSE) {
							value = runtime.Value{Type: runtime.VAL_BOOL, Bool: false}
							emitConstant(value)
						} else if match(token.TOKEN_NULL) {
							value = runtime.Value{Type: runtime.VAL_NULL}
							emitConstant(value)
						} else {
							reportError("Map values must be literals (number, string, true, false, null).")
							value = runtime.Value{Type: runtime.VAL_NULL}
							expression() // Consume invalid expression
						}
						pairs[key] = value
						if !match(token.TOKEN_COMMA) {
							break
						}
					}
					consume(token.TOKEN_RIGHT_BRACE, "Expected '}' after map literal.")
					// Create ObjMap and emit OP_MAP
					objMap := runtime.NewMap()
					for k, v := range pairs {
						objMap.Entries[k] = v
					}
					defVal = runtime.Value{Type: runtime.VAL_OBJ, Obj: objMap}
					emitBytes(byte(runtime.OP_MAP), byte(len(pairs)))
				} else {
					reportError("Expected a literal value (number, string, true, false, null, array, or map) for variable initializer in module.")
					defVal = runtime.Value{Type: runtime.VAL_NULL}
				}
			} else {
				defVal = runtime.Value{Type: runtime.VAL_NULL}
			}
			consumeOptionalSemicolon()
			fieldNames = append(fieldNames, fName)
			fieldDefaults = append(fieldDefaults, defVal)
		} else if match(token.TOKEN_FUNC) {
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
	consume(token.TOKEN_DEDENT, "Expected dedent after module block.")

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
		consumeOptionalSemicolon()
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
		consumeOptionalSemicolon()
	}
}

func useDeclaration() {
	// Parse library name: use "mylib"
	consume(token.TOKEN_STRING, "Expected a string literal after 'use' (e.g., 'use \"mylib\";').")
	libName := parser.previous.Start[1 : len(parser.previous.Start)-1] // Remove quotes
	libPathConstant := makeConstant(runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(libName)})

	consume(token.TOKEN_COLON, "Expected ':' after library name in 'use' statement.")
	consume(token.TOKEN_INDENT, "Expected indented block after ':'.")

	// Emit OP_USE with library name
	emitBytes(byte(runtime.OP_USE), libPathConstant)

	// Parse function declarations until '}'
	for !check(token.TOKEN_DEDENT) && !check(token.TOKEN_EOF) {
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

		// Optional semicolon after each function declaration
		consumeOptionalSemicolon()
	}
	consume(token.TOKEN_DEDENT, "Expected dedent after use block.")

	// Optional semicolon after use statement
	consumeOptionalSemicolon()
}
