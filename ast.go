package main

import (
	"errors"
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

type Symbol struct {
	name     string
	value    string
	varValue int
	t        string // type
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
	// ev.st.
}

func (ev *ExecVisitor) exec(ast AST) interface{} {
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
		for ast := range ast.list {
			ev.exec(ast)
		}
		return nil
	case ASTAssign:
	}
	panic(ast)
}
