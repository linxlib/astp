package types

import (
	"github.com/linxlib/astp/constants"
	"slices"
)

type IElem[T any] interface {
	String() string
	Clone() T
}

var deepClone bool = true

func SetDeepClone(deep bool) {
	deepClone = deep
}

func CopySlice[T IElem[T]](src []T) []T {
	if src == nil {
		return nil
	}
	if !deepClone {
		return slices.Clone(src)
	}
	result := make([]T, 0)
	for _, t := range src {
		//if t == nil {
		//	continue
		//}

		result = append(result, t.Clone())
	}
	return result
}
func CopySliceWithFilter[T IElem[T]](src []T, filter func(T) bool) []T {
	if src == nil {
		return nil
	}
	result := make([]T, 0)
	for _, t := range src {
		if filter(t) {
			result = append(result, t.Clone())
		}
	}
	return result
}

type PkgType struct {
	IsGeneric bool
	IsSlice   bool
	IsPtr     bool
	PkgPath   string
	PkgName   string
	TypeName  string
	PkgType   constants.PackageType
}

type TypePkgInfo struct {
	Imports    []*Import
	ModPkg     string
	CurrentPkg string
	Valid      bool
	Pointer    bool
	Slice      bool
	PkgName    string
	Name       string
	PkgType    constants.PackageType
	PkgPath    string
	Generic    bool
	FullName   string
	Children   []*TypePkgInfo
}

func NewTypePkgInfo(modPkg string, currentPkg string, imports []*Import) *TypePkgInfo {
	return &TypePkgInfo{Imports: imports, ModPkg: modPkg, CurrentPkg: currentPkg}
}
