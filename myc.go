package main

import (
	"io/ioutil"
	"log"
	"os"
)

func main() {
	b, err := ioutil.ReadFile("myc/test01.myc")
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
	f, err := os.Create("myc/test01.c")
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
