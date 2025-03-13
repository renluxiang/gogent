package gogent

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenType int

const (
	illegal tokenType = iota
	eof

	ident // identifiers and functions
	mInt
	mFloat
	mString
	mBool
	lParen
	rParen
	comma
	point
)

func (t tokenType) String() string {
	return tokenTypeToString[t]
}

type token struct {
	Type    tokenType
	Literal string
}

type lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current char under examination
}

type node any

type callExpression struct {
	PackageName  string
	FunctionName string
	Arguments    []node
}

func newLexer(input string) *lexer {
	l := &lexer{input: input}
	l.readChar()
	return l
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func (l *lexer) readChar() {
	l.position = l.readPosition
	if l.readPosition >= len(l.input) {
		l.ch = 0 // eof
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.readPosition += size
	}
}
func (l *lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}
func (l *lexer) nextToken() token {
	var tok token

	l.skipWhitespace()

	if l.ch == 0 { // Check for end of file.
		tok.Type = eof
		tok.Literal = ""
		return tok
	}

	switch l.ch {
	case '(':
		tok = newToken(lParen)
		l.readChar()
	case ')':
		tok = newToken(rParen)
		l.readChar()
	case ',':
		tok = newToken(comma)
		l.readChar()
	case '.':
		tok = newToken(point)
		l.readChar()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		tok.Type, tok.Literal = l.readNumberOrFloat()
	case '"', '\'', '`':
		tok.Type = mString
		tok.Literal = l.readString(l.ch)
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			if tok.Literal == "true" || tok.Literal == "false" {
				tok.Type = mBool
			} else {
				tok.Type = ident
			}
			return tok
		}
		tok.Type = illegal
		tok.Literal = string(l.ch)
	}
	return tok
}

func (l *lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || l.ch == '_' || l.ch == '-' || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *lexer) readNumberOrFloat() (tokenType, string) {
	position := l.position
	hasDot := false
	for isDigit(l.ch) || (l.ch == '.' && !hasDot) {
		if l.ch == '.' {
			hasDot = true
		}
		l.readChar()
	}
	tokenType := mInt
	if hasDot {
		tokenType = mFloat
	}
	return tokenType, l.input[position:l.position]
}

// Read a string until the closing quote.ch is the opening quote.
func (l *lexer) readString(ch rune) string {
	position := l.position + 1 // Skip the opening quote
	for {
		l.readChar()
		if l.ch == ch {
			break
		}
		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar() // Skip the next character
		}
	}
	result := l.input[position:l.position]
	l.readChar()
	return result
}

// Add a mapping from TokenType to its string representation.
var tokenTypeToString = map[tokenType]string{
	illegal: "illegal",
	eof:     "eof",
	ident:   "ident",
	mInt:    "mInt",
	mFloat:  "mFloat",
	mString: "mString",
	mBool:   "mBool",
	lParen:  "(",
	rParen:  ")",
	comma:   ",",
	point:   ".",
}

func newToken(t tokenType) token {
	literal, ok := tokenTypeToString[t]
	if !ok {
		literal = "UNKNOWN"
	}
	return token{Type: t, Literal: literal}
}

// Parser functions
func parse(input string) (*callExpression, error) {
	if strings.HasPrefix(input, "call:") {
		input = input[5:]
	}
	lexer := newLexer(input)
	expr, err := parseCallExpression(lexer)
	if err != nil {
		return nil, err
	}
	if expr.PackageName == "" {
		expr.PackageName = "std"
	}
	return expr, nil
}

func parseCallExpression(l *lexer) (*callExpression, error) {
	result := &callExpression{}
	tok := l.nextToken()
	idents := []string{}
	for tok.Type == ident || tok.Type == point {
		if tok.Type == ident {
			idents = append(idents, tok.Literal)
		}
		tok = l.nextToken()
	}
	if len(idents) == 0 {
		return nil, fmt.Errorf("expected package name, got %v", tok.Type)
	}
	if tok.Type != lParen && tok.Type != eof {
		return nil, fmt.Errorf("expected '(' or eof, got %v %v", tok.Type, tok.Literal)
	}
	result.PackageName = strings.Join(idents[:len(idents)-1], ".")
	result.FunctionName = idents[len(idents)-1]

	var arguments []node
	if tok.Type == eof {
		result.Arguments = arguments
		return result, nil
	}
	tok = l.nextToken()
	for tok.Type != rParen {
		if tok.Type == illegal {
			return nil, fmt.Errorf("unexpected token %v", tok.Type)
		}
		if tok.Type != mInt && tok.Type != mString && tok.Type != mBool && tok.Type != mFloat {
			return nil, fmt.Errorf("expected argument, got type: %v value: %v, argument should be int, string, bool or float", tok.Type.String(), tok.Literal)
		}
		arguments = append(arguments, tok.Literal)
		tok = l.nextToken()
		if tok.Type == comma {
			tok = l.nextToken()
		}
	}
	result.Arguments = arguments
	return result, nil
}
