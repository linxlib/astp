package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"

	"go/ast"
)

func parseField(fields []*ast.Field, tps []*types.TypeParam, imports []*types.Import, proj *types.Project, belongToStruct *types.Struct) []*types.Field {
	var sf = make([]*types.Field, 0)
	for idx, field := range fields {
		af1 := new(types.Field)
		af1.Index = idx

		af1.Package = new(types.Package)
		if field.Names != nil {
			af1.Name = field.Names[0].Name
			af1.Private = internal.IsPrivate(af1.Name)
		} else {
			af1.Name = constants.EmptyName
			af1.Parent = true
			af1.Private = false
		}
		af1.Comment = parseDoc(field.Comment, af1.Name)
		af1.Doc = parseDoc(field.Doc, af1.Name)
		if field.Tag != nil {
			af1.Tag = field.Tag.Value
		}

		// 对于某个字段, 查找其类型的包.
		// 包含该类型结构的包, 类型中泛型类型所在的包等等
		info := types.NewTypePkgInfo(proj.ModPkg, "", imports)
		findPackageV2(field.Type, info)
		if info.Valid {
			af1.Type = info.Name
			af1.Slice = info.Slice
			af1.Pointer = info.Pointer
			af1.Generic = info.Generic
			af1.Struct = findType(info.PkgPath, info.Name, proj.BaseDir, proj.ModPkg, proj)
			if af1.Struct != nil {
				af1.Package = af1.Struct.Package.Clone()
			}

			af1.Package.Type = info.PkgType
			af1.TypeName = info.FullName
			if af1.Generic {
				if info.Children != nil {
					for idx2, child := range info.Children {
						tp := &types.TypeParam{
							Type:          child.Name,
							TypeName:      child.FullName,
							Index:         idx2,
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
						tp.Struct = findType(child.PkgPath, child.Name, proj.BaseDir, proj.ModPkg, proj)

						af1.TypeParam = append(af1.TypeParam, tp)
					}
				} else {
					idx5 := 0
					if len(tps) > 0 {
						for _, tp := range tps {
							if tp.Type == info.Name {
								idx5 = tp.Index
								break
							}
						}
					}
					tmp := &types.TypeParam{
						Type:          info.Name,
						TypeName:      info.FullName,
						Index:         idx5,
						ElemType:      constants.ElemGeneric,
						Pointer:       info.Pointer,
						Slice:         info.Slice,
						TypeInterface: "",
						Struct:        nil,
						Package:       new(types.Package),
					}
					tmp.Package.Type = info.PkgType
					af1.TypeParam = append(af1.TypeParam, tmp)
				}

			}
		}

		if af1.Name == "" || af1.Name == "_" {
			af1.Parent = true

		}

		if field.Tag != nil {
			af1.Tag = field.Tag.Value
		}

		sf = append(sf, af1)
	}
	return sf
}
