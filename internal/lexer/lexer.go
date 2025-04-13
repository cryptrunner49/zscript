package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/cryptrunner49/spy/internal/common"
	"github.com/cryptrunner49/spy/internal/token"
)

type Lexer struct {
	source         string
	start          int
	current        int
	line           int
	indents        []int // Stack of indentation levels
	pendingIndents int
	pendingDedents int
	atLineStart    bool
}

var lexer Lexer

func InitLexer(source string) {
	lexer = Lexer{
		source:      source,
		start:       0,
		current:     0,
		line:        1,
		indents:     []int{0},
		atLineStart: true,
	}

	// Skip shebang line if present
	if len(source) >= 2 && source[0] == '#' && source[1] == '!' {
		// Find the end of the first line
		for i := 0; i < len(source); i++ {
			if source[i] == '\n' {
				lexer.current = i + 1 // Move past the newline
				lexer.line = 2        // Actual code starts on line 2
				break
			}
		}
		// If no newline is found, skip the entire source
		if lexer.current == 0 {
			lexer.current = len(source)
		}
	}
}

func ScanToken() token.Token {
	lexer.skipWhitespace()

	if lexer.pendingDedents > 0 {
		lexer.pendingDedents--
		return lexer.makeToken(token.TOKEN_DEDENT)
	}
	if lexer.pendingIndents > 0 {
		lexer.pendingIndents--
		return lexer.makeToken(token.TOKEN_INDENT)
	}

	lexer.start = lexer.current
	if lexer.isAtEnd() {
		if len(lexer.indents) > 1 {
			lexer.pendingDedents = len(lexer.indents) - 1
			lexer.indents = lexer.indents[:1]
			return lexer.makeToken(token.TOKEN_DEDENT)
		}
		return lexer.makeToken(token.TOKEN_EOF)
	}

	r := lexer.advance()
	if unicode.IsDigit(r) {
		return lexer.number()
	}
	if !isOperatorRune(r) && !unicode.IsSpace(r) {
		return lexer.identifier()
	}

	switch r {
	case '(':
		return lexer.makeToken(token.TOKEN_LEFT_PAREN)
	case ')':
		return lexer.makeToken(token.TOKEN_RIGHT_PAREN)
	case '{':
		return lexer.makeToken(token.TOKEN_LEFT_BRACE)
	case '}':
		return lexer.makeToken(token.TOKEN_RIGHT_BRACE)
	case '[':
		return lexer.makeToken(token.TOKEN_LEFT_BRACKET)
	case ']':
		return lexer.makeToken(token.TOKEN_RIGHT_BRACKET)
	case ';':
		return lexer.makeToken(token.TOKEN_SEMICOLON)
	case ',':
		return lexer.makeToken(token.TOKEN_COMMA)
	case '.':
		return lexer.makeToken(token.TOKEN_DOT)
	case '-':
		if lexer.match('-') {
			return lexer.makeToken(token.TOKEN_MINUS_MINUS)
		}
		return lexer.makeToken(token.TOKEN_MINUS)
	case '+':
		if lexer.match('+') {
			return lexer.makeToken(token.TOKEN_PLUS_PLUS)
		}
		return lexer.makeToken(token.TOKEN_PLUS)
	case '*':
		if lexer.match('*') {
			return lexer.makeToken(token.TOKEN_STAR_STAR)
		}
		return lexer.makeToken(token.TOKEN_STAR)
	case '/':
		if lexer.match('_') {
			return lexer.makeToken(token.TOKEN_FLOOR)
		}
		return lexer.makeToken(token.TOKEN_SLASH)
	case '%':
		if lexer.match('%') {
			return lexer.makeToken(token.TOKEN_PERCENT_PERCENT)
		}
		return lexer.makeToken(token.TOKEN_PERCENT)
	case '!':
		if lexer.match('=') {
			return lexer.makeToken(token.TOKEN_BANG_EQUAL)
		}
		return lexer.makeToken(token.TOKEN_BANG)
	case '=':
		if lexer.match('=') {
			return lexer.makeToken(token.TOKEN_EQUAL_EQUAL)
		}
		return lexer.makeToken(token.TOKEN_EQUAL)
	case '<':
		if lexer.match('=') {
			return lexer.makeToken(token.TOKEN_LESS_EQUAL)
		}
		return lexer.makeToken(token.TOKEN_LESS)
	case '>':
		if lexer.match('=') {
			return lexer.makeToken(token.TOKEN_GREATER_EQUAL)
		}
		return lexer.makeToken(token.TOKEN_GREATER)
	case '"':
		return lexer.string()
	case '\'':
		return lexer.char()
	case '|':
		return lexer.makeToken(token.TOKEN_PIPE)
	case '?':
		return lexer.makeToken(token.TOKEN_QUESTION)
	case '@':
		return lexer.makeToken(token.TOKEN_AT)
	case '#':
		return lexer.makeToken(token.TOKEN_HASH)
	case '$':
		return lexer.makeToken(token.TOKEN_DOLLAR)
	case ':':
		return lexer.makeToken(token.TOKEN_COLON)
	}

	return lexer.errorToken("Unexpected character.")
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

func (l *Lexer) advance() rune {
	if l.isAtEnd() {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.source[l.current:])
	l.current += size
	return r
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.current:])
	return r
}

