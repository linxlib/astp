package main

import (
	"fmt"
	"github.com/linxlib/astp"
)

func main() {
	parser := &astp.Parser{}
	err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = parser.Write("gen.gz")
	if err != nil {
		fmt.Println(err)
		return
	}
}
