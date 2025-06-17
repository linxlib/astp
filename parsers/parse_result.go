package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func ParseResults(params *ast.FieldList, imports []*types.Import, proj *types.Project) []*types.Param {
	if params == nil {
		return nil
	}
	pars := make([]*types.Param, 0)
	var pIndex int
	for _, param := range params.List {
		if param.Names != nil {
			for _, name := range param.Names {
				par := &types.Param{
					Index:    pIndex,
					Name:     name.Name,
					ElemType: constants.ElemResult,
					Package:  new(types.Package),
				}

				ps := FindPackage(param.Type, imports, proj.ModPkg)
				tpString := make([]string, 0)
				for i, p := range ps {
					//TODO: 如果有多个类型(泛型)
					if i == 0 {
						par.Type = p.TypeName
						par.Struct = FindType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
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
							par.TypeName += p.PkgPath + "." + p.TypeName
						} else {
							par.TypeName += p.TypeName
						}
						//par.TypeName += p.PkgPath + "." + p.TypeName

					} else {
						//par.Generic = p.IsGeneric
						//par.Package.Type = p.PkgType
						//par.Type = p.TypeName
						if par.Generic {
							if p.PkgPath == "" {
								tpString = append(tpString, p.TypeName)
							} else {
								tpString = append(tpString, p.PkgPath+"."+p.TypeName)
							}

						}
						par.Slice = p.IsSlice
						// tParams
						//for _, tParam := range tParams {
						//	if tParam.TypeName == p.TypeName {
						//		par.Struct = tParam.Struct.Clone()
						//		par.Package.Path = tParam.Package.Path
						//		par.TypeName = tParam.TypeName
						//	}
						//}
					}
				}

				pars = append(pars, par)
				pIndex++
			}
		} else { //返回值可能为隐式参数

			par := &types.Param{
				Index:    pIndex,
				Name:     constants.EmptyName,
				ElemType: constants.ElemResult,
				Package:  new(types.Package),
			}
			ps := FindPackage(param.Type, imports, proj.ModPkg)

			// 泛型类型处理
			// 这里ps的len>1
			// eg. [0] 为BasePageResp [1] 为 models.User
			tpString := make([]string, 0)
			for i, p := range ps {
				//tmp := CheckPackage(proj.ModPkg, p.PkgPath)
				if i == 0 {
					par.Type = p.TypeName
					par.Struct = FindType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
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
					//par.TypeName += p.PkgPath + "." + p.TypeName

				} else {
					//par.Generic = p.IsGeneric
					//par.Package.Type = p.PkgType
					//par.Type = p.TypeName
					if par.Generic {
						if p.PkgPath == "" {
							tpString = append(tpString, p.TypeName)
						} else {
							tpString = append(tpString, p.PkgPath+"."+p.TypeName)
						}

					}
					par.Slice = p.IsSlice
					// tParams
					//for _, tParam := range tParams {
					//	if tParam.TypeName == p.TypeName {
					//		par.Struct = tParam.Struct.Clone()
					//		par.Package.Path = tParam.Package.Path
					//		par.TypeName = tParam.TypeName
					//	}
					//}
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