func (l *Lexer) peekNext() rune {
	if l.current >= len(l.source) {
		return 0
	}
	_, size := utf8.DecodeRuneInString(l.source[l.current:])
	if l.current+size >= len(l.source) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.current+size:])
	return r
}

func (l *Lexer) match(expected rune) bool {
	if l.isAtEnd() {
		return false
	}
	r, size := utf8.DecodeRuneInString(l.source[l.current:])
	if r != expected {
		return false
	}
	l.current += size
	return true
}

func (l *Lexer) makeToken(typ token.TokenType) token.Token {
	return token.Token{
		Type:   typ,
		Start:  l.source[l.start:l.current],
		Length: l.current - l.start,
		Line:   l.line,
	}
}

func (l *Lexer) errorToken(message string) token.Token {
	return token.Token{
		Type:   token.TOKEN_ERROR,
		Start:  message,
		Length: len(message),
		Line:   l.line,
	}
}

func (l *Lexer) skipWhitespace() {
	for {
		r := l.peek()
		if r == 0 {
			break
		}
		if r == '\n' {
			l.line++
			l.advance()
			l.processIndentation() // Process indentation for the next line immediately
		} else if unicode.IsSpace(r) {
			l.advance()
		} else if r == '/' {
			// Handle comments
			if l.peekNext() == '/' {
				for l.peek() != '\n' && !l.isAtEnd() {
					l.advance()
				}
			} else if l.peekNext() == '*' {
				l.advance() // Consume '/'
				l.advance() // Consume '*'
				for !l.isAtEnd() {
					if l.peek() == '*' && l.peekNext() == '/' {
						l.advance() // Consume '*'
						l.advance() // Consume '/'
						break
					}
					if l.advance() == '\n' {
						l.line++
					}
				}
			} else {
				break
			}
		} else {
			break
		}
	}
}

func (l *Lexer) string() token.Token {
	for l.peek() != '"' && !l.isAtEnd() {
		if l.peek() == '\n' {
			l.line++
		}
		l.advance()
	}
	if l.isAtEnd() {
		return l.errorToken("Unterminated string.")
	}
	l.advance() // Closing quote
	return l.makeToken(token.TOKEN_STRING)
}

func (l *Lexer) char() token.Token {
	// If we reached the end, we have an error.
	if l.isAtEnd() {
		return l.errorToken("Unterminated character literal.")
	}

	// Read the character value (supporting escape sequences if your language allows them)
	var value []rune

	// Check for escape sequence:
	if l.peek() == '\\' {
		l.advance() // Consume the backslash
		if l.isAtEnd() {
			return l.errorToken("Unterminated escape sequence in character literal.")
		}
		// Consume escaped character (this is simplistic, adapt if you support more escapes)
		value = append(value, rune(l.advance()))
	} else {
		// Normal character
		value = append(value, rune(l.advance()))
	}

	// Make sure there is no more than one character for a valid character literal.
	// Optionally: you might want to issue an error if more than one character is encountered
	// before the closing quote.
	if !l.isAtEnd() && l.peek() != '\'' {
		// Optional: consume characters until the closing single quote (or end) to sync error recovery.
		for l.peek() != '\'' && !l.isAtEnd() {
			l.advance()
		}
		if l.isAtEnd() {
			return l.errorToken("Unterminated character literal.")
		}
		return l.errorToken("Character literal must contain exactly one character.")
	}

	// Expect and consume the closing single quote.
	if l.isAtEnd() || l.peek() != '\'' {
		return l.errorToken("Unterminated character literal.")
	}
	l.advance() // Consume the closing quote

	// Assuming that l.makeToken takes the token type and uses the text from start to the current position.
	return l.makeToken(token.TOKEN_CHAR)
}

