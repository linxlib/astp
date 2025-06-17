package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"go/token"
)

func ParseVar(af *ast.File, proj *types.Project, imports []*types.Import) []*types.Variable {
	result := make([]*types.Variable, 0)

	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.VAR:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &types.Variable{
								Name:     v.Name,
								ElemType: constants.ElemVar,
								Value:    nil,
								TypeName: "",
							}

							if len(spec.Values) == len(spec.Names) {
								if a, ok := spec.Values[i].(*ast.BasicLit); ok {
									vv.Value = a.Value
								}
							}
							ps := FindPackage(spec.Type, imports, proj.ModPkg)
							for _, p := range ps {
								if p.PkgType != constants.PackageSamePackage && p.PkgType != constants.PackageBuiltin && p.PkgType != constants.PackageThirdPackage {
									vv.Struct = FindType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
									vv.TypeName = vv.Struct.Name
								}
							}
							result = append(result, vv)
						}
					}
				}
			default:
				continue
			}
		}
	}
	return result
}
