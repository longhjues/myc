package main

// [^\r\n]*
func (a *Analysis) sNote2(nextChar rune) analysisType {
	switch {
	case nextChar != '\r' && nextChar != '\n':
		return sNote2
	default:
		a.setLast("note", string(a.body[:a.nextPtr]))
		return a.sNote5(nextChar)
	}
}

// -
func (a *Analysis) sNote3(nextChar rune) analysisType {
	switch {
	case nextChar == '-':
		return sNote4
	default:
		return sError
	}
}

// {sNote3}-
func (a *Analysis) sNote4(nextChar rune) analysisType {
	switch {
	case nextChar == '-':
		return sNote5
	default:
		return sError
	}
}

// {sNote2|sNote4}{sNote2}[\r\n]
// note
func (a *Analysis) sNote5(nextChar rune) analysisType {
	switch {
	case nextChar == '\r' || nextChar == '\n':
		return sNote5
	default:
		return sNotMatch
	}
}

// {sNote2}.*(?:*/)
func (a *Analysis) sNote6(nextChar rune) analysisType {
	switch {
	case nextChar == '*':
		return sNote7
	default:
		return sNote6
	}
}

// */
// note
func (a *Analysis) sNote7(nextChar rune) analysisType {
	switch {
	case nextChar == '/':
		a.setLast("note", string(a.body[2:a.nextPtr-1]))
		return sSure
	default:
		return sNote6
	}
}
