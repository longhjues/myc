package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

type AST interface{}

type ASTProject struct {
	_import  []ASTImport
	stmtList AST
}

func (ast ASTProject) String() string {
	return fmt.Sprintf("(imports %v,program %v)", ast._import, ast.stmtList)
}

type ASTImport struct {
	path string
}

func (ast ASTImport) String() string {
	return fmt.Sprintf("(import %v)", ast.path)
}

type ASTNumber struct {
	num string
}

func (ast ASTNumber) String() string {
	return fmt.Sprintf("[N:%v]", ast.num)
}

type ASTString struct {
	s string
}

func (ast ASTString) String() string {
	return fmt.Sprintf("[S:%v]", ast.s)
}

type ASTUnaryOp struct {
	op string
	AST
}

func (ast ASTUnaryOp) String() string {
	return fmt.Sprintf("(op %v %v)", ast.op, ast.AST)
}

type ASTBinaryOp struct {
	left  AST
	op    string
	right AST
}

func (ast ASTBinaryOp) String() string {
	return fmt.Sprintf("(op %v %v %v)", ast.left, ast.op, ast.right)
}

type ASTVariable struct {
	name string
	ty   string // type
}

func (ast ASTVariable) String() string {
	return fmt.Sprintf("[V:%v:%v]", ast.name, ast.ty)
}

// type ASTType struct{}

type ASTStmt struct {
	list []AST
}

func (ast ASTStmt) String() string {
	return fmt.Sprintf("(stmt %v)", ast.list)
}

type ASTAssign struct {
	left      []ASTVariable
	op        string
	right     []AST
	isDefined bool
}

func (ast ASTAssign) String() string {
	return fmt.Sprintf("(assign %v %v(%v) %v)", ast.left, ast.op, ast.isDefined, ast.right)
}

type ASTBranch struct {
	logic AST
	true  AST
	false AST
}

func (ast ASTBranch) String() string {
	return fmt.Sprintf("(branch %v %v %v)", ast.logic, ast.true, ast.false)
}

type ASTLogic struct {
	op    string
	left  AST
	right AST
}

func (ast ASTLogic) String() string {
	return fmt.Sprintf("(logic %s %v %v)", ast.op, ast.left, ast.right)
}

type ASTFunction struct {
	name    ASTVariable
	params  []ASTVariable
	_return []ASTVariable
	stmt    AST
}

func (ast ASTFunction) String() string {
	return fmt.Sprintf("(def_func %v (%v) (%v) %v)", ast.name, ast.params, ast._return, ast.stmt)
}

type ASTCallFunc struct {
	name   ASTVariable
	params []AST
}

func (ast ASTCallFunc) String() string {
	return fmt.Sprintf("(call_func %v (%v))", ast.name, ast.params)
}

type ASTReturn struct {
	expr  []AST
	error string
}

func (ast ASTReturn) String() string {
	return fmt.Sprintf("(return %v %s)", ast.expr, ast.error)
}

type ASTEmpty struct{}

func (ast ASTEmpty) String() string {
	return "(VOID)"
}

type Symbol struct {
	name     string
	value    string
	varValue int
	t        string // type
}

func (s *Symbol) String() string {
	return fmt.Sprintf("(%s:%v:%v)", s.name, s.value, s.varValue)
}

func NewSymbolTable(prev *SymbolTable) *SymbolTable {
	return &SymbolTable{
		prev: prev,
		t:    make(map[string]*Symbol),
	}
}

type SymbolTable struct {
	prev *SymbolTable
	t    map[string]*Symbol
}

func (st *SymbolTable) String() string {
	return fmt.Sprintf("{%v\n%v}", st.t, st.prev)
}

func (st *SymbolTable) Get(name string) *Symbol {
	if s, ok := st.t[name]; ok {
		return s
	}
	if st.prev == nil {
		return nil
	}
	return st.prev.Get(name)
}

func (st *SymbolTable) SetVar(name string, value int) {
	if _, ok := st.t[name]; ok {
		if err := st.set(name, "var", value); err != nil {
			panic(err)
		}
		return
	}
	if st.prev == nil {
		panic(name)
	}
	st.prev.SetVar(name, value)
}

func (st *SymbolTable) DefinedVar(name string, value int) {
	if s, ok := st.t[name]; ok {
		panic(s)
	}
	st.set(name, "var", value)
}

func (st *SymbolTable) set(name, t string, value int) error {
	if s, ok := st.t[name]; ok {
		if s.t != t {
			return errors.New("type is not much")
		}
		s.t = t
		return nil
	}
	st.t[name] = &Symbol{
		name:     name,
		t:        t,
		varValue: value,
	}
	return nil
}

