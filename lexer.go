package main

import (
	"fmt"
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
	TokenString

	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenAssign
	TokenComma
	TokenColon
	TokenDot

	TokenIf
	TokenVar
	TokenElse
	TokenThen
	TokenAs
	TokenImport
	TokenAndSlower
	TokenOrSlower
	TokenNotSlower
	TokenFunction
	TokenReturn

	TokenAnd
	TokenOr
	TokenNot
	TokenCompare

	TokenOpBit
	TokenOpAnd
	TokenUnaryOp
)

var KeyWords = map[string]*Token{
	"var":    {Type: TokenVar,Value:"VAR"},
	"if":     {Type: TokenIf,Value:"IF"},
	"then":   {Type: TokenThen,Value:"THEN"},
	"else":   {Type: TokenElse,Value:"ELSE"},
	"and":    {Type: TokenAndSlower,Value:"AND"},
	"or":     {Type: TokenOrSlower,Value:"OR"},
	"not":    {Type: TokenNotSlower,Value:"NOT"},
	"func":   {Type: TokenFunction,Value:"FUNC"},
	"return": {Type: TokenReturn,Value:"RETURN"},
	"as":     {Type: TokenAs,Value:"AS"},
	"import": {Type: TokenImport,Value:"IMPORT"},
}

type Token struct {
	Type  TokenType
	Value string

	line   int
	offset int
}

func (t Token) String() string {
	return fmt.Sprintf("(%d:%d %v:%v)", t.line,t.offset,t.Type, t.Value)
}

func NewLexer(b []byte) *Lexer {
	return &Lexer{b: b}
}

type Lexer struct {
	b   []byte
	pos int
	line int
	offset int
}

func (l *Lexer) LexerToken() []*Token {
	var t []*Token
	var v = l.GetNextToken()
	for ; v.Type != TokenEOF; v = l.GetNextToken() {
		t = append(t, v)
	}
	t = append(t, v)
	return t
}

func (l *Lexer) Advance() byte {
	if l.pos >= len(l.b) {
		return 0
	}
	var b =l.b[l.pos]
	if b=='\n'{
		l.line++
		l.offset=0
	}
	fmt.Print(string(b))
	l.pos++
	l.offset++
	return b
}

// AdvanceUntil do not contain c
func (l *Lexer) AdvanceUntil(c byte) int {
	var n int
	for {
		if l.pos >= len(l.b) || l.b[l.pos] == c {
			break
		}
		if l.b[l.pos]=='\n'{
			l.line++
			l.offset=0
		}
		l.pos++
		l.offset++
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
		fmt.Print(".")
		return &Token{Type: TokenEOF,Value: "EOF",line: l.line,offset: l.offset}
	case ' ', '\t': // white spec
		fmt.Print(".")
		return l.GetNextToken()
	case '\r', '\n':
		c = l.Peek()
		for c != 0 && (c == ' ' || c == '\t' || c == '\r' || c == '\n') {
			l.Advance()
			c = l.Peek()
		}
		fmt.Print(".")
		return &Token{Type: TokenEnter,Value: "ENTER",line: l.line,offset: l.offset}
	case '+':
		if l.Peek() == '=' {
			fmt.Print(".")
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()}),line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenPlus, Value: "+",line: l.line,offset: l.offset}
	case '-':
		// if l.Peek() == '-' {
		// 	l.AdvanceUntil('\n')
		// 	return l.GetNextToken()
		// }
		if l.Peek() == '=' {
			fmt.Print(".")
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()}),line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenMinus, Value: "-",line: l.line,offset: l.offset}
	case '*':
		if l.Peek() == '=' {
			fmt.Print(".")
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()}),line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenMul, Value: "*",line: l.line,offset: l.offset}
	case '/':
		if l.Peek() == '/' {
			l.AdvanceUntil('\n')
			fmt.Print(".")
			return l.GetNextToken()
		}
		if l.Peek() == '=' {
			fmt.Print(".")
			return &Token{Type: TokenAssign, Value: string([]byte{c, l.Advance()}),line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenDiv, Value: "/",line: l.line,offset: l.offset}
	case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		var num = []byte{c}
		for {
			c = l.Peek()
			if !strings.Contains("1234567890abcdef._oxb", string(c)) {
				fmt.Print(".")
				return &Token{Type: TokenNumber, Value: string(num),line: l.line,offset: l.offset}
			}
			num = append(num, l.Advance())
		}
	case '"': // String
		var s []byte
		for {
			c = l.Peek()
			if c == '\\' { // TODO: 处理转移字符
				l.Advance()
			} else if c == '"' {
				l.Advance()
				fmt.Print(".")
				return &Token{Type: TokenString, Value: string(s),line: l.line,offset: l.offset}
			}
			s = append(s, l.Advance())
		}
	case '(':
		fmt.Print(".")
		return &Token{Type: TokenLParen,Value: "(",line: l.line,offset: l.offset}
	case ')':
		fmt.Print(".")
		return &Token{Type: TokenRParen,Value: ")",line: l.line,offset: l.offset}
	case '{':
		fmt.Print(".")
		return &Token{Type: TokenLBrace,Value: "{",line: l.line,offset: l.offset}
	case '}':
		fmt.Print(".")
		return &Token{Type: TokenRBrace,Value: "}",line: l.line,offset: l.offset}
	case '=':
		if l.Peek() == '=' {
			l.Advance()
			fmt.Print(".")
			return &Token{Type: TokenCompare, Value: "==",line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenAssign, Value: "=",line: l.line,offset: l.offset}
	case ',':
		fmt.Print(".")
		return &Token{Type: TokenComma, Value: ",",line: l.line,offset: l.offset}
	case '.':
		fmt.Print(".")
		return &Token{Type: TokenDot, Value: ".",line: l.line,offset: l.offset}
	case ':':
		fmt.Print(".")
		return &Token{Type: TokenColon, Value: ":",line: l.line,offset: l.offset}
	case '<':
		if l.Peek() == '=' {
			l.Advance()
			fmt.Print(".")
			return &Token{Type: TokenCompare, Value: "<=",line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenCompare, Value: "<",line: l.line,offset: l.offset}
	case '>':
		if l.Peek() == '=' {
			l.Advance()
			fmt.Print(".")
			return &Token{Type: TokenCompare, Value: ">=",line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenCompare, Value: ">",line: l.line,offset: l.offset}
	case '&':
		if l.Peek() == '&' {
			l.Advance()
			fmt.Print(".")
			return &Token{Type: TokenAnd,Value: "&&",line: l.line,offset: l.offset}
		}
	case '|':
		if l.Peek() == '|' {
			l.Advance()
			fmt.Print(".")
			return &Token{Type: TokenOr,Value: "||",line: l.line,offset: l.offset}
		}
	case '!':
		if l.Peek() == '=' {
			l.Advance()
			fmt.Print(".")
			return &Token{Type: TokenCompare, Value: "!=",line: l.line,offset: l.offset}
		}
		fmt.Print(".")
		return &Token{Type: TokenNot,Value: "!",line: l.line,offset: l.offset}
	}

	// ID
	var id = []byte{c}
	for {
		c = l.Peek()
		if c == 0 || strings.Contains(" \\\t\r\n\"';:`~!@#$%^&*()+-=|{}[]<>,./?", string(c)) {
			if t, ok := KeyWords[string(id)]; ok {
				fmt.Print(".")
				return t
			}
			fmt.Print(".")
			return &Token{Type: TokenID, Value: string(id),line: l.line,offset: l.offset}
		}
		id = append(id, l.Advance())
	}
}
