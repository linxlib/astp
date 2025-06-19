package parsers

import (
	"github.com/linxlib/astp/types"
	"os"
	"path/filepath"
)

// parseDir 解析一个目录
// 对于引用一个包的时候，直接解析其目录下的所有文件（不包含子目录）
func parseDir(dir string, proj *types.Project) map[string]*types.File {
	files := make(map[string]*types.File)

	fs, _ := os.ReadDir(dir)
	for _, f := range fs {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".go" {
			f1 := ParseFile(filepath.Join(dir, f.Name()), proj)
			files[f1.KeyHash] = f1.Clone()
		}
	}
	return files
}
