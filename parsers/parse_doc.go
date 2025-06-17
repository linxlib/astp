package parsers

import (
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
)

func HandleDoc(cg *ast.CommentGroup, selfName string) []*types.Comment {
	var result = make([]*types.Comment, 0)
	if cg != nil && cg.List != nil {
		comments := internal.GetComments(cg)
		for i, comment := range comments {
			result = append(result, types.OfComment(i, comment, selfName))
		}
	}
	return result
}

//func HandleDocs(cgs []*ast.CommentGroup) []*types.Comment {
//	var result = make([]*types.Comment, 0)
//	for _, cg := range cgs {
//		if cg != nil && cg.List != nil {
//			comments := internal.GetComments(cg)
//			for i, comment := range comments {
//				result = append(result, types.OfComment(i, comment))
//			}
//		}
//	}
//	return result
//}
