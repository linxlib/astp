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
	flag.StringVar(&outFile, "o", "gen.gz", "-o gen.gz")
}
func main() {
	flag.Parse()
	p := &astp.Parser{}
	p.Parse()
	if outFile == "" {
		outFile = "gen.gz"
	}
	err := p.Write(outFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("complete!")
}
