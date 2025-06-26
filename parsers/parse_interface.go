package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"strings"
)

func parseInterface(list []*ast.Field, imports []*types.Import, proj *types.Project) []*types.Interface {
	interaceFields := make([]*types.Interface, 0)

	for i, field := range list {
		name := ""
		if field.Names != nil {
			name = field.Names[0].Name
		}
		item := &types.Interface{
			Name:  name,
			Index: i,
			Doc:   parseDoc(field.Doc, name),
		}
		switch spec := field.Type.(type) {
		case *ast.FuncType:
			item.ElemType = constants.ElemFunc
			item.Param = parseParam(spec.Params, []*types.TypeParam{}, imports, proj)
			item.Result = parseResults(spec.Results, []*types.TypeParam{}, imports, proj)
			item.TypeName = name
		case *ast.BinaryExpr:
			item.ElemType = constants.ElemConstrain
			vv := parseBinaryExpr(spec)
			item.TypeName = strings.Join(vv, "|")
		case *ast.Ident:
			item.ElemType = constants.ElemConstrain
			item.TypeName = spec.Name
		default:
			item.ElemType = constants.ElemConstrain
			item.TypeName = "fuck!!!! here is type constraints!"
		}
		interaceFields = append(interaceFields, item)
	}

	return interaceFields
}
