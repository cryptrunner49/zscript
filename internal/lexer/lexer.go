package lexer

import (
	"unicode"

	"github.com/cryptrunner49/gorex/internal/token"
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

	c := lexer.advance()
	if unicode.IsLetter(rune(c)) {
		return lexer.identifier()
	}
	if unicode.IsDigit(rune(c)) {
		return lexer.number()
	}

	switch c {
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

func (l *Lexer) advance() byte {
	l.current++
	return l.source[l.current-1]
}

func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.current]
}

func (l *Lexer) peekNext() byte {
	if l.current+1 >= len(l.source) {
		return 0
	}
	return l.source[l.current+1]
}

func (l *Lexer) match(expected byte) bool {
	if l.isAtEnd() || l.source[l.current] != expected {
		return false
	}
	l.current++
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
		c := l.peek()
		switch c {
		case ' ', '\r', '\t':
			l.advance()
		case '\n':
			l.line++
			l.advance()
		case '/':
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
		default:
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
	for unicode.IsDigit(rune(l.peek())) {
		l.advance()
	}
	if l.peek() == '.' && unicode.IsDigit(rune(l.peekNext())) {
		l.advance() // Consume '.'
		for unicode.IsDigit(rune(l.peek())) {
			l.advance()
		}
	}
	return l.makeToken(token.TOKEN_NUMBER)
}

func (l *Lexer) identifier() token.Token {
	for unicode.IsLetter(rune(l.peek())) || unicode.IsDigit(rune(l.peek())) {
		l.advance()
	}
	return l.makeToken(l.identifierType())
}

func (l *Lexer) identifierType() token.TokenType {
	switch l.source[l.start] {
	case 'a':
		return l.checkKeyword(1, "nd", token.TOKEN_AND)
	case 'c':
		return l.checkKeyword(1, "lass", token.TOKEN_CLASS)
	case 'e':
		return l.checkKeyword(1, "lse", token.TOKEN_ELSE)
	case 'f':
		if l.current-l.start > 1 {
			switch l.source[l.start+1] {
			case 'a':
				return l.checkKeyword(2, "lse", token.TOKEN_FALSE)
			case 'o':
				return l.checkKeyword(2, "r", token.TOKEN_FOR)
			case 'n':
				return l.checkKeyword(1, "n", token.TOKEN_FN)
			}
		}
	case 'i':
		return l.checkKeyword(1, "f", token.TOKEN_IF)
	case 'n':
		return l.checkKeyword(1, "ull", token.TOKEN_NULL)
	case 'o':
		return l.checkKeyword(1, "r", token.TOKEN_OR)
	case 'p':
		return l.checkKeyword(1, "rint", token.TOKEN_PRINT)
	case 'r':
		return l.checkKeyword(1, "eturn", token.TOKEN_RETURN)
	case 's':
		return l.checkKeyword(1, "uper", token.TOKEN_SUPER)
	case 't':
		if l.current-l.start > 1 {
			switch l.source[l.start+1] {
			case 'h':
				return l.checkKeyword(2, "is", token.TOKEN_THIS)
			case 'r':
				return l.checkKeyword(2, "ue", token.TOKEN_TRUE)
			}
		}
	case 'v':
		return l.checkKeyword(1, "ar", token.TOKEN_VAR)
	case 'w':
		return l.checkKeyword(1, "hile", token.TOKEN_WHILE)
	}
	return token.TOKEN_IDENTIFIER
}

func (l *Lexer) checkKeyword(start int, rest string, typ token.TokenType) token.TokenType {
	if l.current-l.start == start+len(rest) &&
		l.source[l.start+start:l.current] == rest {
		return typ
	}
	return token.TOKEN_IDENTIFIER
}
