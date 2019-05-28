package main

import "fmt"

func NewParse(l *Lexer) *Parse {
	return &Parse{l: l, t: l.GetNextToken()}
}

type Parse struct {
	l *Lexer
	t *Token
}

func (p *Parse) mustEat(t TokenType) string {
	if p.t.Type == t {
		t := p.t
		p.t = p.l.GetNextToken()
		return t.Value
	}
	fmt.Println(t, p.t)
	panic(p.t.Type)
}

func (p *Parse) eat(t TokenType) (bool, string) {
	if p.t.Type == t {
		t := p.t
		p.t = p.l.GetNextToken()
		return true, t.Value
	}
	return false, ""
}

func (p *Parse) parse() AST {
	return p.program()
}

// program : stmt_list EOF
func (p *Parse) program() AST {
	fmt.Println("program", p.t)
	defer p.mustEat(TokenEOF)
	return p.stmtList()
}

// stmt_list : stmt RETURN | stmt RETURN stmt_list | empty
func (p *Parse) stmtList() AST {
	fmt.Println("stmtList", p.t)
	var list []AST
	for p.t.Type != TokenEOF && p.t.Type != TokenRBrace {
		list = append(list, p.stmt())
		p.mustEat(TokenReturn)
	}
	return ASTStmt{list: list}
}

// stmt : variable (COMMA variable)* ASSIGN expr (COMMA expr)*
func (p *Parse) stmt() AST {
	fmt.Println("stmt", p.t)
	var left []ASTVariable
	left = append(left, p.variable().(ASTVariable))
	for p.t.Type == TokenComma {
		fmt.Println("333333")
		p.mustEat(TokenComma)
		left = append(left, p.variable().(ASTVariable))
	}
	var op = p.mustEat(TokenAssign)
	var right []AST
	right = append(right, p.expr())
	for p.t.Type == TokenComma {
		fmt.Println("444444")
		p.mustEat(TokenComma)
		right = append(right, p.expr())
	}
	return ASTAssign{
		left:  left,
		op:    op,
		right: right,
	}
}

// variable : ID
func (p *Parse) variable() AST {
	fmt.Println("variable", p.t)
	return ASTVariable{name: p.mustEat(TokenID)}
}

// expr : term | term PLUS expr | term MINUS expr
func (p *Parse) expr() AST {
	fmt.Println("expr", p.t)
	left := p.term()
	var op string
	switch p.t.Type {
	case TokenPlus:
		op = p.mustEat(TokenPlus)
	case TokenMinus:
		op = p.mustEat(TokenMinus)
	default:
		return left
	}
	return ASTBinaryOp{
		left:  left,
		op:    op,
		right: p.expr(),
	}
}

// term : factor | factor MUL term | factor DIV term
func (p *Parse) term() AST {
	fmt.Println("term", p.t)
	left := p.factor()
	var op string
	switch p.t.Type {
	case TokenMul:
		op = p.mustEat(TokenMul)
	case TokenDiv:
		op = p.mustEat(TokenDiv)
	default:
		return left
	}
	return ASTBinaryOp{
		left:  left,
		op:    op,
		right: p.term(),
	}
}

// factor : PLUS factor | MINUS factor | NUMBER | LPAREN expr RPAREN | variable
func (p *Parse) factor() AST {
	fmt.Println("factor", p.t)
	switch p.t.Type {
	case TokenPlus:
		return ASTUnaryOp{op: p.mustEat(TokenPlus), AST: p.factor()}
	case TokenMinus:
		return ASTUnaryOp{op: p.mustEat(TokenMinus), AST: p.factor()}
	case TokenNumber:
		return ASTNumber{num: p.mustEat(TokenNumber)}
	case TokenLParen:
		p.mustEat(TokenLParen)
		defer p.mustEat(TokenRParen)
		return p.expr()
	default:
		return p.variable()
	}
}
