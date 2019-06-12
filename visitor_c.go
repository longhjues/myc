package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
)

func NewExportCVisitor(ast AST, w io.Writer) *ExportCVisitor {
	return &ExportCVisitor{
		ast:    ast,
		Writer: w,
	}
}

type ExportCVisitor struct {
	ast AST
	st  *SymbolTable

	io.Writer
}

func (ev *ExportCVisitor) Exec() {
	ev.st = NewSymbolTable(nil)
	ev.exec(ev.ast)
	// ev.st.
}

func (ev *ExportCVisitor) exec(ast AST) interface{} {
	log.Println("exec:", ast)
	switch ast := ast.(type) {
	case ASTProject: // TODO: only one
		for i := range ast._import {
			fmt.Fprintf(ev, "#include\"%s\";\n", ast._import[i].path)
		}
		return ev.exec(ast.stmtList)
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