func (l *Lexer) number() token.Token {
	for unicode.IsDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && unicode.IsDigit(l.peekNext()) {
		l.advance() // Consume '.'
		for unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}
	return l.makeToken(token.TOKEN_NUMBER)
}

func (l *Lexer) identifier() token.Token {
	for {
		r := l.peek()
		if r == 0 || unicode.IsSpace(r) || isOperatorRune(r) {
			break
		}
		l.advance()
	}
	return l.makeToken(l.identifierType())
}

func isOperatorRune(r rune) bool {
	switch r {
	case '(', ')', '{', '}', '[', ']', '|', ':', '?', ';', ',', '.', '-', '+', '/', '%', '@', '#', '$', '*', '!', '=', '<', '>', '"', '\'':
		return true
	default:
		return false
	}
}

func (l *Lexer) identifierType() token.TokenType {
	startStr := l.source[l.start:l.current]
	switch startStr {
	case "and":
		return token.TOKEN_AND
	case "else":
		return token.TOKEN_ELSE
	case "false":
		return token.TOKEN_FALSE
	case "for":
		return token.TOKEN_FOR
	case "func":
		return token.TOKEN_FUNC
	case "if":
		return token.TOKEN_IF
	case "null":
		return token.TOKEN_NULL
	case "or":
		return token.TOKEN_OR
	case "return":
		return token.TOKEN_RETURN
	case "struct":
		return token.TOKEN_STRUCT
	case "this":
		return token.TOKEN_THIS
	case "true":
		return token.TOKEN_TRUE
	case "var":
		return token.TOKEN_VAR
	case "while":
		return token.TOKEN_WHILE
	case "iter":
		return token.TOKEN_ITER
	case "in":
		return token.TOKEN_IN
	case "break":
		return token.TOKEN_BREAK
	case "continue":
		return token.TOKEN_CONTINUE
	case "match":
		return token.TOKEN_MATCH
	case "with":
		return token.TOKEN_WITH
	case "through":
		return token.TOKEN_THROUGH
	case "import":
		return token.TOKEN_IMPORT
	case "export":
		return token.TOKEN_EXPORT
	case "use":
		return token.TOKEN_USE
	case "def":
		return token.TOKEN_DEF
	case "mod":
		return token.TOKEN_MOD
	case "as":
		return token.TOKEN_AS
	case "pass":
		return token.TOKEN_PASS
	case "enable_debug_indent":
		common.DebugIndent = true
		return token.TOKEN_IDENTIFIER
	case "disable_debug_indent":
		common.DebugIndent = false
		return token.TOKEN_IDENTIFIER
	default:
		return token.TOKEN_IDENTIFIER
	}
}

func (l *Lexer) processIndentation() {
	// Save the position after the newline
	startOfLine := l.current

	// Count leading whitespace
	indentLevel := 0
	for {
		r := l.peek()
		if r == ' ' {
			indentLevel++
			l.advance()
		} else if r == '\t' {
			indentLevel += 4 // Tabs count as 4 spaces
			l.advance()
		} else {
			break
		}
	}

	if common.DebugIndent {
		fmt.Printf("[Line %d]: Indent level = %d\n", l.line, indentLevel)
	}

	// Check if the line is blank or a comment
	isBlank := true
	for {
		r := l.peek()
		if r == '\n' || r == 0 {
			break
		}
		if !unicode.IsSpace(r) {
			if r == '/' && l.peekNext() == '/' {
				// Comment line
				for l.peek() != '\n' && !l.isAtEnd() {
					l.advance()
				}
			} else {
				isBlank = false
			}
			break
		}
		l.advance()
	}

	if isBlank {
		// Blank line: reset position and skip indentation adjustment
		l.current = startOfLine
		return
	}

	// Adjust indentation levels for non-blank lines
	currentLevel := l.indents[len(l.indents)-1]
	if indentLevel > currentLevel {
		l.indents = append(l.indents, indentLevel)
		l.pendingIndents++
	} else if indentLevel < currentLevel {
		for len(l.indents) > 1 && l.indents[len(l.indents)-1] > indentLevel {
			l.indents = l.indents[:len(l.indents)-1]
			l.pendingDedents++
		}
		if l.indents[len(l.indents)-1] != indentLevel {
			l.errorToken("Indentation mismatch")
		}
	}
	// l.current is already at the first non-space character after indentation
}
