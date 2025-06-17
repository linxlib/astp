package parsers

import (
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
)

//func ParseDoc(af *ast.File) []*types.Comment {
//	var result = make([]*types.Comment, 0)
//	if af.Doc != nil && af.Doc.List != nil {
//		comments := internal.GetComments(af.Doc)
//		for i, comment := range comments {
//			result = append(result, types.OfComment(i, comment))
//		}
//	}
//	return result
//}

func HandleDoc(cg *ast.CommentGroup) []*types.Comment {
	var result = make([]*types.Comment, 0)
	if cg != nil && cg.List != nil {
		comments := internal.GetComments(cg)
		for i, comment := range comments {
			result = append(result, types.OfComment(i, comment))
		}
	}
	return result
}
func HandleDocs(cgs []*ast.CommentGroup) []*types.Comment {
	var result = make([]*types.Comment, 0)
	for _, cg := range cgs {
		if cg != nil && cg.List != nil {
			comments := internal.GetComments(cg)
			for i, comment := range comments {
				result = append(result, types.OfComment(i, comment))
			}
		}
	}
	return result
}
