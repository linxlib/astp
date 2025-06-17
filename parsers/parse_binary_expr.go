package parsers

import (
	"go/ast"
	"go/token"
)

func ParseBinaryExpr(expr ast.Expr) []string {

	result := make([]string, 0)
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		if expr.Op == token.OR {
			xx := ParseBinaryExpr(expr.X)
			result = append(result, xx...)
			yy := ParseBinaryExpr(expr.Y)
			result = append(result, yy...)
			return result
		}
	case *ast.Ident:
		result = append(result, expr.Name)
		return result
	case *ast.IndexExpr:
		xx := ParseBinaryExpr(expr.X)
		result = append(result, xx...)
		idx := ParseBinaryExpr(expr.Index)
		result = append(result, idx...)
		return result
	case *ast.IndexListExpr:
		xx := ParseBinaryExpr(expr.X)
		result = append(result, xx...)
		for _, indic := range expr.Indices {
			idx := ParseBinaryExpr(indic)
			result = append(result, idx...)
		}
		return result
	default:
		panic("unhandled expression")
	}
	return result
}
