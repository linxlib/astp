package parsers

import (
	"fmt"
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
	"go/token"
)

func parseStruct(af *ast.File, p *types.Package, imports []*types.Import, proj *types.Project) []*types.Struct {
	var structs []*types.Struct
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for index, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						e := &types.Struct{
							Name:     spec.Name.Name,
							TypeName: spec.Name.Name,
							Type:     spec.Name.Name,
							ElemType: constants.ElemStruct,
							Index:    index,
							Package:  p.Clone(),
							Top:      true,
							Comment:  parseDoc(spec.Comment, spec.Name.Name),
						}
						e.TypeName = e.Package.Name + "." + e.Type
						e.Key = internal.GetKey(p.Path, e.Name)
						e.KeyHash = internal.GetKeyHash(p.Path, e.Name)
						e.Private = internal.IsPrivate(spec.Name.Name)
						if spec.Doc == nil {
							e.Doc = parseDoc(decl.Doc, spec.Name.Name)
						} else {
							e.Doc = parseDoc(spec.Doc, spec.Name.Name)
						}
						if spec.TypeParams != nil {
							e.Generic = true
							e.TypeParam = parseTypeParamV2(spec.TypeParams, imports, proj)
							for _, param := range e.TypeParam {
								param.Key = fmt.Sprintf("%s_%d_%s", e.Type, param.Index, param.Type)
							}

						}
						switch spec1 := spec.Type.(type) {
						case *ast.StructType:
							{
								// 解析字段时, 如果其中有泛型类型, 应该和上面的泛型类型一一对应, 可以生成一个唯一的key
								// 这样方便后续使用实际类型去覆盖泛型类型时好匹配到
								e.Field = parseField(spec1.Fields.List, e.TypeParam, imports, proj)

								e.Method = parseMethod(af, e, imports, proj)

							}

						case *ast.Ident:
							// just like type A = string / int
							e.TypeName = spec1.Name
							e.Type = spec1.Name
						case *ast.InterfaceType:

						default:

						}
						structs = append(structs, e)

					}
				}
			}
		}
	}
	return structs
}
