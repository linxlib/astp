package main

import (
	"fmt"
	"github.com/linxlib/astp"
)

func main() {
	p := astp.NewParser()
	p.Parse()
	err := p.WriteOut("gen.json")
	if err != nil {
		fmt.Println(err)
	}
}
