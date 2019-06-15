package main

import "fmt"

func NewParse(tokens []*Token) *Parse {
	return &Parse{token: tokens}
}

type Parse struct {
	token []*Token
	pos   int
}

func (p *Parse) peek() TokenType {
	if len(p.token) <= p.pos+1 {
		return TokenEOF
	}
	return p.token[p.pos+1].Type
}

func (p *Parse) mustEat(t TokenType) string {
	fmt.Println(t, len(p.token), p.pos, p.token[p.pos])
	if p.token[p.pos].Type == t {
		if len(p.token) <= p.pos+1 {
			fmt.Println("-------EOF--------")
			return "EOF"
		}
		p.pos++
		return p.token[p.pos-1].Value
	}
	panic(p.token[p.pos].Type)
}

// func (p *Parse) eat(t TokenType) (bool, string) {
// 	if p.token[p.pos].Type == t {
// 		t := p.token[p.pos]
// 		if p.peekToken != nil {
// 			p.token[p.pos], p.peekToken = p.peekToken, nil
// 		} else {
// 			p.token[p.pos] = p.l.GetNextToken()
// 		}
// 		return true, t.Value
// 	}
// 	return false, ""
// }

func (p *Parse) parse() AST {
	return p.program()
}

// program : Enter* import stmt_list EOF
func (p *Parse) program() AST {
	for p.token[p.pos].Type == TokenEnter {
		p.mustEat(TokenEnter)
	}
	defer p.mustEat(TokenEOF)
	return ASTProject{p._import(), p.stmtList()}
}

