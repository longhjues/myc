package main

import (
	"fmt"
	"io"
	"log"
	"strings"
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
			fmt.Fprintf(ev, "#include\"%s\"\n", ast._import[i].path)
			fmt.Printf("#include\"%s\"\n", ast._import[i].path)
		}
		var tmp = ev.exec(ast.stmtList)
		fmt.Fprintf(ev, "%v\n", tmp)
		fmt.Printf("%v\n", tmp)
		return nil
	case ASTNumber:
		return ast.num
	case ASTUnaryOp:
		if ast.op == "-" {
			return fmt.Sprint("-", ev.exec(ast.AST))
		}
		return fmt.Sprint(ev.exec(ast.AST))
	case ASTBinaryOp:
		return fmt.Sprintf("(%v %s %v)", ev.exec(ast.left), ast.op, ev.exec(ast.right))
	case ASTVariable:
		return fmt.Sprintf("%s", ast.name)
	case ASTStmt:
		var tmp []string
		for _, ast := range ast.list {
			tmp = append(tmp, fmt.Sprintf("%v;", ev.exec(ast)))
		}
		return strings.Join(tmp, "\n")
	case ASTFunction:
		var tmp1 []string
		for _, a := range ast._return {
			tmp1 = append(tmp1, fmt.Sprint(ev.exec(a)))
		}
		var tmp2 []string
		for _, a := range ast._return {
			tmp2 = append(tmp2, fmt.Sprint(ev.exec(a)))
		}
		return fmt.Sprintf("(%v) %s (%v) {\n%v\n}",
			strings.Join(tmp1, ","),
			ev.exec(ast.name),
			strings.Join(tmp2, ","),
			ev.exec(ast.stmt),
		)
	case ASTReturn:
		var tmp []string
		for _, ast := range ast.expr {
			tmp = append(tmp, fmt.Sprintf("%v", ev.exec(ast)))
		}
		return fmt.Sprintf("return %v/*:%s*/", strings.Join(tmp, ","), ast.error)
	case ASTCallFunc:
		var tmp []string
		for _, ast := range ast.params {
			tmp = append(tmp, fmt.Sprintf("%v", ev.exec(ast)))
		}
		return fmt.Sprintf("%v(%s)", ev.exec(ast.name), strings.Join(tmp, ","))
	case ASTString:
		return "\"" + strings.ReplaceAll(strings.ReplaceAll(ast.s, "\\", "\\\\"), "\"", "\\\"") + "\""
	case ASTAssign:
		var tmp1 []string
		var tmp2 []string
		for _,a:=range ast.left{
			tmp1=append(tmp1,fmt.Sprintf("%v",ev.exec(a)))
		}
		for _,a:=range ast.right{
			tmp2=append(tmp2,fmt.Sprintf("%v",ev.exec(a)))
		}
		var tmp3 string
		for i:=range tmp1{
			if len(tmp2)>i{
				tmp3+=tmp1[i]+ast.op+tmp2[i]+";\n"
				continue
			}
				tmp3+=tmp1[i]+";\n"
		}
		return tmp3
	case ASTBranch:
		if ast.false==nil{
			return fmt.Sprintf("if%v{\n%v\n}",ev.exec(ast.logic),ev.exec(ast.true))
		}
		return fmt.Sprintf("if%v{\n%v\n}else{\n%v\n}",ev.exec(ast.logic),ev.exec(ast.true),ev.exec(ast.false))
	case ASTLogic:
		if ast.left==nil{
			return fmt.Sprintf("(%s%v)",ast.op,ev.exec(ast.right))
		}
		return fmt.Sprintf("(%v%s%v)",ev.exec(ast.left),ast.op,ev.exec(ast.right))
	case ASTEmpty: // skip
		return ""
	}
	panic(ast)
}
