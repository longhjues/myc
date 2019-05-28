package main

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
	left  []ASTVariable
	op    string
	right []AST
}
