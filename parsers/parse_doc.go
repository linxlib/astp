package parsers

import (
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func parseDoc(cg *ast.CommentGroup, selfName string) []*types.Comment {
	var result = make([]*types.Comment, 0)
	if cg != nil && cg.List != nil {
		comments := internal.GetComments(cg)
		for i, comment := range comments {
			result = append(result, types.OfComment(i, comment, selfName))
		}
	}
	return result
}

func HandleDocs(cgs []*ast.CommentGroup, name string) []*types.Comment {
	var result = make([]*types.Comment, 0)
	for _, cg := range cgs {
		if cg != nil && cg.List != nil {
			comments := internal.GetComments(cg)
			for i, comment := range comments {
				result = append(result, types.OfComment(i, comment, "Package "+name))
			}
		}
	}
	return result
}
