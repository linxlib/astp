package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func ParseMethod(af *ast.File, s *types.Struct, imports []*types.Import, proj *types.Project) []*types.Function {
	methods := make([]*types.Function, 0)
	for idx, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				continue
			}
			recv := ParseReceiver(decl.Recv, s, imports, proj)
			//TODO: 此处的当前结构的方法的判断应该由 parseReceiver 处理
			if recv != nil && recv.Struct != nil && recv.Struct.Name == s.Name {
				//log.Printf("      解析方法: %s \n", decl.Name.Name)
				method := &types.Function{
					Index:    idx,
					Name:     decl.Name.Name,
					Doc:      HandleDoc(decl.Doc, decl.Name.Name),
					Package:  s.Package.Clone(),
					ElemType: constants.ElemFunc,
					TypeName: decl.Name.Name,
					Private:  internal.IsPrivate(decl.Name.Name),
				}
				method.Key = internal.GetKey(s.Package.Path, method.Name)
				method.KeyHash = internal.GetKeyHash(s.Package.Path, method.Name)
				method.Receiver = recv

				method.Param = ParseParam(decl.Type.Params, imports, proj)
				method.Result = ParseResults(decl.Type.Results, imports, proj)

				methods = append(methods, method)
			}

		}
	}
	return methods
}
