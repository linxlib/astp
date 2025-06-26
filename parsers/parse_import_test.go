package parsers

import (
	"github.com/linxlib/astp/types"
	"go/parser"
	"go/token"
	"testing"
)

func Test_parseImport(t *testing.T) {
	node, _ := parser.ParseFile(token.NewFileSet(), "./tests/for_imports.txt", nil, parser.ParseComments)
	proj := &types.Project{
		BaseDir: "./tests",
		ModPkg:  "tests",
	}
	pkg := parsePackage(node, "./tests/for_package.go", proj)
	if pkg.Name != "tests" {
		t.FailNow()
	}

	imports := parseImport(node)
	if len(imports) != 3 {
		t.Fail()
	}
	if !imports[0].Ignore || imports[0].Alias != "_" {
		t.Fail()
	}
	if imports[1].Alias != "aaa" {
		t.Fail()
	}
	if imports[2].Name != "constants" {
		t.Fail()
	}

}