func (st *SymbolTable) DefinedOrSetVar(name string, value int) {
	s := st.Get(name)
	if s == nil {
		st.set(name, "var", value)
		return
	}
	if s.t != "var" {
		panic(s.t)
	}
	s.varValue = value
}

func NewExecVisitor(ast AST) *ExecVisitor {
	return &ExecVisitor{
		ast: ast,
	}
}

type ExecVisitor struct {
	ast AST
	st  *SymbolTable
}

func (ev *ExecVisitor) Exec() {
	ev.st = NewSymbolTable(nil)
	ev.exec(ev.ast)
	// ev.st.
}

func (ev *ExecVisitor) exec(ast AST) interface{} {
	log.Println("exec:", ast)
	switch ast := ast.(type) {
	case ASTNumber:
		tmp, err := strconv.Atoi(ast.num)
		if err != nil {
			panic(err)
		}
		return tmp
	case ASTUnaryOp:
		if ast.op == "-" {
			return -ev.exec(ast.AST).(int)
		}
		return ev.exec(ast.AST)
	case ASTBinaryOp:
		left := ev.exec(ast.left).(int)
		right := ev.exec(ast.right).(int)
		switch ast.op {
		case "+":
			return left + right
		case "-":
			return left - right
		case "*":
			return left * right
		case "/":
			return left / right
		default:
			panic(ast.op)
		}
	case ASTVariable:
		tmp := ev.st.Get(ast.name)
		if tmp == nil || tmp.t != "var" {
			panic(ast.name)
		}
		return tmp.varValue
	case ASTStmt:
		for _, ast := range ast.list {
			ev.exec(ast)
		}
		return nil
	case ASTAssign:
		if ast.isDefined && ast.op != "=" {
			panic(ast.op)
		}
		var right []int
		for _, ast := range ast.right {
			right = append(right, ev.exec(ast).(int))
		}
		if len(right) == 1 { // exp. var a,b,c=1
			for i := range ast.left {
				if ast.isDefined {
					ev.st.DefinedVar(ast.left[i].name, right[0])
				} else if ast.op == "=" {
					ev.st.SetVar(ast.left[i].name, right[0])
				} else {
					s := ev.st.Get(ast.left[i].name)
					switch ast.op {
					case "+=":
						s.varValue += right[0]
					case "-=":
						s.varValue -= right[0]
					case "*=":
						s.varValue *= right[0]
					case "/=":
						s.varValue /= right[0]
					default:
						panic(ast.op)
					}
				}
			}
			return right
		}
		if len(ast.left) != len(right) {
			panic(len(ast.left) - len(right))
		}
		for i := range ast.left {
			if ast.isDefined {
				ev.st.DefinedVar(ast.left[i].name, right[i])
			} else if ast.op == "=" {
				ev.st.SetVar(ast.left[i].name, right[i])
			} else {
				s := ev.st.Get(ast.left[i].name)
				switch ast.op {
				case "+=":
					s.varValue += right[i]
				case "-=":
					s.varValue -= right[i]
				case "*=":
					s.varValue *= right[i]
				case "/=":
					s.varValue /= right[i]
				default:
					panic(ast.op)
				}
			}
		}
		return right
	case ASTLogic:
		switch ast.op {
		case "and":
			left := ev.exec(ast.left).(int)
			if left == 0 {
				return 0
			}
			return ev.exec(ast.right)
		case "or":
			left := ev.exec(ast.left).(int)
			if left != 0 {
				return left
			}
			return ev.exec(ast.right)
		case "not":
			right := ev.exec(ast.right).(int)
			if right == 0 {
				return 1
			}
			return 0
		case "<":
			if ev.exec(ast.left).(int) < ev.exec(ast.right).(int) {
				return 1
			}
			return 0
		case "<=":
			if ev.exec(ast.left).(int) <= ev.exec(ast.right).(int) {
				return 1
			}
			return 0
		case "==":
			if ev.exec(ast.left).(int) == ev.exec(ast.right).(int) {
				return 1
			}
			return 0
		case "!=":
			if ev.exec(ast.left).(int) != ev.exec(ast.right).(int) {
				return 1
			}
			return 0
		case ">":
			if ev.exec(ast.left).(int) > ev.exec(ast.right).(int) {
				return 1
			}
			return 0
		case ">=":
			if ev.exec(ast.left).(int) >= ev.exec(ast.right).(int) {
				return 1
			}
			return 0
		}
	case ASTEmpty:
		return nil
	case ASTBranch:
		if ev.exec(ast.logic).(int) == 0 {
			return ev.exec(ast.false)
		} else {
			return ev.exec(ast.true)
		}
	case ASTFunction: // skip
		return nil
	}
	panic(ast)
}
