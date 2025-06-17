package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
	"go/token"
)

func ParseStruct(af *ast.File, p *types.Package, imports []*types.Import, proj *types.Project) []*types.Struct {
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
							Comment:  HandleDoc(spec.Comment, spec.Name.Name),
						}
						e.Key = internal.GetKey(p.Path, e.Name)
						e.KeyHash = internal.GetKeyHash(p.Path, e.Name)
						e.Private = internal.IsPrivate(spec.Name.Name)
						if spec.Doc == nil {
							e.Doc = HandleDoc(decl.Doc, spec.Name.Name)
						} else {
							e.Doc = HandleDoc(spec.Doc, spec.Name.Name)
						}
						if spec.TypeParams != nil {
							e.Generic = true
							e.TypeParam = ParseTypeParam(spec.TypeParams, imports, proj)
						}
						switch spec1 := spec.Type.(type) {
						case *ast.StructType:
							{
								e.Field = ParseField(spec1.Fields.List, imports, proj)
								for _, field := range e.Field {
									e.TypeParam = append(e.TypeParam, field.TypeParam...)
								}
								e.Method = ParseMethod(af, e, imports, proj)
								
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
