package main

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

// {sNote2|sNote4}[^\r\n]*[\r\n]
// note
func (a *Analysis) sNote5(nextChar rune) analysisType {
	switch {
	case nextChar == '\r' || nextChar == '\n':
		a.setLast("note", string(a.body[:a.nextPtr]))
		return sNotMatch
	default:
		return sError
	}
}

// {sNote2}.*(?:*/)
func (a *Analysis) sNote6(nextChar rune) analysisType {
	switch {
	case nextChar == '*':
		return sNote7
	default:
		return sError
	}
}

// */
// note
func (a *Analysis) sNote7(nextChar rune) analysisType {
	switch {
	case nextChar == '/':
		a.setLast("note", string(a.body[:a.nextPtr-1]))
		return sNotMatch
	default:
		return sNote6
	}
}
