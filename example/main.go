package main

import (
	"fmt"
	"github.com/linxlib/astp"
	"github.com/linxlib/astp/example/testpackage"
)

func main() {
	testpackage.TestMethod2()
	p := astp.NewParser()
	p.Parse("./testpackage/example2.go")
	err := p.WriteOut("gen.json")
	if err != nil {
		fmt.Println(err)
	}
}
