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
				//tmp := checkPackage(proj.ModPkg, p.PkgPath)
				if i == 0 {
					//主参数的类型 (不带* [] 包名)
					par.Type = p.TypeName
					// 主参数 放到Struct
					par.Struct = findType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
					if par.Struct != nil {
						par.Package = par.Struct.Package.Clone()
					} else {
						par.Package.Type = p.PkgType
						par.Package.Path = p.PkgPath
						par.Package.Name = p.PkgName
					}

					par.Slice = p.IsSlice
					if par.Slice {
						par.TypeName += "[]"
					}

					par.Pointer = p.IsPtr
					if par.Pointer {
						par.TypeName += "*"
					}
					par.Generic = p.IsGeneric
					if par.Generic {
						if p.PkgPath == "" {
							// 不带包名 则为this
							par.TypeName += p.TypeName
						} else {
							par.TypeName += p.PkgName + "." + p.TypeName
						}
					} else {
						par.TypeName += p.TypeName
					}
					if p.PkgType == constants.PackageBuiltin && par.Generic {
						tp := &types.TypeParam{
							Type:          par.Type,
							TypeName:      par.TypeName,
							Index:         0, //表示在这个结构中的第几个
							ElemType:      constants.ElemGeneric,
							Pointer:       par.Pointer,
							Slice:         par.Slice,
							TypeInterface: par.TypeName,
							Package:       par.Package.Clone(),
							Struct:        par.Struct.Clone(),
						}
						if par.TypeParam == nil {
							par.TypeParam = make([]*types.TypeParam, 0)
						}
						par.TypeParam = append(par.TypeParam, tp)

					}

				} else {
					// 泛型也可能为 []Type 这样的形式
					if par.Generic { //一般到了这里 必为true
						tp := &types.TypeParam{
							Type:          p.TypeName,
							TypeName:      "",
							Index:         i - 1, //表示在这个结构中的第几个
							ElemType:      constants.ElemGeneric,
							Pointer:       p.IsPtr,
							Slice:         false,
							TypeInterface: p.TypeName,
							Package:       new(types.Package),
						}
						if par.TypeParam == nil {
							par.TypeParam = make([]*types.TypeParam, 0)
						}
						tp.Pointer = p.IsPtr
						tp.Slice = p.IsSlice
						if tp.Slice {
							tp.TypeName += "[]"
						}
						if tp.Pointer {
							tp.TypeName += "*"
						}
						if p.PkgPath == "" { //this 包
							tp.TypeName += p.TypeName
						} else {
							tp.TypeName += p.PkgName + "." + p.TypeName
						}
						// 暂时先不考虑 *[]*Type 这样的复杂情况, 仅考虑 []*Type
						tp.Struct = findType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
						if tp.Struct != nil {
							tp.Package = par.Struct.Package.Clone()
						} else {
							tp.Package.Type = p.PkgType
							tp.Package.Path = p.PkgPath
							tp.Package.Name = p.PkgName
						}
						//tp.TypeName += p.TypeName
						tpString = append(tpString, tp.TypeName)
						par.TypeParam = append(par.TypeParam, tp)

					}

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
