package main

import (
	"fmt"
	"unicode"
)

type analysisType int

const (
	sNotMatch analysisType = iota
	sSure
	sError

	sStart
	sWhite

	sNum1
	sNum3

	sWord1

	sNote2
	sNote3
	sNote4
	sNote5
	sNote6
	sNote7

	sOp1
	sOp2
	sOp3
	sOp4
	sOp5
	sOp6
)

type Token struct {
	Type  string
	Value string
}

type Analysis struct {
	body    []rune
	Tokens  []Token
	last    Token
	lastPtr int
	nextPtr int
}

func analysis(b []rune) *Analysis {
	var a Analysis
	a.body = b
	a.nextPtr = 0
	var r = a.body[0]
	var s = sStart
	var i = 0
	for r != 0 && i <= 1000 {
		fmt.Println(i, s, r, string(r))
		i++
		switch s {
		case sStart, sNotMatch, sSure:
			s = a.sStart(r)
		case sWhite:
			s = a.sWhite(r)
		case sWord1:
			s = a.sWord1(r)
		case sNum1:
			s = a.sNum1(r)
		case sNum3:
			s = a.sNum3(r)
		case sOp1:
			s = a.sOp1(r)
		case sOp2:
			s = a.sOp2(r)
		case sOp3:
			s = a.sOp3(r)
		case sOp4:
			s = a.sOp4(r)
		case sNote2:
			s = a.sNote2(r)
		case sNote3:
			s = a.sNote3(r)
		case sNote4:
			s = a.sNote4(r)
		case sNote5:
			s = a.sNote5(r)
		case sNote6:
			s = a.sNote6(r)
		case sNote7:
			s = a.sNote7(r)
		default:
			panic("unknown status")
		}
		switch s {
		case sNotMatch:
			a.sureLast()
			r = a.body[0]
		case sSure:
			a.sureLast()
			a.lastPtr -= 1
			r = a.body[0]
		case sError:
			panic("sError")
		default:
			r = a.next()
		}
	}
	return &a
}

func (a *Analysis) setLast(t, v string) {
	a.lastPtr = a.nextPtr + 1
	a.last.Type = t
	a.last.Value = v
}

func (a *Analysis) sureLast() {
	if a.last.Type != "" {
		a.Tokens = append(a.Tokens, a.last)
		a.last.Type = ""
		a.last.Value = ""
	}
	a.body = a.body[a.lastPtr:]
	a.nextPtr, a.lastPtr = 0, 0
}

func (a *Analysis) next() rune {
	a.nextPtr++
	if a.nextPtr >= len(a.body) {
		return 0
	}
	return a.body[a.nextPtr]
}

// start
func (a *Analysis) sStart(nextChar rune) analysisType {
	switch {
	case isWhiteLetter(nextChar):
		return a.sWhite(nextChar)
	case unicode.IsDigit(nextChar):
		return a.sNum1(nextChar)
	case unicode.IsLetter(nextChar):
		return a.sWord1(nextChar)
	default:
		return a.sOp1(nextChar)
	}
}

func (a *Analysis) sWhite(nextChar rune) analysisType {
	switch {
	case isWhiteLetter(nextChar):
		a.setLast("", "")
		return sWhite
	default:
		return sNotMatch
	}
}

// [a-zA-Z0-9_]+
// word & keyword
func (a *Analysis) sWord1(nextChar rune) analysisType {
	switch {
	case unicode.IsLetter(nextChar) || unicode.IsDigit(nextChar) || nextChar == '_':
		if keywords[string(a.body[:a.nextPtr+1])] {
			a.setLast("keyword", string(a.body[:a.nextPtr+1]))
			fmt.Println("keyword", string(a.body[:a.nextPtr+1]))
			return sWord1
		}
		a.setLast("word", string(a.body[:a.nextPtr+1]))
		fmt.Println("word", string(a.body[:a.nextPtr+1]))
		return sWord1
	default:
		return sNotMatch
	}
}

// [0-9_][0-9a-fA-F_xX]*
// int
func (a *Analysis) sNum1(nextChar rune) analysisType {
	switch {
	case nextChar == 'x' || nextChar == 'X':
		fallthrough
	case nextChar >= 'a' && nextChar <= 'z' || nextChar >= 'A' && nextChar <= 'Z':
		fallthrough
	case unicode.IsDigit(nextChar):
		fallthrough
	case a.body[a.nextPtr] == '_':
		// fmt.Println("int---", string(a.body))
		a.setLast("int", string(a.body[:a.nextPtr+1]))
		return sNum1
	case a.body[a.nextPtr] == '.':
		return sNum3
	default:
		return sNotMatch
	}
}

// {{sNum1}\.}[0-9a-fA-F_]+
// float
func (a *Analysis) sNum3(nextChar rune) analysisType {
	switch {
	case nextChar >= 'a' && nextChar <= 'z' || nextChar >= 'A' && nextChar <= 'Z':
		fallthrough
	case unicode.IsDigit(nextChar):
		fallthrough
	case nextChar == '_':
		a.setLast("float", string(a.body[:a.nextPtr+1]))
		return sNum3
	default:
		return sNotMatch
	}
}

func isWhiteLetter(r rune) bool {
	if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
		return true
	}
	return false
}

var keywords = map[string]bool{
	"if": true,
}
