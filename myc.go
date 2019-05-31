package main

import (
	"io/ioutil"
	"log"
)

func main() {
	b, err := ioutil.ReadFile("myc/test.myc")
	if err != nil {
		panic(err)
	}
	l := NewLexer(b)
	// var t = l.GetNextToken()
	// fmt.Println(t)
	// for t != nil {
	// 	t = l.GetNextToken()
	// 	fmt.Println(t)
	// }
	// return
	p := NewParse(l)
	ev := NewExecVisitor(p.parse())
	ev.Exec()
	log.Printf("%v", ev.st)
}
