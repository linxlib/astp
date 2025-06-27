package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"go/token"
)

func parseVar(af *ast.File, proj *types.Project, imports []*types.Import) []*types.Variable {
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
							}
							if len(spec.Values) == len(spec.Names) {
								if a, ok := spec.Values[i].(*ast.BasicLit); ok {
									vv.Value = a.Value
								}
							}
							info := types.NewTypePkgInfo(proj.ModPkg, "", imports)
							findPackageV2(spec.Type, info)
							if info.Valid {
								vv.Type = info.Name
								vv.TypeName = info.FullName
								if info.PkgType == constants.PackageOtherPackage {
									vv.Struct = findType(info.PkgPath, info.Name, proj.BaseDir, proj.ModPkg, proj)
									if vv.Struct != nil {
										vv.Package = vv.Struct.Package.Clone()
									}
									vv.Package.Type = info.PkgType
								} else {
									vv.Package.Type = info.PkgType
									vv.Package.Path = info.PkgPath
									vv.Package.Name = info.PkgName
								}
							}
							result = append(result, vv)
						}
					}
				}
			default:

			}
		}
	}
	return result
}
