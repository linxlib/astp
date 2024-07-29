# astp
A golang ast syntax tree parser


## TODO List

- [ ] store files group by package 
- [ ] 解析各种奇奇怪怪的类型
- [x] generic type support
- [x] go generate support
- [x] Struct 继承，合并字段和方法


## Usage

### Common Use
```go
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
```

### Load

```go
package main

import (
	"github.com/linxlib/astp"
)
func main() {
    parser := astp.NewParser()
    parser.Load("gen.json")
    parser.VisitAllStructs("<Struct Name>", func(s *astp.Struct) bool {
		return false
	})
}
```

### Use it with `go generate`

```go
//go:generate go run github.com/linxlib/astp/astpg -o gen.json
```

### Examples

check [fw_example](github.com/linxlib/fw_example)