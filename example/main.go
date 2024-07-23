package main

import (
	"fmt"
	"github.com/linxlib/astp"
)

func main() {
	p := astp.NewParser()
	p.Parse()
	p.WriteOut()
	p1 := astp.NewParser()
	p1.Load()
	fmt.Println()
}
