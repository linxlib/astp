package parsers

import (
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

func ParseFile(file string, proj *types.Project) *types.File {
	name := filepath.Base(file)
	node, _ := parser.ParseFile(token.NewFileSet(), file, nil, parser.ParseComments)

	p := ParsePackage(node, file, proj)

	doc := HandleDoc(node.Doc, p.Name)
	i := ParseImport(node)
	v := ParseVar(node, proj, i)

	v1 := ParseConst(node, p)

	//f1 := ParseFunction(node, p, i, proj)
	//f2 := ParseInterface(node,  i, proj)

	s := ParseStruct(node, p, i, proj)
	f := &types.File{
		Key:       internal.GetKey(p.Path, name),
		KeyHash:   internal.GetKeyHash(p.Path, name),
		Name:      name,
		Package:   p,
		Comment:   doc,
		Import:    i,
		Variable:  v,
		Const:     v1,
		Function:  nil,
		Interface: nil,
		Struct:    s,
	}
	proj.AddFile(f)
	if f.IsMainPackage() {
		for _, i1 := range f.Import {
			if strings.HasPrefix(i1.Path, f.Package.Path) {
				dir := GetPackageDir(i1.Path, proj.BaseDir, proj.ModPkg)
				files := ParseDir(dir, proj)
				proj.Merge(files)

			}
		}
	}

	return f
}
