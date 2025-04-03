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

func compileModuleInitializer() *runtime.ObjFunction {
	// TODO
	return nil
}

func modDeclaration() {
	// TODO
	reportError("The 'mod' keyword is not yet implemented.")
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
