package token

type TokenType int

const (
	// Single-character tokens
	TOKEN_LEFT_PAREN TokenType = iota
	TOKEN_RIGHT_PAREN
	TOKEN_LEFT_BRACE
	TOKEN_RIGHT_BRACE
	TOKEN_LEFT_BRACKET
	TOKEN_RIGHT_BRACKET
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_MINUS
	TOKEN_PLUS
	TOKEN_SEMICOLON
	TOKEN_SLASH
	TOKEN_PERCENT
	TOKEN_STAR
	TOKEN_PIPE
	TOKEN_QUESTION
	TOKEN_AT
	TOKEN_HASH
	TOKEN_DOLLAR
	TOKEN_COLON

	// One or two character tokens
	TOKEN_BANG
	TOKEN_BANG_EQUAL
	TOKEN_EQUAL
	TOKEN_EQUAL_EQUAL
	TOKEN_GREATER
	TOKEN_GREATER_EQUAL
	TOKEN_LESS
	TOKEN_LESS_EQUAL
	TOKEN_PLUS_PLUS
	TOKEN_MINUS_MINUS
	TOKEN_STAR_STAR
	TOKEN_FLOOR
	TOKEN_PERCENT_PERCENT

	// Literals
	TOKEN_IDENTIFIER
	TOKEN_USE_TYPE
	TOKEN_CHAR
	TOKEN_STRING
	TOKEN_NUMBER

	// Indentation
	TOKEN_INDENT
	TOKEN_DEDENT
	TOKEN_PASS

	// Keywords
	TOKEN_AND
	TOKEN_STRUCT
	TOKEN_CLASS
	TOKEN_ELSE
	TOKEN_FALSE
	TOKEN_FOR
	TOKEN_FUNC
	TOKEN_IF
	TOKEN_NULL
	TOKEN_OR
	TOKEN_RETURN
	TOKEN_SUPER
	TOKEN_THIS
	TOKEN_TRUE
	TOKEN_VAR
	TOKEN_WHILE
	TOKEN_ITER
	TOKEN_IN
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_MATCH
	TOKEN_WITH
	TOKEN_THROUGH
	TOKEN_RANDOM
	TOKEN_IMPORT
	TOKEN_EXPORT
	TOKEN_DEF
	TOKEN_MOD
	TOKEN_AS
	TOKEN_USE
	TOKEN_ERROR
	TOKEN_EOF
)

type Token struct {
	Type   TokenType
	Start  string
	Length int
	Line   int
}
