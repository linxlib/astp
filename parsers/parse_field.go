package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func ParseField(fields []*ast.Field, imports []*types.Import, proj *types.Project) []*types.Field {
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
		}
		af1.Comment = HandleDoc(field.Comment, af1.Name)
		af1.Doc = HandleDoc(field.Doc, af1.Name)
		if field.Tag != nil {
			af1.Tag = field.Tag.Value
		}

		// 对于某个字段, 查找其类型的包.
		// 包含该类型结构的包, 类型中泛型类型所在的包等等
		ps := FindPackage(field.Type, imports, proj.ModPkg)
		idx1 := 0
		tpString := make([]string, 0)
		for idx2, p := range ps {
			if idx2 == 0 {
				//第一个 主类型, 并且在其他包
				if p.PkgType == constants.PackageOtherPackage {
					af1.Struct = FindType(p.PkgPath, p.TypeName, proj.BaseDir, proj.ModPkg, proj)
					af1.Slice = p.IsSlice
					if af1.Slice {
						af1.TypeName += "[]"
					}
					af1.Type = af1.Struct.Type
					af1.Pointer = p.IsPtr
					if af1.Pointer {
						af1.TypeName += "*" + af1.Struct.Package.Name + "." + af1.Struct.Type
					} else {
						af1.TypeName += af1.Struct.Package.Name + "." + af1.Struct.Type
					}

					af1.Package.Path = af1.Struct.Package.Path
					af1.Package.Name = af1.Struct.Package.Name
					af1.Package.Type = constants.PackageOtherPackage
					if p.IsGeneric {
						af1.Generic = true
						if af1.TypeParam == nil {
							af1.TypeParam = make([]*types.TypeParam, 0)
						}
					}
				} else {
					af1.Package.Type = p.PkgType
					af1.Slice = p.IsSlice
					if af1.Slice {
						af1.TypeName += "[]"
					}
					af1.Type = p.TypeName
					af1.TypeName += p.TypeName
					if p.IsGeneric {
						af1.Generic = true
						if af1.TypeParam == nil {
							af1.TypeParam = make([]*types.TypeParam, 0)
						}
					}
				}
			} else {
				if af1.Generic {
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
						af1.TypeParam = append(af1.TypeParam, tp)
					} else {
						af1.TypeParam = append(af1.TypeParam, &types.TypeParam{
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
			af1.TypeName += "[" + strings.Join(tpString, ",") + "]"
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
