package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"path/filepath"

	"strings"
)

func parsePackage(af *ast.File, file string, proj *types.Project) *types.Package {

	f, _ := filepath.Abs(file)
	f = strings.ReplaceAll(f, "\\", "/")
	n := strings.LastIndex(f, "/")
	f1 := strings.ReplaceAll(proj.BaseDir, "\\", "/")
	f = f[:n]
	p := &types.Package{
		FileName: filepath.Base(file),
		FilePath: file,
		Name:     af.Name.Name,
		Path:     proj.ModPkg + f[len(f1):],
		Type:     constants.PackageNormal,
	}
	return p
}
