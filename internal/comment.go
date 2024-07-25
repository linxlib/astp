package internal

import (
	"go/ast"
	"strings"
)

func trimComment(line string) string {
	return strings.TrimLeft(strings.TrimLeft(line, "//"), " ")
}

func GetComment(cg *ast.CommentGroup) string {
	var result string
	if cg != nil && cg.List != nil {
		for _, comment := range cg.List {
			result += trimComment(comment.Text) + " "
		}
	}
	return result
}
func GetComments(cg *ast.CommentGroup) []string {
	var result []string
	if cg != nil && cg.List != nil {
		for _, comment := range cg.List {
			result = append(result, trimComment(comment.Text))
		}
	}
	return result
}

func GetDocs(cg *ast.CommentGroup) []string {
	var result []string
	if cg != nil && cg.List != nil {
		for _, comment := range cg.List {
			result = append(result, trimComment(comment.Text))
		}
	}
	return result
}
