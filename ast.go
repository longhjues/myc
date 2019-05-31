package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

type AST interface{}

type ASTNumber struct {
	num string
}

type ASTUnaryOp struct {
	op string
	AST
}

type ASTBinaryOp struct {
	left  AST
	op    string
	right AST
}

type ASTVariable struct {
	name string
}

type ASTStmt struct {
	list []AST
}

type ASTAssign struct {
	left      []ASTVariable
	op        string
	right     []AST
	isDefined bool
}

type ASTBranch struct {
	logic AST
	true  AST
	false AST
}

type ASTLogic struct {
	op    string
	left  AST
	right AST
}

type ASTEmpty struct{}

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
	}
	panic(ast)
}
