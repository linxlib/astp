package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func parseReceiver(recv *ast.FieldList, s *types.Struct, imports []*types.Import, proj *types.Project) *types.Receiver {
	receiver := recv.List[0]

	info := types.NewTypePkgInfo(proj.ModPkg, s.Package.Path, imports)
	findPackageV2(receiver.Type, info)
	if info.Name != s.Type {
		return nil
	}

	result := &types.Receiver{
		ElemType: constants.ElemReceiver,
		Name:     receiver.Names[0].Name,
	}
	result.Struct = s.Clone()
	result.ElemType = constants.ElemReceiver
	result.TypeName = info.FullName
	result.Type = info.Name
	result.Generic = info.Generic
	if result.Generic {
		for idx, child := range info.Children {
			tp := &types.TypeParam{
				Type:          child.Name,
				TypeName:      child.FullName,
				Index:         idx,
				ElemType:      constants.ElemGeneric,
				Pointer:       child.Pointer,
				Slice:         child.Slice,
				TypeInterface: "",
				Struct:        nil,
				Package:       new(types.Package),
			}
			tp.Package.Type = child.PkgType
			tp.Package.Path = child.PkgPath
			tp.Package.Name = child.PkgName
			tp.Struct = findType(child.PkgPath, child.Name, proj.BaseDir, proj.ModPkg, proj).Clone()
			result.TypeParam = append(result.TypeParam, tp)
		}

	}

	//ps := parseTypeParamV2(recv, imports, proj)
	//result.TypeParam = append(result.TypeParam, ps...)
	return result
}
