package main

import (
	"log"
)

func NewParse(l *Lexer) *Parse {
	return &Parse{l: l, t: l.GetNextToken()}
}

type Parse struct {
	l         *Lexer
	t         *Token
	peekToken *Token
}

func (p *Parse) peek() TokenType {
	if p.peekToken == nil {
		p.peekToken = p.l.GetNextToken()
	}
	return p.peekToken.Type
}

func (p *Parse) mustEat(t TokenType) string {
	log.Println(t, p.t)
	if p.t.Type == t {
		t := p.t
		if p.peekToken != nil {
			p.t, p.peekToken = p.peekToken, nil
		} else {
			p.t = p.l.GetNextToken()
		}
		return t.Value
	}
	panic(p.t.Type)
}

// func (p *Parse) eat(t TokenType) (bool, string) {
// 	if p.t.Type == t {
// 		t := p.t
// 		if p.peekToken != nil {
// 			p.t, p.peekToken = p.peekToken, nil
// 		} else {
// 			p.t = p.l.GetNextToken()
// 		}
// 		return true, t.Value
// 	}
// 	return false, ""
// }

func (p *Parse) parse() AST {
	return p.program()
}

// program : stmt_list EOF
func (p *Parse) program() AST {
	defer p.mustEat(TokenEOF)
	return p.stmtList()
}

// stmt_list : stmt | stmt Enter stmt_list
func (p *Parse) stmtList() AST {
	var list []AST
	var ast = p.stmt()
	if ast != nil {
		list = append(list, ast)
	}
	for p.t.Type == TokenEnter {
		p.mustEat(TokenEnter)
		ast = p.stmt()
		if ast != nil {
			list = append(list, ast)
		}
	}
	return ASTStmt{list: list}
}

// stmt : LBrace stmt_list RBrace
//      | IF logic LBrace stmt_list RBrace _else
//      | IF logic THEN stmt _else
//      | Function variable LParen params RParen stmt
//      | Return expr (Colon Number)?
//      | (Var)? (variable (Comma variable)* ASSIGN)+ expr (Comma expr)*
//      | expr
//      | empty
func (p *Parse) stmt() AST {
	if p.t.Type == TokenLBrace {
		p.mustEat(TokenLBrace)
		defer p.mustEat(TokenRBrace)
		return p.stmtList()
	}

	if p.t.Type == TokenIf {
		p.mustEat(TokenIf)
		logic := p.logic()
		if p.t.Type == TokenLBrace {
			p.mustEat(TokenLBrace)
			stmtList := p.stmtList()
			p.mustEat(TokenRBrace)
			return ASTBranch{
				logic: logic,
				true:  stmtList,
				false: p._else(),
			}
		}
		p.mustEat(TokenThen)
		return ASTBranch{
			logic: logic,
			true:  p.stmt(),
			false: p._else(),
		}
	}

	if p.t.Type == TokenFunction {
		p.mustEat(TokenFunction)
		name := p.variable()
		p.mustEat(TokenLParen)
		params := p.params()
		p.mustEat(TokenRParen)
		return ASTFunction{
			name:   name,
			params: params,
			stmt:   p.stmt(),
		}
	}

	if p.t.Type == TokenReturn {
		p.mustEat(TokenReturn)
		expr := p.expr()
		if p.t.Type != TokenColon {
			return ASTReturn{expr: expr}
		}
		p.mustEat(TokenColon)
		return ASTReturn{expr: expr, error: p.mustEat(TokenNumber)}
	}

	// if p.t.Type == TokenVar || p.peek() == TokenComma || p.peek() == TokenAssign {
	if p.t.Type == TokenVar || p.t.Type == TokenID {
		var isDefined bool
		if p.t.Type == TokenVar {
			p.mustEat(TokenVar)
			isDefined = true
		}
		var left []ASTVariable
		left = append(left, p.variable())
		for p.t.Type == TokenComma {
			p.mustEat(TokenComma)
			left = append(left, p.variable())
		}
		var op = p.mustEat(TokenAssign)
		var right []AST
		right = append(right, p.expr())
		for p.t.Type == TokenComma {
			p.mustEat(TokenComma)
			right = append(right, p.expr())
		}
		return ASTAssign{
			left:      left,
			op:        op,
			right:     right,
			isDefined: isDefined,
		}
	}

	return nil // ASTEmpty{}
}

