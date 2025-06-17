package internal

import "strings"

var internalType = []string{"string", "bool", "int", "uint", "byte", "rune",
	"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "uintptr",
	"float32", "float64", "map", "Time", "error", "any", "FileHeader"}

// IsInternalType 是否是内部类型
func IsInternalType(t string) bool {
	for _, v := range internalType {
		if strings.EqualFold(t, v) {
			return true
		}
	}
	return false
}

var internalGenericTypes = []string{
	"T",
	"E",
	"R",
	"S",
	"K",
	"V",
}

// IsInternalGenericType 是否是内部泛型类型
func IsInternalGenericType(t string) bool {
	for _, v := range internalGenericTypes {
		if strings.EqualFold(t, v) {
			return true
		}
	}
	return false
}
