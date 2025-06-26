package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func getPackageType(pkgName string, typeName string, modPkg string) string {
	if internal.IsInternalType(typeName) {
		return constants.PackageBuiltin
	}
	if pkgName == "" {
		if internal.IsInternalGenericType(typeName) {
			return constants.PackageBuiltin
		}
		return constants.PackageSamePackage
	} else {
		if strings.HasPrefix(pkgName, modPkg) {
			return constants.PackageOtherPackage
		} else {
			return constants.PackageThirdPackage
		}
	}
}

var builtInPackages = map[string]bool{
	"mime":    true,
	"time":    true,
	"errors":  true,
	"net":     true,
	"go":      true,
	"math":    true,
	"strconv": true,
	"path":    true,
	"os":      true,
}

// checkPackage 返回某个包是何种类型的包
func checkPackage(modPkg string, pkg string) string {
	if pkg == constants.PackageSamePackage || (strings.HasPrefix(pkg, modPkg) && pkg == modPkg) {
		return constants.PackageSamePackage
	}
	idx := strings.Index(pkg, "/")
	var pkgPrefix string
	if idx < 0 {
		pkgPrefix = pkg
	} else {
		pkgPrefix = pkg[:idx]
	}
	if _, ok := builtInPackages[pkgPrefix]; ok {
		return constants.PackageBuiltin
	}
	if pkg == constants.PackageBuiltin {
		return constants.PackageBuiltin
	}
	if strings.HasPrefix(pkg, modPkg) {
		return constants.PackageOtherPackage
	}
	return constants.PackageThirdPackage
}
func findPackage(expr ast.Expr, imports []*types.Import, modPkg string) []*types.PkgType {
	if expr == nil {
		return []*types.PkgType{}
	}

	result := make([]*types.PkgType, 0)
	switch spec := expr.(type) {
	case *ast.Ident: //直接一个类型

		return []*types.PkgType{
			{
				IsGeneric: internal.IsInternalGenericType(spec.Name),
				PkgPath:   "",
				PkgName:   "",
				TypeName:  spec.Name,
				PkgType:   getPackageType("", spec.Name, modPkg),
			},
		}
	case *ast.SelectorExpr: //带包的类型
		pkgName := spec.X.(*ast.Ident).Name
		typeName := spec.Sel.Name
		pkgPath := ""
		pkgType := ""
		for _, i3 := range imports {
			if i3.Name == pkgName || i3.Alias == pkgName {
				pkgPath = i3.Path
				pkgType = checkPackage(modPkg, i3.Path)
			}
		}
		return []*types.PkgType{
			{
				IsGeneric: false,
				PkgPath:   pkgPath,
				TypeName:  typeName,
				PkgName:   pkgName,
				PkgType:   pkgType,
			},
		}
	case *ast.StarExpr: //指针
		aa := findPackage(spec.X, imports, modPkg)
		aa[0].IsPtr = true
		//for _, pkgType := range aa {
		//	pkgType.IsPtr = true
		//}
		result = append(result, aa...)
		return result
	case *ast.ArrayType: //数组
		aa := findPackage(spec.Elt, imports, modPkg)
		for _, pkgType := range aa {
			pkgType.IsSlice = true
		}
		result = append(result, aa...)
		return result
	case *ast.Ellipsis: // ...
		aa := findPackage(spec.Elt, imports, modPkg)
		for _, pkgType := range aa {
			pkgType.IsSlice = true
		}
		result = append(result, aa...)
		return result
	case *ast.MapType:
		bb := findPackage(spec.Key, imports, modPkg)
		result = append(result, bb...)
		aa := findPackage(spec.Value, imports, modPkg)
		result = append(result, aa...)
		return result
	case *ast.IndexExpr: //泛型
		bb := findPackage(spec.X, imports, modPkg) //主类型
		for _, pkgType := range bb {
			pkgType.IsGeneric = true
		}
		result = append(result, bb...)
		aa := findPackage(spec.Index, imports, modPkg) //泛型类型
		result = append(result, aa...)
		return result
	case *ast.IndexListExpr: //多个泛型参数
		bb := findPackage(spec.X, imports, modPkg)
		for _, pkgType := range bb {
			pkgType.IsGeneric = true
		}
		result = append(result, bb...)
		for _, indic := range spec.Indices {
			aa := findPackage(indic, imports, modPkg)
			for _, pkgType := range aa {
				pkgType.IsGeneric = true
			}
			result = append(result, aa...)
		}

		return result
	case *ast.BinaryExpr:
		aa := findPackage(spec.X, imports, modPkg)
		bb := findPackage(spec.Y, imports, modPkg)
		result = append(result, aa...)
		result = append(result, bb...)
		return result
	case *ast.InterfaceType:
		return []*types.PkgType{
			{
				IsGeneric: false,
				PkgType:   constants.PackageBuiltin,
				TypeName:  "interface{}",
			},
		}
	case *ast.ChanType:
		return []*types.PkgType{
			{
				IsGeneric: false,
				PkgType:   constants.PackageBuiltin,
				TypeName:  "chan",
			},
		}
	case *ast.StructType:
		if expr.(*ast.StructType).Fields == nil {
			return []*types.PkgType{
				{
					IsGeneric: false,
					PkgType:   constants.PackageBuiltin,
					TypeName:  "struct",
				},
			}
		} else {
			return []*types.PkgType{
				{
					IsGeneric: false,
					PkgType:   constants.PackageSamePackage,
					TypeName:  "struct",
				},
			}
		}

	default:
		return []*types.PkgType{}
	}
	panic("unreachable")
}

