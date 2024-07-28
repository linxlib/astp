package main

import (
	"flag"
	"fmt"
	"github.com/linxlib/astp"
)

var (
	outFile string
)

func init() {
	flag.StringVar(&outFile, "o", "default", "-o gen.json")
}
func main() {
	flag.Parse()
	p := astp.NewParser()
	p.Parse()
	if outFile == "" {
		outFile = "gen.json"
	}
	err := p.WriteOut(outFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("complete!")
}
