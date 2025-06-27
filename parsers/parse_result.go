package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func parseResults(params *ast.FieldList, tps []*types.TypeParam, imports []*types.Import, proj *types.Project) []*types.Param {
	if params == nil {
		return nil
	}
	pars := make([]*types.Param, 0)
	var pIndex int
	for _, param := range params.List {
		if param.Names != nil {
			// 循环遍历 为了兼容 a,b int 类似这样的返回值
			for _, name := range param.Names {
				par := &types.Param{
					Index:    pIndex,
					Name:     name.Name,
					ElemType: constants.ElemResult,
					Package:  new(types.Package),
				}

				info := types.NewTypePkgInfo(proj.ModPkg, "", imports)
				findPackageV2(param.Type, info)
				if info.Valid {
					if info.Valid {
						par.Slice = info.Slice
						par.Pointer = info.Pointer
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
						if !par.Generic {
							for _, tp := range tps {
								if par.Type == tp.Type {
									par.TypeParam = append(par.TypeParam, tp.CloneTiny())
									break
								}
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
								for _, tp1 := range tps {
									if par.Type == tp1.Type {
										tp.Key = tp1.Key
										break
									}
								}
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
		} else { //返回值可能为隐式参数

			par := &types.Param{
				Index:    pIndex,
				Name:     constants.EmptyName,
				ElemType: constants.ElemResult,
				Package:  new(types.Package),
			}

			info := types.NewTypePkgInfo(proj.ModPkg, "", imports)
			findPackageV2(param.Type, info)
			if info.Valid {
				if info.Valid {
					par.Slice = info.Slice
					par.Pointer = info.Pointer
					par.Generic = info.Generic
					par.Type = info.Name
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
					if !par.Generic { // *E
						for _, tp := range tps {
							if par.Type == tp.Type {
								par.TypeParam = append(par.TypeParam, tp.CloneTiny())
								break
							}
						}
					}

					if par.Generic {
						for _, child := range info.Children {
							tp := &types.TypeParam{
								Type:          child.Name,
								TypeName:      child.FullName,
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

							for _, tp1 := range tps {
								if par.Type == tp1.Type {
									tp.Index = tp1.Index
									break
								}
							}

							//children 可能还有children
							if child.Children != nil {
								for _, child1 := range child.Children {
									tp1 := &types.TypeParam{
										Type:          child1.Name,
										TypeName:      child1.FullName,
										ElemType:      constants.ElemGeneric,
										Pointer:       child1.Pointer,
										Slice:         child1.Slice,
										TypeInterface: "",
										Struct:        nil,
										Package:       new(types.Package),
									}
									tp1.Package.Type = child1.PkgType
									tp1.Package.Path = child1.PkgPath
									tp1.Package.Name = child1.PkgName
									tp1.Struct = findType(child1.PkgPath, child1.Name, proj.BaseDir, proj.ModPkg, proj).Clone()

									if tp.Struct != nil {
										hasTp1 := false
										for _, tp2 := range tps {
											if tp2.Type == child1.Name {
												hasTp1 = true
												tp1.Index = tp2.Index
												tp1.Key = tp2.Key
												break
											}
										}
										if hasTp1 {
											tp.Struct.TypeParam = make([]*types.TypeParam, 0)
											tp.Struct.TypeParam = append(tp.Struct.TypeParam, tp1)
										}
									}
								}
								tp.ElemType = constants.ElemStruct
							} else {
								var hasTp = false
								for _, tp1 := range tps {
									if tp1.Type == child.Name {
										hasTp = true
										tp.Index = tp1.Index
										tp.Key = tp1.Key
										break
									}

								}
								if hasTp {
									//tp.Index = -1
								}
							}
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
