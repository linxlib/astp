package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func parseParam(params *ast.FieldList, tps []*types.TypeParam, imports []*types.Import, proj *types.Project) []*types.Param {
	if params == nil {
		return nil
	}
	pars := make([]*types.Param, 0)
	var pIndex int

	for _, param := range params.List {
		for _, name := range param.Names {
			par := &types.Param{
				Index:    pIndex,
				Name:     name.Name,
				ElemType: constants.ElemParam,
				Package:  new(types.Package),
			}
			info := types.NewTypePkgInfo(proj.ModPkg, "", imports)
			findPackageV2(param.Type, info)
			//slog.Info(info.FullName)
			if info.Valid {
				par.Slice = info.Slice
				par.Pointer = info.Pointer
				par.Type = info.Name
				par.Generic = info.Generic
				par.TypeName = info.FullName
				if info.PkgType == constants.PackageOtherPackage {
					par.Struct = findType(info.PkgPath, info.Name, proj.BaseDir, proj.ModPkg, proj).Clone()
					if par.Struct != nil {
						par.Package = par.Struct.Package.Clone()
					}
					par.Package.Type = info.PkgType
				} else {
					par.Package.Type = info.PkgType
					par.Package.Path = info.PkgPath
					par.Package.Name = info.PkgName
				}
				for _, tp := range tps {
					if par.Type == tp.Type {
						par.TypeParam = append(par.TypeParam, tp.CloneTiny())
					}
				}

				if par.Generic {
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
						var hasTp = false
						for _, tp := range tps {
							if par.Type == tp.Type {
								hasTp = true
								par.TypeParam = append(par.TypeParam, tp.CloneTiny())
							}
						}
						if !hasTp {
							tp.Package.Type = child.PkgType
							tp.Package.Path = child.PkgPath
							tp.Package.Name = child.PkgName
							tp.Struct = findType(child.PkgPath, child.Name, proj.BaseDir, proj.ModPkg, proj).Clone()

							par.TypeParam = append(par.TypeParam, tp)
						}

					}

				}
			}

			pars = append(pars, par)
			pIndex++
		}
	}
	return pars
}
