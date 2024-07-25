package main

import (
	"fmt"
	"github.com/linxlib/astp"
)

func main() {
	p := astp.NewParser()
	p.Parse()
	err := p.WriteOut()
	if err != nil {
		fmt.Println(err)
	}
	//p1 := astp.NewParser()
	//p1.Load()
	//fmt.Println()
}
