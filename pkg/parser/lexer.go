package parser

import (
	"strings"
)

type TokenType int

const (
	TOKEN_EOF TokenType = iota
	TOKEN_NEWLINE
	TOKEN_INDENT
	TOKEN_DEDENT
	TOKEN_KEY
	TOKEN_STRING
	TOKEN_DASH
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

type Lexer struct {
	input          string
	position       int
	line           int
	col            int
	indentStack    []int
	pendingDedents int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:       input,
		position:    0,
		line:        1,
		col:         1,
		indentStack: []int{0},
	}
}

func (l *Lexer) peek() byte {
	if l.position >= len(l.input) {
		return 0
	}
	return l.input[l.position]
}

func (l *Lexer) peekNext() byte {
	if l.position+1 >= len(l.input) {
		return 0
	}
	return l.input[l.position+1]
}

func (l *Lexer) advance() byte {
	if l.position >= len(l.input) {
		return 0
	}
	ch := l.input[l.position]
	l.position++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.peek() == ' ' || l.peek() == '\t' {
		l.advance()
	}
}

func (l *Lexer) skipComment() {
	if l.peek() == '#' {
		for l.peek() != '\n' && l.peek() != 0 {
			l.advance()
		}
	}
}

func (l *Lexer) readUntilColon() string {
	start := l.position
	for l.peek() != 0 && l.peek() != ':' && l.peek() != '\n' {
		l.advance()
	}
	return strings.TrimSpace(l.input[start:l.position])
}

func (l *Lexer) readValue() string {
	l.skipWhitespace()
	start := l.position
	for l.peek() != 0 && l.peek() != '\n' && l.peek() != '#' {
		l.advance()
	}
	return strings.TrimSpace(l.input[start:l.position])
}

func (l *Lexer) handleIndentation() *Token {
	indent := 0
	for l.peek() == ' ' {
		l.advance()
		indent++
	}

	if l.peek() == '\n' || l.peek() == '#' || l.peek() == 0 {
		return nil
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if indent > currentIndent {
		l.indentStack = append(l.indentStack, indent)
		return &Token{Type: TOKEN_INDENT, Value: "", Line: l.line, Col: l.col}
	} else if indent < currentIndent {
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.pendingDedents++
		}
		l.pendingDedents--
		return &Token{Type: TOKEN_DEDENT, Value: "", Line: l.line, Col: l.col}
	}

	return nil
}

func (l *Lexer) NextToken() *Token {
	if l.pendingDedents > 0 {
		l.pendingDedents--
		return &Token{Type: TOKEN_DEDENT, Value: "", Line: l.line, Col: l.col}
	}

	l.skipComment()

	if l.peek() == 0 {
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			return &Token{Type: TOKEN_DEDENT, Value: "", Line: l.line, Col: l.col}
		}
		return &Token{Type: TOKEN_EOF, Line: l.line, Col: l.col}
	}

	if l.peek() == '\n' {
		l.advance()
		token := l.handleIndentation()
		if token != nil {
			return token
		}
		return l.NextToken()
	}

	if l.col == 1 {
		token := l.handleIndentation()
		if token != nil {
			return token
		}
	}

	l.skipWhitespace()

	if l.peek() == '-' && (l.peekNext() == ' ' || l.peekNext() == '\t' || l.peekNext() == '\n' || l.peekNext() == 0) {
		line := l.line
		col := l.col
		l.advance()
		l.skipWhitespace()
		return &Token{Type: TOKEN_DASH, Value: "-", Line: line, Col: col}
	}

	identifier := l.readUntilColon()
	if identifier == "" {
		return l.NextToken()
	}

	if l.peek() == ':' {
		line := l.line
		col := l.col
		l.advance()

		value := l.readValue()

		return &Token{Type: TOKEN_KEY, Value: identifier + ":" + value, Line: line, Col: col}
	}

	return &Token{Type: TOKEN_STRING, Value: identifier, Line: l.line, Col: l.col}
}

func (l *Lexer) Tokenize() []Token {
	tokens := []Token{}
	for {
		token := l.NextToken()
		if token == nil {
			continue
		}
		tokens = append(tokens, *token)
		if token.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}
