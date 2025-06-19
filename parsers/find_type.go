package parsers

import (
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"os"
	"path/filepath"
	"strings"
)

func getPackageDir(pkgPath string, modDir string, modPkg string) string {
	if strings.EqualFold(pkgPath, "main") { // if main return default path
		return modDir
	}
	if strings.HasPrefix(pkgPath, modPkg) {
		return modDir + strings.Replace(pkgPath[len(modPkg):], ".", "/", -1)
	}
	return ""
}

func findType(pkg string, name string, modDir string, modPkg string, proj *types.Project) *types.Struct {
	dir := getPackageDir(pkg, modDir, modPkg)
	if dir == "" {
		return nil
	}
	files, _ := os.ReadDir(dir)
	keyHash := internal.GetKeyHash(pkg, name)
	//在一个package对应的目录下，遍历所有文件 找到对应的文件
	for _, file := range files {
		b := filepath.Base(file.Name())
		key := internal.GetKeyHash(pkg, b)

		for _, f := range proj.FileMap {
			if f.KeyHash == key {
				if s := f.FindStruct(keyHash); s != nil {
					return s
				}
			}
		}
	}
	// 如果之前未解析过，则对该目录进行目录解析
	filesa := parseDir(dir, proj)
	proj.Merge(filesa)

	for _, v := range filesa {
		if s := v.FindStruct(keyHash); s != nil {
			return s
		}
	}

	return nil
}
