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

// program : import stmt_list EOF
func (p *Parse) program() AST {
	defer p.mustEat(TokenEOF)
	return ASTProject{p._import(), p.stmtList()}
}

// import : (Import String Enter)*
func (p *Parse) _import() []ASTImport {
	var list []ASTImport
	for p.t.Type == TokenImport {
		p.mustEat(TokenImport)
		list = append(list, ASTImport{p.mustEat(TokenString)})
		for p.t.Type == TokenEnter {
			p.mustEat(TokenEnter)
		}
	}
	return list
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

// op_0 : [] () . ->
// op_1 : - * & ! ~ sizeof
// op_2 : as
// op_3 : / * %
// op_4 : + -
// op_5 : << >> & ^ |
// op_6 : > >= < <= == !=
// op_7 : &&
// op_8 : ||
// op_9 : = /= *= %= += -= <<= >>= &= ^= |=

// expr : op_8 (Comma op_8)*
func (p *Parse) expr() ASTExpr {
	var list ASTExpr
	list.list = []AST{p.op8()}
	for p.t.Type == TokenComma {
		list.list = append(list.list, p.op8())
	}
	return list
}

// op_8 : op_7 (Or op_7)*
func (p *Parse) op8() AST {
	var left = p.op7()
	for p.t.Type == TokenOr {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(TokenOr),
			right: p.op7(),
		}
	}
	return left
}

// op_7 : op_6 (And op_6)*
func (p *Parse) op7() AST {
	var left = p.op6()
	for p.t.Type == TokenAnd {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(TokenAnd),
			right: p.op6(),
		}
	}
	return left
}

// op_6 : op_5 (Compare op_5)*
func (p *Parse) op6() AST {
	var left = p.op5()
	for p.t.Type == TokenCompare {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(TokenCompare),
			right: p.op5(),
		}
	}
	return left
}

// op_5 : op_4 ((BitOp | OpAnd) op_4)*
func (p *Parse) op5() AST {
	var left = p.op4()
	for p.t.Type == TokenOpBit || p.t.Type == TokenOpAnd {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.t.Type),
			right: p.op4(),
		}
	}
	return left
}

// op_4 : op_3 ((Plus | Minus) op_3)*
func (p *Parse) op4() AST {
	var left = p.op3()
	for p.t.Type == TokenPlus || p.t.Type == TokenMinus {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.t.Type),
			right: p.op3(),
		}
	}
	return left
}

// op_3 : op_2 ((Mul | Div) op_2)*
func (p *Parse) op3() AST {
	var left = p.op2()
	for p.t.Type == TokenMul || p.t.Type == TokenDiv {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.t.Type),
			right: p.op2(),
		}
	}
	return left
}

// op_2 : op_1 (As op_1)*
func (p *Parse) op2() AST {
	var left = p.op1()
	for p.t.Type == TokenAs {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(TokenAs),
			right: p.op1(),
		}
	}
	return left
}

// op_1 : (Mul | Minus | OpAnd | UnaryOp)* factor
func (p *Parse) op1() AST {
	var ast AST
	var tmp = ast
	ast = tmp
	for p.t.Type == TokenMul || p.t.Type == TokenMinus || p.t.Type == TokenOpAnd || p.t.Type == TokenUnaryOp {
		ast = ASTUnaryOp{
			op:  p.mustEat(p.t.Type),
			AST: ast,
		}
	}
	tmp = p.factor()
	return ast
}

// factor : Number | LParen op_8 RParen | variable
func (p *Parse) factor() AST {
	switch p.t.Type {
	case TokenNumber:
		return ASTNumber{num: p.mustEat(TokenNumber)}
	case TokenLParen:
		p.mustEat(TokenLParen)
		defer p.mustEat(TokenRParen)
		return p.op8()
	default:
		return p.variable()
	}
}

// logic(logic_or_slower) : logic_and_slower (OrSlower logic_and_slower)*
func (p *Parse) logic() AST {
	var left = p.logicAndSlower()
	for p.t.Type == TokenOrSlower {
		left = ASTLogic{
			left:  left,
			op:    p.mustEat(TokenOrSlower),
			right: p.logicAndSlower(),
		}
	}
	return left
}

// logic_and_slower : logic_not_slower (AndSlower logic_not_slower)*
func (p *Parse) logicAndSlower() AST {
	var left = p.logicNotSlower()
	for p.t.Type == TokenAndSlower {
		left = ASTLogic{
			left:  left,
			op:    p.mustEat(TokenAndSlower),
			right: p.logicNotSlower(),
		}
	}
	return left
}

// logic_not_slower : LParen logic RParen | NotSlower expr | expr
func (p *Parse) logicNotSlower() AST {
	switch p.t.Type {
	case TokenLParen:
		p.mustEat(TokenLParen)
		defer p.mustEat(TokenRParen)
		return p.logic()
	case TokenNotSlower:
		return ASTLogic{
			op:    p.mustEat(TokenNotSlower),
			right: p.expr(),
		}
	default:
		return p.expr()
	}
}
