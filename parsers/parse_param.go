package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func parseParam(params *ast.FieldList, imports []*types.Import, proj *types.Project) []*types.Param {
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
			ps := findPackage(param.Type, imports, proj.ModPkg)
			tpString := make([]string, 0)
			for i, p := range ps {
				if i == 0 {
					par.Type = p.TypeName
					par.Struct = findType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
					if par.Struct != nil {
						par.Package = par.Struct.Package.Clone()
					} else {
						par.Package.Type = p.PkgType
						par.Package.Path = p.PkgPath
					}

					par.Slice = p.IsSlice
					if par.Slice {
						par.TypeName += "[]"
					}
					par.Generic = p.IsGeneric

					par.Pointer = p.IsPtr
					if par.Pointer {
						par.TypeName += "*"
					}
					if par.Generic {
						if p.PkgPath == "" {
							par.TypeName += p.TypeName
						} else {
							par.TypeName += p.PkgPath + "." + p.TypeName
						}
					} else {
						par.TypeName += p.TypeName
					}

				} else {
					if par.Generic {
						if p.PkgPath == "" {
							tpString = append(tpString, p.TypeName)
						} else {
							tpString = append(tpString, p.PkgPath+"."+p.TypeName)
						}
					}
					par.Slice = p.IsSlice
				}
			}
			if len(tpString) > 0 {
				par.TypeName += "[" + strings.Join(tpString, ",") + "]"
			}
			pars = append(pars, par)
			pIndex++
		}
	}
	return pars
}
