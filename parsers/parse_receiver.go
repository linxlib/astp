package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func ParseReceiver(recv *ast.FieldList, s *types.Struct, imports []*types.Import, proj *types.Project) *types.Receiver {
	receiver := recv.List[0]
	ps := FindPackage(receiver.Type, imports, proj.ModPkg)
	result := &types.Receiver{
		ElemType: constants.ElemReceiver,
		Name:     receiver.Names[0].Name,
	}
	idx1 := 0
	tpString := make([]string, 0)
	for i, p := range ps {
		if i == 0 {
			result.Generic = p.IsGeneric
			if result.Generic {
				result.TypeParam = make([]*types.TypeParam, 0)
			}
			result.Pointer = p.IsPtr
			if result.Pointer {
				result.TypeName += "*"
			}
			result.TypeName += p.TypeName
			result.Struct = s.Clone()
		} else {
			if result.Generic {
				tmp2 := CheckPackage(proj.ModPkg, p.PkgPath)
				if tmp2 == constants.PackageOtherPackage {
					tmp1 := FindType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
					genericType := tmp1.Clone()
					tp := &types.TypeParam{
						TypeName: p.TypeName,
						ElemType: constants.ElemGeneric,
						Index:    idx1,
						Type:     p.TypeName,
						Struct:   genericType,
					}
					tpString = append(tpString, p.PkgPath+"."+p.TypeName)
					result.TypeParam = append(result.TypeParam, tp)
				} else {
					result.TypeParam = append(result.TypeParam, &types.TypeParam{
						Type:     p.TypeName,
						Index:    idx1,
						TypeName: p.TypeName,
						Package: &types.Package{
							Type: p.PkgType,
							Path: p.PkgPath,
						},
					})
					if p.PkgPath == "" {
						tpString = append(tpString, p.TypeName)
					} else {
						tpString = append(tpString, p.PkgPath+"."+p.TypeName)
					}

				}
				idx1++
			}
		}
	}
	if len(tpString) > 0 {
		result.TypeName += "[" + strings.Join(tpString, ",") + "]"
	}
	return result
}
