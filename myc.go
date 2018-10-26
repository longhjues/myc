package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	b, err := ioutil.ReadFile("test.myc")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	tokens := analysis([]rune(string(b)))
	fmt.Println(tokens)
}
