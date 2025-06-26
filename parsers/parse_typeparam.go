package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func parseTypeParam(list *ast.FieldList, imports []*types.Import, proj *types.Project) []*types.TypeParam {
	result := make([]*types.TypeParam, 0)
	tpIndex := 0
	for _, field := range list.List {

		//TODO: 无需关心泛型的约束
		for _, name := range field.Names {
			t := new(types.TypeParam)
			t.Package = new(types.Package)
			t.Index = tpIndex
			tpIndex++
			switch spec := field.Type.(type) {
			case *ast.BinaryExpr:
				ss := parseBinaryExpr(spec)
				t.TypeInterface = strings.Join(ss, "|")
				if internal.IsInternalType(ss[0]) {
					t.Package.Type = constants.PackageBuiltin
				}
			case *ast.Ident:
				t.TypeInterface = spec.String()
				if internal.IsInternalType(t.TypeName) {
					t.Package.Type = constants.PackageBuiltin
				}
			case *ast.IndexExpr:
				ss := parseBinaryExpr(spec)
				if len(ss) == 2 {
					t.TypeInterface = ss[1]
					if internal.IsInternalGenericType(ss[1]) {
						t.Package.Type = constants.PackageBuiltin
					}
				}

			}
			t.Type = name.Name
			t.TypeName = name.Name
			t.ElemType = constants.ElemGeneric

			ps := findPackage(field.Type, imports, proj.ModPkg)
			for _, p := range ps {
				if p.PkgType != constants.PackageSamePackage && p.PkgType != constants.PackageBuiltin && p.PkgType != constants.PackageThirdPackage {
					t.Struct = findType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
					if t.Struct != nil {
						t.Package = t.Struct.Package.Clone()
					}
				} else {
					t.Package.Type = p.PkgType
					t.Slice = p.IsSlice
					t.Pointer = p.IsPtr
				}
			}

			result = append(result, t)
		}

	}
	return result
}

func parseTypeParamV2(list *ast.FieldList, imports []*types.Import, proj *types.Project) []*types.TypeParam {
	result := make([]*types.TypeParam, 0)
	idx := 0
	for _, tp := range list.List {
		for _, name := range tp.Names {
			t := new(types.TypeParam)
			t.Package = new(types.Package)
			t.Index = idx
			t.Type = name.Name
			t.ElemType = constants.ElemGeneric

			info := types.NewTypePkgInfo(proj.ModPkg, "", imports)
			findPackageV2(tp.Type, info)
			if info.Valid {
				t.Slice = info.Slice
				t.Pointer = info.Pointer
				t.TypeName = info.FullName
				if info.PkgType == constants.PackageOtherPackage {
					t.Struct = findType(info.PkgPath, info.Name, proj.BaseDir, proj.ModPkg, proj)
					if t.Struct != nil {
						t.Package = t.Struct.Package.Clone()
					}
				} else {
					t.Package.Type = info.PkgType
					t.Package.Path = info.PkgPath
					t.Package.Name = info.PkgName
				}
			}
			idx++
			result = append(result, t)
		}
	}
	return result
}
