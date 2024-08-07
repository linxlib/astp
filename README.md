# astp
A golang ast syntax tree parser

parse your source code, and describe them with json or yaml

## Features

- 解析一个项目代码，生成可序列化内容
- 内置处理枚举写法
- 内置处理泛型和对象继承写法


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
	// 循环
    parser.VisitStruct(func(element *astp.Element) bool {
		return true
	}, func(e *astp.Element) {
		
    })
}
```

### Use it with `go generate`

```go
//go:generate go run github.com/linxlib/astp/astpg -o gen.json
```

### Examples

check [fw](github.com/linxlib/fw) and [fw_example](github.com/linxlib/fw_example) for details