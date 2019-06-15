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
	f, err := os.Create("_test01.c")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	ev := NewExportCVisitor(p.parse(), f)
	// ev := NewExecVisitor(p.parse())
	ev.Exec()
	log.Printf("%v", ev.st)
}
