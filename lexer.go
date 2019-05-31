package main

import (
	"strings"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenEnter
	TokenPlus
	TokenMinus
	TokenMul
	TokenDiv
	TokenID
	TokenNumber
	TokenIf
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenAssign
	TokenComma
	TokenColon
	TokenVar
	TokenElse
	TokenThen
	TokenAnd
	TokenOr
	TokenNot
	TokenAndSlower
	TokenOrSlower
	TokenNotSlower
	TokenCompare
	TokenFunction
	TokenReturn
)

var KeyWords = map[string]*Token{
	"var":    {Type: TokenVar},
	"if":     {Type: TokenIf},
	"then":   {Type: TokenThen},
	"else":   {Type: TokenElse},
	"and":    {Type: TokenAndSlower},
	"or":     {Type: TokenOrSlower},
	"not":    {Type: TokenNotSlower},
	"func":   {Type: TokenFunction},
	"return": {Type: TokenReturn},
}

type Token struct {
	Type  TokenType
	Value string
}

func NewLexer(b []byte) *Lexer {
	return &Lexer{b: b}
}

type Lexer struct {
	b   []byte
	pos int
}

func (l *Lexer) Advance() byte {
	if l.pos >= len(l.b) {
		return 0
	}
	l.pos++
	return l.b[l.pos-1]
}

// AdvanceUntil do not contain c
func (l *Lexer) AdvanceUntil(c byte) int {
	var n int
	for {
		if l.pos >= len(l.b) || l.b[l.pos] == c {
			break
		}
		l.pos++
		n++
	}
	return n
}

func (l *Lexer) Peek() byte {
	if l.pos >= len(l.b) {
		return 0
	}
	return l.b[l.pos]
}

func (l *Lexer) GetNextToken() *Token {
	var c = l.Advance()
	switch c {
	case 0: // eof
		return &Token{Type: TokenEOF}
	case ' ', '\t': // white spec
		return l.GetNextToken()
	case '\r', '\n':
		c = l.Peek()
		for c != 0 && (c == ' ' || c == '\t' || c == '\r' || c == '\n') {
			l.Advance()
			c = l.Peek()
		}
		return &Token{Type: TokenEnter}
	case '+':
		if l.Peek() == '=' {
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()})}
		}
		return &Token{Type: TokenPlus, Value: "+"}
	case '-':
		// if l.Peek() == '-' {
		// 	l.AdvanceUntil('\n')
		// 	return l.GetNextToken()
		// }
		if l.Peek() == '=' {
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()})}
		}
		return &Token{Type: TokenMinus, Value: "-"}
	case '*':
		if l.Peek() == '=' {
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()})}
		}
		return &Token{Type: TokenMul, Value: "*"}
	case '/':
		if l.Peek() == '/' {
			l.AdvanceUntil('\n')
			return l.GetNextToken()
		}
		if l.Peek() == '=' {
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()})}
		}
		return &Token{Type: TokenDiv, Value: "/"}
	case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		var num = []byte{c}
		for {
			c = l.Peek()
			if !strings.Contains("1234567890abcdef._oxb", string(c)) {
				return &Token{Type: TokenNumber, Value: string(num)}
			}
			num = append(num, l.Advance())
		}
	case '(':
		return &Token{Type: TokenLParen}
	case ')':
		return &Token{Type: TokenRParen}
	case '{':
		return &Token{Type: TokenLBrace}
	case '}':
		return &Token{Type: TokenRBrace}
	case '=':
		if l.Peek() == '=' {
			l.Advance()
			return &Token{Type: TokenCompare, Value: "=="}
		}
		return &Token{Type: TokenAssign, Value: "="}
	case ',':
		return &Token{Type: TokenComma}
	case ':':
		return &Token{Type: TokenColon}
	case '<':
		if l.Peek() == '=' {
			l.Advance()
			return &Token{Type: TokenCompare, Value: "<="}
		}
		return &Token{Type: TokenCompare, Value: "<"}
	case '>':
		if l.Peek() == '=' {
			l.Advance()
			return &Token{Type: TokenCompare, Value: ">="}
		}
		return &Token{Type: TokenCompare, Value: ">"}
	case '&':
		if l.Peek() == '&' {
			l.Advance()
			return &Token{Type: TokenAnd}
		}
	case '|':
		if l.Peek() == '|' {
			l.Advance()
			return &Token{Type: TokenOr}
		}
	case '!':
		if l.Peek() == '=' {
			l.Advance()
			return &Token{Type: TokenCompare, Value: "!="}
		}
		return &Token{Type: TokenNot}
	}

	// ID
	var id = []byte{c}
	for {
		c = l.Peek()
		if c == 0 || strings.Contains(" \\\t\r\n\"';:`~!@#$%^&*()+-=|{}[]<>,./?", string(c)) {
			if t, ok := KeyWords[string(id)]; ok {
				return t
			}
			return &Token{Type: TokenID, Value: string(id)}
		}
		id = append(id, l.Advance())
	}
}
