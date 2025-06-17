package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func ParseConst(af *ast.File, p *types.Package) []*types.Const {

	consts := make([]*types.Const, 0)
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.CONST {
				// 当前const区域所属的类型
				var constAreaType string
				// 当前值
				var curValue int
				// 是否有iota
				var hasIota bool
				// 这里一个表示一个const区域
				for idx, spec := range decl.Specs {

					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &types.Const{
								Name:    v.Name,
								Package: p.Clone(),
								Index:   idx,
								Doc:     HandleDoc(spec.Doc),
								Comment: HandleDoc(spec.Comment),
							}
							isEnum := false
							// 标记了类型，则有可能是枚举
							if spec.Type != nil {
								isEnum = true
								// 再看是否有iota
								if spec.Values != nil && len(spec.Values) > 0 {
									switch vvv := spec.Values[i].(type) {
									case *ast.Ident:
										if vvv.Name == "iota" {
											vv.Value = 0
											curValue = 0
											hasIota = true
											isEnum = true
										}

									case *ast.BasicLit:
										vv.Value = strings.Trim(vvv.Value, `"`)
										hasIota = false // 直接赋值
										isEnum = true
										vv.ElemType = constants.ElemEnum
									case *ast.BinaryExpr:
										if vvv.X.(*ast.Ident).Name == "iota" {
											hasIota = true
											isEnum = true
											curValue = 0
											switch vvv.Y.(*ast.BasicLit).Kind {
											case token.INT:
												temp, _ := strconv.Atoi(vvv.Y.(*ast.BasicLit).Value)
												if vvv.Op == token.ADD {
													curValue += temp
												} else if vvv.Op == token.SUB {
													curValue -= temp
												}
												//TODO: token.OR token.AND etc...
											default:
												panic("unhandled default case")
											}
											vv.Value = curValue
										}

									}
								}

								if a, ok := spec.Type.(*ast.Ident); ok {
									vv.Type = a.Name
									vv.TypeName = a.Name
									constAreaType = a.Name
									isEnum = true
								} else {
									isEnum = false
									vv.Type = "ignore"
									vv.TypeName = "ignore"
								}
							} else {
								if hasIota || constAreaType != "" {
									isEnum = true
									vv.Type = constAreaType
									vv.TypeName = constAreaType
									curValue++
									vv.Value = curValue

								}
							}

							if isEnum {
								vv.ElemType = constants.ElemEnum
							}
							consts = append(consts, vv)

						}
					}
				}
			}
		}
	}
	return consts
}