// params : (ID (Comma Enter*)?)*
func (p *Parse) params() []ASTVariable {
	var list []ASTVariable
	for p.t.Type == TokenID {
		list = append(list, p.variable())
		if p.t.Type == TokenComma {
			p.mustEat(TokenComma)
			for p.t.Type == TokenEnter {
				p.mustEat(TokenEnter)
			}
		}
	}
	return list
}

// _else : ELSE stmt
//       | empty
func (p *Parse) _else() AST {
	if p.t.Type != TokenElse {
		return nil
	}
	p.mustEat(TokenElse)
	return p.stmt()
}

// variable : ID
func (p *Parse) variable() ASTVariable {
	return ASTVariable{name: p.mustEat(TokenID)}
}

// op_1 : [] () . ->
// op_2 : - * & ! ~ sizeof
// op_3 : as
// op_4 : / * %
// op_5 : + -
// op_6 : << >>
// op_7 : & ^ |
// op_8 : > >= < <= == !=
// op_9 : &&
// op_10 : ||
// op_11 : = /= *= %= += -= <<= >>= &= ^= |=
// op_12 : ,

// expr : term_plus_minus | term_plus_minus Comma expr
func (p *Parse) expr() ASTExpr {
	var list ASTExpr
	list.list = []AST{p.termPlusMinus()}
	for p.t.Type == TokenComma {
		list.list = append(list.list, p.termPlusMinus())
	}
	return list
}

// term_2 : term ((Plus | Minus) term)*
func (p *Parse) termPlusMinus() AST {
	left := p.term()
	for p.t.Type == TokenPlus || p.t.Type == TokenMinus {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.t.Type),
			right: p.term(),
		}
	}
	return left
}

// term : factor ((Mul | Div) factor)*
func (p *Parse) term() AST {
	left := p.factor()
	for p.t.Type == TokenMul || p.t.Type == TokenDiv {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.t.Type),
			right: p.factor(),
		}
	}
	return left
}

// factor : Plus factor | Minus factor | Number | LParen term_plus_minus RParen | variable
func (p *Parse) factor() AST {
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
		return p.termPlusMinus()
	default:
		return p.variable()
	}
}

// logic(logic_or_slower) : logic_and_slower | logic_and_slower Or logic
func (p *Parse) logic() AST {
	logic := p.logicAndSlower()
	if p.t.Type == TokenOrSlower {
		p.mustEat(TokenOrSlower)
		return ASTLogic{
			op:    "or",
			left:  logic,
			right: p.logic(),
		}
	}
	return logic
}

// logic_and_slower : logic_not_slower | logic_not_slower And logic_and_slower
func (p *Parse) logicAndSlower() AST {
	logic := p.logicNotSlower()
	if p.t.Type == TokenAndSlower {
		p.mustEat(TokenAndSlower)
		return ASTLogic{
			op:    "or",
			left:  logic,
			right: p.logicAndSlower(),
		}
	}
	return logic
}

// logic_not_slower : logic_or | Not logic_not_slower
func (p *Parse) logicNotSlower() AST {
	if p.t.Type == TokenNotSlower {
		p.mustEat(TokenNotSlower)
		return ASTLogic{
			op:    "not",
			right: p.logicNotSlower(),
		}
	}
	return p.logicOr()
}

// logic_or : logic_and | logic_and Or logic_or
func (p *Parse) logicOr() AST {
	logic := p.logicAnd()
	if p.t.Type == TokenOr {
		p.mustEat(TokenOr)
		return ASTLogic{
			op:    "or",
			left:  logic,
			right: p.logicOr(),
		}
	}
	return logic
}

// logic_and : logic_not | logic_not And logic_and
func (p *Parse) logicAnd() AST {
	logic := p.logicNot()
	if p.t.Type == TokenAnd {
		p.mustEat(TokenAnd)
		return ASTLogic{
			op:    "and",
			left:  logic,
			right: p.logicAnd(),
		}
	}
	return logic
}

// logic_not : compare | Not logic_not
func (p *Parse) logicNot() AST {
	if p.t.Type == TokenNot {
		p.mustEat(TokenNot)
		return ASTLogic{
			op:    "not",
			right: p.logicNot(),
		}
	}
	return p.compare()
}

// compare : expr | expr Compare expr
func (p *Parse) compare() AST {
	logic := p.expr()
	if p.t.Type == TokenCompare {
		return ASTLogic{
			op:    p.mustEat(TokenCompare),
			left:  logic,
			right: p.expr(),
		}
	}
	return logic
}
