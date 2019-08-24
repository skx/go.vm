// Package lexer contains our lexer.
package lexer

import (
	"github.com/skx/go.vm/token"
)

// Lexer is used as a lexer for our VM
type Lexer struct {
	position     int    //current character position
	readPosition int    //next character position
	ch           rune   //current character
	characters   []rune //rune slice of input string
}

// New a Lexer instance from string input.
func New(input string) *Lexer {
	l := &Lexer{characters: []rune(input)}
	l.readChar()
	return l
}

// read one forward character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken to read next token, skipping the white space.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	//
	// skip single-line comments
	// Unless they are immediately followed by a number, because
	// our registers are "#N".
	//
	if l.ch == rune('#') {
		if !isDigit(l.peekChar()) {
			l.skipComment()
			return (l.NextToken())
		}
	}

	switch l.ch {
	case rune(','):
		tok = newToken(token.COMMA, l.ch)
	case rune('"'):
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case rune(':'):
		tok.Type = token.LABEL
		tok.Literal = l.readLabel()
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isDigit(l.ch) {
			return l.readDecimal()
		}

		tok.Literal = l.readIdentifier()
		tok.Type = token.LookupIdentifier(tok.Literal)
		return tok
	}
	l.readChar()
	return tok
}

// return new token
func newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// read Identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdentifier(l.ch) {
		l.readChar()
	}
	return string(l.characters[position:l.position])
}

// skip white space
func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// skip comment (until the end of the line).
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

// read number
func (l *Lexer) readNumber() string {
	position := l.position
	for isHexDigit(l.ch) {
		l.readChar()
	}
	return string(l.characters[position:l.position])
}

// read until white space
func (l *Lexer) readUntilWhitespace() string {
	position := l.position
	for !isWhitespace(l.ch) {
		l.readChar()
	}
	return string(l.characters[position:l.position])
}

// read decimal - this needs love to handle decimal and hex.
func (l *Lexer) readDecimal() token.Token {
	integer := l.readNumber()

	if isEmpty(l.ch) || isWhitespace(l.ch) || l.ch == rune(',') {
		return token.Token{Type: token.INT, Literal: integer}
	}
	illegalPart := l.readUntilWhitespace()
	return token.Token{Type: token.ILLEGAL, Literal: integer + illegalPart}
}

// read string
func (l *Lexer) readString() string {
	out := ""

	for {
		l.readChar()
		if l.ch == '"' {
			break
		}

		//
		// Handle \n, \r, \t, \", etc.
		//
		if l.ch == '\\' {
			l.readChar()

			if l.ch == rune('n') {
				l.ch = '\n'
			}
			if l.ch == rune('r') {
				l.ch = '\r'
			}
			if l.ch == rune('t') {
				l.ch = '\t'
			}
			if l.ch == rune('"') {
				l.ch = '"'
			}
			if l.ch == rune('\\') {
				l.ch = '\\'
			}
		}
		out = out + string(l.ch)
	}

	return out
}

func (l *Lexer) readLabel() string {
	return l.readUntilWhitespace()
}

// peek character
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

func isIdentifier(ch rune) bool {
	return ch != rune(',') && !isWhitespace(ch) && !isEmpty(ch)
}

// is white space
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

// is empty
func isEmpty(ch rune) bool {
	return rune(0) == ch
}

// is Digit
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}

func isHexDigit(ch rune) bool {
	if isDigit(ch) {
		return true
	}
	if rune('a') <= ch && ch <= rune('f') {
		return true
	}
	if rune('A') <= ch && ch <= rune('F') {
		return true
	}
	if (rune('x') == ch) || (rune('X') == ch) {
		return true
	}
	return false
}