func findPackageV2(expr ast.Expr, root *types.TypePkgInfo) {
	if expr == nil {
		root.Valid = false
		return
	}
	switch spec := expr.(type) {
	case *ast.Ident: //直接一个类型
		root.Generic = internal.IsInternalGenericType(spec.Name)
		root.PkgPath = ""
		root.PkgName = ""
		root.Name = spec.Name
		root.FullName = spec.Name
		root.Valid = true
		root.PkgType = getPackageType("", spec.Name, root.ModPkg)
		return
	case *ast.SelectorExpr: //带包的类型
		pkgName := spec.X.(*ast.Ident).Name
		typeName := spec.Sel.Name
		pkgPath := ""
		pkgType := ""
		for _, i3 := range root.Imports {
			if i3.Name == pkgName || i3.Alias == pkgName {
				pkgPath = i3.Path
				pkgType = checkPackage(root.ModPkg, i3.Path)
			}
		}
		root.Generic = false
		root.PkgPath = pkgPath
		root.PkgName = pkgName
		root.Name = typeName
		root.Valid = true
		root.FullName = pkgName + "." + typeName
		root.PkgType = pkgType
		return
	case *ast.StarExpr: //指针
		findPackageV2(spec.X, root)
		root.Pointer = true
		root.FullName = "*" + root.FullName
		return
	case *ast.ArrayType: //数组
		findPackageV2(spec.Elt, root)
		root.Slice = true
		root.Valid = true
		root.FullName = "[]" + root.FullName
		return
	case *ast.Ellipsis: // ...
		findPackageV2(spec.Elt, root)
		root.Slice = true
		root.Valid = true
		root.FullName = "[]" + root.FullName
		return
	case *ast.MapType:
		root.Name = "map"
		root.PkgType = constants.PackageBuiltin
		root.PkgPath = ""
		root.PkgName = ""
		root.FullName = "map"
		root.Valid = true
		child := types.NewTypePkgInfo(root.ModPkg, root.CurrentPkg, root.Imports)
		child.Imports = root.Imports
		child.ModPkg = root.ModPkg
		findPackageV2(spec.Key, child)
		root.Children = append(root.Children, child)
		root.FullName += "[" + child.FullName + "]"
		child1 := types.NewTypePkgInfo(root.ModPkg, root.CurrentPkg, root.Imports)
		findPackageV2(spec.Value, child1)
		root.Children = append(root.Children, child1)
		root.FullName += child1.FullName
		return
	case *ast.IndexExpr: //泛型
		findPackageV2(spec.X, root) //主类型
		root.Generic = true
		root.Valid = true
		child := types.NewTypePkgInfo(root.ModPkg, root.CurrentPkg, root.Imports)
		findPackageV2(spec.Index, child) //泛型类型
		child.Generic = true

		root.Children = append(root.Children, child)
		root.FullName += "[" + child.FullName + "]"
		return
	case *ast.IndexListExpr: //多个泛型参数
		findPackageV2(spec.X, root) //主类型
		root.Generic = true
		var tpString []string
		for _, indic := range spec.Indices {
			child := types.NewTypePkgInfo(root.ModPkg, root.CurrentPkg, root.Imports)
			findPackageV2(indic, child)
			child.Generic = true
			root.Children = append(root.Children, child)
			tpString = append(tpString, child.FullName)
		}
		root.Valid = true
		root.FullName += "[" + strings.Join(tpString, ",") + "]"
		return
	case *ast.BinaryExpr:
		root.Valid = false
		return
	case *ast.InterfaceType:
		root.Name = "interface{}"
		root.Valid = true
		root.PkgType = constants.PackageBuiltin
		return
	case *ast.ChanType:
		root.Name = "chan"
		root.Valid = true
		root.PkgType = constants.PackageBuiltin
		return
	case *ast.StructType:
		if expr.(*ast.StructType).Fields == nil {
			root.Name = "struct"
			root.Valid = true
			root.PkgType = constants.PackageBuiltin
			return
		} else {
			root.Name = "struct"
			root.Valid = true
			root.PkgType = constants.PackageSamePackage
			root.PkgPath = root.CurrentPkg

			return
		}

	default:
		return
	}
}
