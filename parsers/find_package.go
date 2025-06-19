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
				PkgType:   pkgType,
			},
		}
	case *ast.StarExpr: //指针
		aa := findPackage(spec.X, imports, modPkg)
		for _, pkgType := range aa {
			pkgType.IsPtr = true
		}
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
