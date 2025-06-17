package parsers

import (
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"os"
	"path/filepath"
	"strings"
)

func GetPackageDir(pkgPath string, modDir string, modPkg string) string {
	if strings.EqualFold(pkgPath, "main") { // if main return default path
		return modDir
	}
	if strings.HasPrefix(pkgPath, modPkg) {
		return modDir + strings.Replace(pkgPath[len(modPkg):], ".", "/", -1)
	}
	return ""
}

func FindType(pkg string, name string, modDir string, modPkg string, proj *types.Project) *types.Struct {
	dir := GetPackageDir(pkg, modDir, modPkg)
	if dir == "" {
		return nil
	}
	files, _ := os.ReadDir(dir)
	keyHash := internal.GetKeyHash(pkg, name)
	//在一个package对应的目录下，遍历所有文件 找到对应的文件
	for _, file := range files {
		b := filepath.Base(file.Name())
		key := internal.GetKeyHash(pkg, b)

		for _, f := range proj.File {
			if f.KeyHash == key {
				if s := FindInFile(f, keyHash); s != nil {
					return s
				}
			}
		}
	}
	// 如果之前未解析过，则对该目录进行目录解析
	filesa := ParseDir(dir, proj)
	proj.Merge(filesa)

	for _, v := range filesa {
		if s := FindInFile(v, keyHash); s != nil {
			return s
		}
	}

	return nil
}

func FindInFile(f *types.File, keyHash string) *types.Struct {
	for _, s := range f.Struct {
		if s.KeyHash == keyHash {
			return s
		}
	}

	return nil
}

func FindInProject(proj *types.Project, keyHash string) *types.Struct {
	for _, f := range proj.File {
		if s := FindInFile(f, keyHash); s != nil {
			return s
		}
	}
	return nil
}