// import : (Import String Enter)*
func (p *Parse) _import() []ASTImport {
	var list []ASTImport
	for p.token[p.pos].Type == TokenImport {
		p.mustEat(TokenImport)
		list = append(list, ASTImport{p.mustEat(TokenString)})
		for p.token[p.pos].Type == TokenEnter {
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
	for p.token[p.pos].Type == TokenEnter {
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
//      | function(Function...)
//      | Return expr (Comma Enter? expr)* (Colon Number)?
//      | (Var)? variable (Comma variable)* ASSIGN expr (Comma expr)*
//      | expr
//      | empty
func (p *Parse) stmt() AST {
	if p.token[p.pos].Type == TokenLBrace {
		p.mustEat(TokenLBrace)
		defer p.mustEat(TokenRBrace)
		return p.stmtList()
	}

	if p.token[p.pos].Type == TokenIf {
		p.mustEat(TokenIf)
		logic := p.logic()
		if p.token[p.pos].Type == TokenLBrace {
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

	if p.token[p.pos].Type == TokenFunction {
		return p.function()
	}

	if p.token[p.pos].Type == TokenReturn {
		p.mustEat(TokenReturn)
		var exprs []AST
		exprs=append(exprs,p.expr())
		for p.token[p.pos].Type==TokenComma{
			for p.token[p.pos].Type==TokenReturn{
				p.mustEat(TokenReturn)
			}
			exprs=append(exprs,p.expr())
		}
		if p.token[p.pos].Type != TokenColon {
			return ASTReturn{expr: exprs}
		}
		p.mustEat(TokenColon)
		return ASTReturn{expr: exprs, error: p.mustEat(TokenNumber)}
	}

	if p.token[p.pos].Type == TokenID && (p.peek() != TokenComma && p.peek() != TokenAssign) { // expr
		return p.expr()
	}

	if p.token[p.pos].Type == TokenVar || p.token[p.pos].Type == TokenID {
		var isDefined bool
		if p.token[p.pos].Type == TokenVar {
			p.mustEat(TokenVar)
			isDefined = true
		}
		var left []ASTVariable
		left = append(left, p.variable())
		for p.token[p.pos].Type == TokenComma {
			p.mustEat(TokenComma)
			left = append(left, p.variable())
		}
		var op = p.mustEat(TokenAssign)
		var right []AST
		right = append(right, p.expr())
		for p.token[p.pos].Type == TokenComma {
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

	return ASTEmpty{}
}

// function : Function variable def_params def_params? LBrace stmt_list RBrace
func (p *Parse) function() ASTFunction {
	p.mustEat(TokenFunction)
	name := p.variable()
	params := p.defParams()
	var _return []ASTVariable
	if p.token[p.pos].Type == TokenLParen {
		_return = p.defParams()
	}
	p.mustEat(TokenLBrace)
	var ast= ASTFunction{
		name:    name,
		params:  params,
		stmt:    p.stmtList(),
		_return: _return,
	}
	p.mustEat(TokenRBrace)
	return ast
}

// def_params : LParen (ID type (Comma Enter*)?)* RParen
func (p *Parse) defParams() []ASTVariable {
	p.mustEat(TokenLParen)
	var list []ASTVariable
	for p.token[p.pos].Type == TokenID {
		list = append(list, p.variable())
		if p.token[p.pos].Type == TokenComma {
			p.mustEat(TokenComma)
			for p.token[p.pos].Type == TokenEnter {
				p.mustEat(TokenEnter)
			}
		}
	}
	p.mustEat(TokenRParen)
	return list
}

// params : LParen (expr (Comma Enter*)?)* RParen
func (p *Parse) params() []AST {
	p.mustEat(TokenLParen)
	var list []AST
	for p.token[p.pos].Type != TokenRParen {
		list = append(list, p.expr())
		if p.token[p.pos].Type == TokenComma {
			p.mustEat(TokenComma)
			for p.token[p.pos].Type == TokenEnter {
				p.mustEat(TokenEnter)
			}
		}
	}
	p.mustEat(TokenRParen)
	return list
}

// _else : ELSE stmt
//       | empty
func (p *Parse) _else() AST {
	if p.token[p.pos].Type != TokenElse {
		return nil
	}
	p.mustEat(TokenElse)
	return p.stmt()
}

// variable : ID (Dot ID)*
func (p *Parse) variable() ASTVariable {
	var name = p.mustEat(TokenID)
	for p.token[p.pos].Type == TokenDot {
		name += p.mustEat(TokenDot)
		name += p.mustEat(TokenID)
	}
	return ASTVariable{name: name}
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

// expr : op_8
func (p *Parse) expr() AST {
	return p.op8()
}

// op_8 : op_7 (Or op_7)*
func (p *Parse) op8() AST {
	var left = p.op7()
	for p.token[p.pos].Type == TokenOr {
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
	for p.token[p.pos].Type == TokenAnd {
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
	for p.token[p.pos].Type == TokenCompare {
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
	for p.token[p.pos].Type == TokenOpBit || p.token[p.pos].Type == TokenOpAnd {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.token[p.pos].Type),
			right: p.op4(),
		}
	}
	return left
}

// op_4 : op_3 ((Plus | Minus) op_3)*
func (p *Parse) op4() AST {
	var left = p.op3()
	for p.token[p.pos].Type == TokenPlus || p.token[p.pos].Type == TokenMinus {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.token[p.pos].Type),
			right: p.op3(),
		}
	}
	return left
}

// op_3 : op_2 ((Mul | Div) op_2)*
func (p *Parse) op3() AST {
	var left = p.op2()
	for p.token[p.pos].Type == TokenMul || p.token[p.pos].Type == TokenDiv {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(p.token[p.pos].Type),
			right: p.op2(),
		}
	}
	return left
}

// op_2 : op_1 (As op_1)*
func (p *Parse) op2() AST {
	var left = p.op1()
	for p.token[p.pos].Type == TokenAs {
		left = ASTBinaryOp{
			left:  left,
			op:    p.mustEat(TokenAs),
			right: p.op1(),
		}
	}
	return left
}

// op_1 : (Mul | Minus | OpAnd | UnaryOp) op_1 | factor
func (p *Parse) op1() AST {
	var t = p.token[p.pos].Type
	if t == TokenMul || t == TokenMinus || t == TokenOpAnd || t == TokenUnaryOp {
		return ASTUnaryOp{
			op:  p.mustEat(t),
			AST: p.op1(),
		}
	}
	return p.factor()
}

// factor : Number | String | LParen op_8 RParen | variable params?
func (p *Parse) factor() AST {
	switch p.token[p.pos].Type {
	case TokenNumber:
		return ASTNumber{num: p.mustEat(TokenNumber)}
	case TokenString:
		return ASTString{s: p.mustEat(TokenString)}
	case TokenLParen:
		p.mustEat(TokenLParen)
		defer p.mustEat(TokenRParen)
		return p.op8()
	default:
		var tmp = p.variable()
		if p.token[p.pos].Type == TokenLParen {
			return ASTCallFunc{
				name:   tmp,
				params: p.params(),
			}
		}
		return tmp
	}
}

// logic(logic_or_slower) : logic_and_slower (OrSlower logic_and_slower)*
func (p *Parse) logic() AST {
	var left = p.logicAndSlower()
	for p.token[p.pos].Type == TokenOrSlower {
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
	for p.token[p.pos].Type == TokenAndSlower {
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
	switch p.token[p.pos].Type {
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
