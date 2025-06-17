package main

import (
	"fmt"
	"github.com/linxlib/astp/parsers"
)

func main() {
	proj := parsers.ParseProj()
	err := proj.Write("gen1")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
