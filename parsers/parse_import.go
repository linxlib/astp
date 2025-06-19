package parsers

import (
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func parseImport(af *ast.File) []*types.Import {
	var result = make([]*types.Import, 0)
	for _, spec := range af.Imports {
		i := &types.Import{
			Name:   "",
			Alias:  "",
			Path:   "",
			Ignore: false,
		}
		v := strings.Trim(spec.Path.Value, `"`)
		n := strings.LastIndex(v, "/")
		if n > 0 {
			i.Name = v[n+1:]
		} else {
			i.Name = v
		}
		i.Path = v
		if spec.Name != nil {
			i.Alias = spec.Name.Name
			i.Ignore = i.Alias == "_"
		} else {
			i.Alias = i.Name
		}

		result = append(result, i)
	}
	return result
}
