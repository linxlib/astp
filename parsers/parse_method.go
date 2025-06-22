package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func parseMethod(af *ast.File, s *types.Struct, imports []*types.Import, proj *types.Project) []*types.Function {
	methods := make([]*types.Function, 0)
	methodIndex := 0
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				continue
			}
			recv := parseReceiver(decl.Recv, s, imports, proj)
			//TODO: 此处的当前结构的方法的判断应该由 parseReceiver 处理
			if recv != nil && recv.Struct != nil && recv.Struct.Name == s.Name {
				//log.Printf("      解析方法: %s \n", decl.Name.Name)
				method := &types.Function{
					Index:    methodIndex,
					Name:     decl.Name.Name,
					Doc:      parseDoc(decl.Doc, decl.Name.Name),
					Package:  s.Package.Clone(),
					ElemType: constants.ElemFunc,
					TypeName: decl.Name.Name,
					Private:  internal.IsPrivate(decl.Name.Name),
				}
				// 无需解析私有方法和非 @标记的方法
				if method.Private || !method.IsOp() {
					continue
				}
				//method.Key = internal.GetKey(s.Package.Path, method.Name)
				//method.KeyHash = internal.GetKeyHash(s.Package.Path, method.Name)
				method.Receiver = recv

				method.Param = parseParam(decl.Type.Params, imports, proj)
				method.Result = parseResults(decl.Type.Results, imports, proj)

				methods = append(methods, method)
				methodIndex++
			}

		}
	}
	return methods
}
