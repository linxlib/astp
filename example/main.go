package main

import (
	"fmt"
	"github.com/linxlib/astp"
)

func main() {
	p := astp.NewParser().SetParseFunctions(false).AddIgnorePkg("github.com/linxlib/godeploy/middlewares").AddIgnorePkg("github.com/linxlib/godeploy/middlewares/session")
	p.Parse()
	err := p.WriteOut("gen.json")
	if err != nil {
		fmt.Println(err)
	}
}
