package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	b, err := ioutil.ReadFile("test01.myc")
	if err != nil {
		panic(err)
	}
	l := NewLexer(b)
	tokens := l.LexerToken()
	fmt.Println("lexer success:", tokens)
	// return
	p := NewParse(tokens)
	f, err := os.Create("tmp/test.go")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	ev := NewExportGoVisitor(p.parse(), f)
	// ev := NewExecVisitor(p.parse())
	ev.Exec()
	log.Printf("%v", ev.st)
}
