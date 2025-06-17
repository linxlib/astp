package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func ParseFunction(af *ast.File, p *types.Package, imports []*types.Import, proj *types.Project) []*types.Function {
	methods := make([]*types.Function, 0)
	funcIndex := 0
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {

				method := &types.Function{
					Name:     decl.Name.Name,
					ElemType: constants.ElemFunc,
					TypeName: decl.Name.Name,

					Doc:     HandleDoc(decl.Doc),
					Private: internal.IsPrivate(decl.Name.Name),
					Index:   funcIndex,
					Package: p.Clone(),
				}
				method.Key = internal.GetKey(p.Path, method.Name)
				method.KeyHash = internal.GetKeyHash(p.Path, method.Name)
				funcIndex++

				if decl.Type.TypeParams != nil {
					method.Generic = true
					method.TypeParam = ParseTypeParam(decl.Type.TypeParams, imports, proj)
				}
				method.Param = ParseParam(decl.Type.Params, imports, proj)
				method.Result = ParseResults(decl.Type.Results, imports, proj)
				methods = append(methods, method)
			}

		}
	}
	return methods
}
