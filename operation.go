package main

// [+-*/()[]{}!&]
func (a *Analysis) sOp1(nextChar rune) analysisType {
	switch nextChar {
	case '+':
		a.setLast("op", "+")
		return sOp3
	case '-':
		a.setLast("op", "-")
		return sOp3
	case '*':
		a.setLast("op", "*")
		return sOp3
	case '/':
		a.setLast("op", "/")
		return sOp2
	case '=':
		a.setLast("op", "=")
		return sOp4
	case '(', ')', '[', ']', '{', '}', '!', '&':
		a.setLast("op", string(a.body[:a.nextPtr+1]))
		return sSure
	default:
		return sError
	}
}

// {/}[/=*]?
func (a *Analysis) sOp2(nextChar rune) analysisType {
	switch nextChar {
	case '/':
		return sNote5
	case '*':
		return sNote6
	case '=':
		a.setLast("op", "/=")
		return sSure
	default:
		return sNotMatch
	}
}

// {+-*}[=]?
func (a *Analysis) sOp3(nextChar rune) analysisType {
	switch nextChar {
	case '=':
		a.setLast("op", string(a.body[:a.nextPtr+1]))
		return sSure
	default:
		return sNotMatch
	}
}

// {=}[=]?
func (a *Analysis) sOp4(nextChar rune) analysisType {
	switch nextChar {
	case '=':
		a.setLast("op", "==")
		return sSure
	default:
		return sNotMatch
	}
}
