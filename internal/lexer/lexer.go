package lexer

import (
	"unicode"
	"unicode/utf8"

	"github.com/cryptrunner49/goseedvm/internal/token"
)

type Lexer struct {
	source  string
	start   int
	current int
	line    int
}

var lexer Lexer

func InitLexer(source string) {
	lexer = Lexer{
		source:  source,
		start:   0,
		current: 0,
		line:    1,
	}
}

func ScanToken() token.Token {
	lexer.skipWhitespace()

	lexer.start = lexer.current
	if lexer.isAtEnd() {
		return lexer.makeToken(token.TOKEN_EOF)
	}

	r := lexer.advance()
	if unicode.IsDigit(r) {
		return lexer.number()
	}

	// Start an identifier if the rune is not an operator or whitespace
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
	case ';':
		return lexer.makeToken(token.TOKEN_SEMICOLON)
	case ',':
		return lexer.makeToken(token.TOKEN_COMMA)
	case '.':
		return lexer.makeToken(token.TOKEN_DOT)
	case '-':
		return lexer.makeToken(token.TOKEN_MINUS)
	case '+':
		return lexer.makeToken(token.TOKEN_PLUS)
	case '/':
		return lexer.makeToken(token.TOKEN_SLASH)
	case '%':
		return lexer.makeToken(token.TOKEN_PERCENT)
	case '*':
		return lexer.makeToken(token.TOKEN_STAR)
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
			return
		}
		if unicode.IsSpace(r) {
			if r == '\n' {
				l.line++
			}
			l.advance()
			continue
		}
		if r == '/' {
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
				return
			}
		} else {
			return
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
	case '(', ')', '{', '}', ';', ',', '.', '-', '+', '/', '%', '@', '#', '$', '*', '!', '=', '<', '>', '"':
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
	case "fn":
		return token.TOKEN_FN
	case "if":
		return token.TOKEN_IF
	case "null":
		return token.TOKEN_NULL
	case "or":
		return token.TOKEN_OR
	case "print":
		return token.TOKEN_PRINT
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
	default:
		return token.TOKEN_IDENTIFIER
	}
}
