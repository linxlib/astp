package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func ParseTypeParam(list *ast.FieldList, imports []*types.Import, proj *types.Project) []*types.TypeParam {
	result := make([]*types.TypeParam, 0)
	tpIndex := 0
	for _, field := range list.List {

		for _, name := range field.Names {
			t := new(types.TypeParam)
			t.Package = new(types.Package)
			t.Index = tpIndex
			tpIndex++
			switch spec := field.Type.(type) {
			case *ast.BinaryExpr:
				ss := ParseBinaryExpr(spec)
				t.TypeInterface = strings.Join(ss, "|")
				if internal.IsInternalType(ss[0]) {
					t.Package.Type = constants.PackageBuiltin
				}
			case *ast.Ident:
				t.TypeInterface = spec.String()
				if internal.IsInternalType(t.TypeName) {
					t.Package.Type = constants.PackageBuiltin
				}
			}
			t.Type = name.Name
			t.TypeName = name.Name
			t.ElemType = constants.ElemGeneric
			ps := FindPackage(field.Type, imports, proj.ModPkg)
			for _, p := range ps {
				if p.PkgType != constants.PackageSamePackage && p.PkgType != constants.PackageBuiltin && p.PkgType != constants.PackageThirdPackage {

					t.Struct = FindType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
					if t.Struct != nil {
						t.Package = t.Struct.Package.Clone()
					}

				} else {
					t.Package.Type = p.PkgType
				}
			}

			result = append(result, t)
		}

	}
	return result
}
