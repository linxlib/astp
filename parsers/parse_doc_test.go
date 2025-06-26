package parsers

import (
	"github.com/linxlib/astp/types"
	"go/parser"
	"go/token"
	"testing"
)

func Test_parseDoc(t *testing.T) {
	node, _ := parser.ParseFile(token.NewFileSet(), "./tests/for_package.go", nil, parser.ParseComments)
	proj := &types.Project{
		BaseDir: "./tests",
		ModPkg:  "tests",
	}
	pkg := parsePackage(node, "./tests/for_package.go", proj)
	if pkg.Name != "tests" {
		t.FailNow()
	}
	list := parseDoc(node.Doc, "Package "+pkg.Name+" ")
	if len(list) != 1 {
		t.FailNow()
	}
	if list[0].AttrValue != "here are some docs for package" {
		t.FailNow()
	}

}
