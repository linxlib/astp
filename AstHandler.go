package astp

import (
	"fmt"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type FindHandler func(pkg string, name string) *types.Element

func NewAstHandler(f *File, astFile *ast.File, findHandler FindHandler) *AstHandler {
	return &AstHandler{
		file:        f,
		af:          astFile,
		findHandler: findHandler,
	}
}

type AstHandler struct {
	file        *File
	af          *ast.File
	findHandler FindHandler
}

func (a *AstHandler) Result() (*File, string) {
	return a.file, internal.GetKey(a.file.PackagePath, filepath.Base(a.file.FilePath))
}

func (a *AstHandler) HandlePackages() *AstHandler {
	log.Printf("parse file package: %s\n", a.file.Name)
	a.file.PackageName = a.af.Name.Name
	a.file.Docs = internal.GetDocs(a.af.Doc)

	if a.af.Comments != nil {
		for _, comment := range a.af.Comments {
			a.file.Comments = append(a.file.Comments, internal.GetComments(comment)...)
		}
	}
	return a
}
func (a *AstHandler) HandleImports() *AstHandler {
	a.file.Imports = make([]*Import, len(a.af.Imports))
	for idx, spec := range a.af.Imports {
		i := new(Import)
		if spec.Name != nil {
			i.Alias = spec.Name.Name
		} else {
			i.IsIgnore = true
			i.Alias = "_"
		}
		v := strings.Trim(spec.Path.Value, `"`)
		n := strings.LastIndex(v, "/")
		if n > 0 {
			i.Name = v[n+1:]
		} else {
			i.Name = v
		}
		i.ImportPath = v
		a.file.Imports[idx] = i
	}
	log.Printf("parse file imports: count: %d", len(a.af.Imports))
	return a
}

func (a *AstHandler) HandleElements() *AstHandler {
	a.handleConsts()
	a.handleVars()
	a.handleFunctions()
	a.handleStructs()
	return a
}

func (a *AstHandler) handleConsts() {
	log.Printf("parse file const:%s\n", a.file.PackagePath)
	for _, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.CONST {
				// 当前const区域所属的类型
				var constAreaType string
				// 当前值
				var curValue int
				//var mustBeEnum bool = false

				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &types.Element{
								Name:        v.Name,
								PackagePath: a.file.PackagePath,
								PackageName: a.file.PackageName,
								ElementType: types.ElementConst,
								ItemType:    types.ElementNone,
								Index:       i,
								Docs:        internal.GetDocs(spec.Doc),
								Comment:     internal.GetComment(spec.Comment),
							}

							// 枚举写法说明：
							// 1. 在一个const区域内有多个常量，1个的时候不符合，即必须为 const () 的声明方式
							// 2. 区域内第一个常量的类型被指定了，并且不是内置类型。 其他常量若未指定类型，则需要第一个常量采用了iota（一般这个枚举类型为 int，如果是string，则每个常量都需指定类型）
							// 3. 被指定的类型需要在当前包下（不能分开不同包，即 *ast.Ident）

							// 当判定为枚举时， ElementType 被置为 types.ElementEnum

							isEnum := true
							if len(spec.Values) != len(spec.Names) {
								if spec.Type == nil {
									isEnum = !(constAreaType == "")
									if isEnum {
										curValue++
										vv.Value = curValue
									}
								} else {
									//多重赋值 不作为枚举的写法
									isEnum = false
								}

							} else {
								if spec.Type == nil {
									isEnum = false
									isEnum = !(constAreaType == "")
									if isEnum {
										curValue++
										vv.Value = curValue
									}

								} else {
									isEnum = true
									if a, ok := spec.Type.(*ast.Ident); ok {
										isEnum = true
										vv.TypeString = a.Name
										vv.ElementString = a.Name
										constAreaType = a.Name
									} else {
										isEnum = false
										vv.TypeString = "ignore"
										vv.ElementString = "ignore"
									}

								}
							}
							if isEnum {
								vv.ElementType = types.ElementEnum
							}

							if spec.Values != nil && len(spec.Values) > 0 {
								switch vvv := spec.Values[i].(type) {
								case *ast.Ident:
									if isEnum {
										if vvv.Name == "iota" {
											vv.Value = 0
											curValue = 0
										}
									}

								case *ast.BasicLit:
									vv.Value = vvv.Value
								case *ast.BinaryExpr:
									if isEnum {
										if vvv.X.(*ast.Ident).Name == "iota" {
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
							}

							a.file.Consts = append(a.file.Consts, vv)
						}
					}
				}
			}
		}
	}
}

type PkgType struct {
	PkgPath  string
	TypeName string
}

func (a *AstHandler) findPackage(expr ast.Expr) []*PkgType {
	result := make([]*PkgType, 0)
	switch spec := expr.(type) {
	case *ast.Ident: //直接一个类型

		return []*PkgType{
			&PkgType{
				PkgPath:  types.PackagePath("", spec.Name),
				TypeName: spec.Name,
			},
		}
	case *ast.SelectorExpr: //带包的类型
		pkgName := spec.X.(*ast.Ident).Name
		typeName := spec.Sel.Name
		pkgPath := ""
		for _, i3 := range a.file.Imports {
			if i3.Name == pkgName {
				pkgPath = i3.ImportPath
			}
		}
		pp := types.PackagePath(pkgName, typeName)
		if pp != "" {
			pkgPath = pp
		}
		return []*PkgType{
			&PkgType{
				PkgPath:  pkgPath,
				TypeName: typeName,
			},
		}
	case *ast.StarExpr: //指针
		aa := a.findPackage(spec.X)
		result = append(result, aa...)
		return result
	case *ast.ArrayType: //数组
		aa := a.findPackage(spec.Elt)
		result = append(result, aa...)
		return result
	case *ast.Ellipsis: // ...
		aa := a.findPackage(spec.Elt)
		result = append(result, aa...)
		return result
	case *ast.MapType:
		bb := a.findPackage(spec.Key)
		result = append(result, bb...)
		aa := a.findPackage(spec.Value)
		result = append(result, aa...)
		return result
	case *ast.IndexExpr:
		bb := a.findPackage(spec.X)
		result = append(result, bb...)
		aa := a.findPackage(spec.Index)
		result = append(result, aa...)
		return result
	case *ast.IndexListExpr:
		bb := a.findPackage(spec.X)
		result = append(result, bb...)
		for _, indic := range spec.Indices {
			aa := a.findPackage(indic)
			result = append(result, aa...)
		}
		return result
	case *ast.BinaryExpr:
		aa := a.findPackage(spec.X)
		bb := a.findPackage(spec.Y)
		result = append(result, aa...)
		result = append(result, bb...)
		return result
	case *ast.InterfaceType:
		return []*PkgType{
			&PkgType{
				PkgPath:  types.PackageBuiltIn,
				TypeName: "interface{}",
			},
		}
	default:
		panic("unhandled expr")
	}
	panic("unreachable")
}

func (a *AstHandler) parseResults(params *ast.FieldList, tParams []*types.Element) []*types.Element {
	if params == nil {
		return nil
	}
	log.Printf("parse method result: count: %d \n", len(params.List))
	pars := make([]*types.Element, 0)
	var pIndex int
	for _, param := range params.List {
		if param.Names != nil {
			for _, name := range param.Names {
				par := &types.Element{
					Index:       pIndex,
					Name:        name.Name,
					PackagePath: a.file.PackagePath,
					PackageName: a.file.PackageName,
					ElementType: types.ElementField,
					Docs:        internal.GetDocs(param.Doc),
					Comment:     internal.GetComment(param.Comment),
				}

				ps := a.findPackage(param.Type)
				for _, p := range ps {
					if p.PkgPath != types.PackageThisPackage && p.PkgPath != types.PackageBuiltIn {
						par.Item = a.findHandler(p.PkgPath, p.TypeName)
						par.ItemType = par.Item.ItemType
						par.PackagePath = par.Item.PackagePath
						par.PackageName = par.Item.PackageName
					} else {
						// tParams
						for _, tParam := range tParams {
							if tParam.Name == p.TypeName {
								par.Item = tParam.Clone()
								par.ItemType = tParam.ElementType

							}
						}
					}
				}

				pars = append(pars, par)
				pIndex++
			}
		} else { //返回值可能为隐式参数

			par := &types.Element{
				Index:       pIndex,
				Name:        "",
				PackagePath: a.file.PackagePath,
				PackageName: a.file.PackageName,
				ElementType: types.ElementField,
				Docs:        internal.GetDocs(param.Doc),
				Comment:     internal.GetComment(param.Comment),
			}
			ps := a.findPackage(param.Type)
			for _, p := range ps {
				if p.PkgPath != types.PackageThisPackage && p.PkgPath != types.PackageBuiltIn {
					par.Item = a.findHandler(p.PkgPath, p.TypeName)
					par.ItemType = par.Item.ItemType
					par.PackagePath = par.Item.PackagePath
					par.PackageName = par.Item.PackageName
				} else {
					// tParams
					for _, tParam := range tParams {
						if tParam.Name == p.TypeName {
							par.Item = tParam.Clone()
							par.ItemType = tParam.ElementType
						}
					}
				}
			}
			pars = append(pars, par)
			pIndex++
		}

	}
	return pars
}

func (a *AstHandler) parseParams(params *ast.FieldList, tParams []*types.Element) []*types.Element {
	if params == nil {
		return nil
	}
	log.Printf("parse method params: count: %d \n", len(params.List))
	pars := make([]*types.Element, 0)
	var pIndex int

	for _, param := range params.List {
		for _, name := range param.Names {
			par := &types.Element{
				Index:       pIndex,
				Name:        name.Name,
				PackagePath: a.file.PackagePath,
				PackageName: a.file.PackageName,
				ElementType: types.ElementField,
				Docs:        internal.GetDocs(param.Doc),
				Comment:     internal.GetComment(param.Comment),
			}
			ps := a.findPackage(param.Type)
			for _, p := range ps {
				if p.PkgPath != types.PackageThisPackage && p.PkgPath != types.PackageBuiltIn {
					par.Item = a.findHandler(p.PkgPath, p.TypeName)
					par.ItemType = par.Item.ItemType
					par.PackagePath = par.Item.PackagePath
					par.PackageName = par.Item.PackageName
				} else {
					// tParams
					for _, tParam := range tParams {
						if tParam.Name == p.TypeName {
							par.Item = tParam.Clone()
							par.ItemType = tParam.ElementType
						}
					}
				}
			}
			pars = append(pars, par)
			pIndex++
		}
	}
	return pars
}

func (a *AstHandler) parseFields(fields []*ast.Field, tParams []*types.Element) []*types.Element {
	log.Printf("parse fields: count: %d\n", len(fields))
	var sf = make([]*types.Element, 0)
	for idx, field := range fields {
		af1 := new(types.Element)
		af1.Index = idx
		af1.Comment = internal.GetComment(field.Comment)
		af1.Docs = internal.GetDocs(field.Doc)
		af1.PackagePath = a.file.PackagePath
		af1.PackageName = a.file.PackageName
		af1.ElementType = types.ElementField
		if field.Names != nil {
			af1.Name = field.Names[0].Name
		}
		ps := a.findPackage(field.Type)
		for _, p := range ps {
			if p.PkgPath != types.PackageThisPackage && p.PkgPath != types.PackageBuiltIn {
				af1.Item = a.findHandler(p.PkgPath, p.TypeName)
				af1.ItemType = af1.Item.ElementType
				af1.PackagePath = af1.Item.PackagePath
				af1.PackageName = af1.Item.PackageName
			} else {
				// tParams
				for _, tParam := range tParams {
					if tParam.Name == p.TypeName {
						af1.Item = tParam.Clone()
						af1.ItemType = tParam.ElementType
					}
				}
			}
		}

		if af1.Name == "" || af1.Name == "_" {
			af1.FromParent = true

			// 将继承的字段合并到当前结构
			//if af1.Item != nil && af1.Item.Elements[types.ElementField] != nil {
			//	sf = append(sf, af1.Item.Elements[types.ElementField]...)
			//}
		}

		if field.Tag != nil {
			af1.Tag = reflect.StructTag(field.Tag.Value)
		}

		sf = append(sf, af1)
	}
	return sf
}

func (a *AstHandler) parseReceiver(fieldList *ast.FieldList, s *types.Element) *types.Element {
	result := &types.Element{
		PackagePath:   a.file.PackagePath,
		PackageName:   a.file.PackageName,
		ElementType:   types.ElementReceiver,
		Index:         0,
		ElementString: "(c *Controller)",
		Signature:     "(c *Controller)",
		Elements:      make(map[types.ElementType][]*types.Element),
	}
	receiver := fieldList.List[0]
	name := fieldList.List[0].Names[0].Name
	result.Name = name
	result.ItemType = types.ElementStruct
	result.Item = s.Clone()
	switch decl := receiver.Type.(type) {
	case *ast.Ident:
		if decl.Name == s.Name {
			result.ElementString = fmt.Sprintf("%s %s", result.Name, result.Item.Name)
			result.Signature = result.ElementString
			result.TypeString = result.Item.Name
		}

	case *ast.StarExpr:
		switch spec := decl.X.(type) {
		case *ast.Ident:
			if spec.Name == s.Name {
				result.ElementString = fmt.Sprintf("%s *%s", result.Name, result.Item.Name)
				result.Signature = result.ElementString
				result.TypeString = result.Item.Name
			}
			//TODO: 格式化输出泛型的表示形式
		case *ast.IndexExpr: //[T]
			if spec.X.(*ast.Ident).Name == s.Name {
				result.ElementString = fmt.Sprintf("%s *%s[T]", result.Name, result.Item.Name)
				result.Signature = result.ElementString
				result.TypeString = result.Item.Name
			}
		case *ast.IndexListExpr: //[T,E]
			if spec.X.(*ast.Ident).Name == s.Name {
				result.ElementString = fmt.Sprintf("%s *%s[T,E]", result.Name, result.Item.Name)
				result.Signature = result.ElementString
				result.TypeString = result.Item.Name
			}
		}
	}
	return result
}

func (a *AstHandler) parseMethods(s *types.Element) []*types.Element {

	methods := make([]*types.Element, 0)
	for idx, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				continue
			}
			recv := a.parseReceiver(decl.Recv, s)
			if recv.Item != nil && recv.Item.Name == s.Name {
				log.Printf("parse methods: %s \n", decl.Name.Name)
				method := &types.Element{
					Index:       idx,
					PackagePath: a.file.PackagePath,
					Name:        decl.Name.Name,
					Docs:        internal.GetDocs(decl.Doc),
					ElementType: types.ElementMethod,
					Elements:    make(map[types.ElementType][]*types.Element),
				}
				method.Elements[types.ElementReceiver] = []*types.Element{recv}

				method.Elements[types.ElementParam] = a.parseParams(decl.Type.Params, s.Elements[types.ElementGeneric])
				method.Elements[types.ElementResult] = a.parseResults(decl.Type.Results, s.Elements[types.ElementGeneric])

				methods = append(methods, method)
			}

		}
	}
	return methods
}

func (a *AstHandler) handleVars() {
	log.Printf("parse file vars:%s\n", a.file.PackagePath)
	a.file.Vars = make([]*types.Element, 0)
	for _, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.VAR:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &types.Element{
								Name:        v.Name,
								PackagePath: a.file.PackagePath,
								PackageName: a.file.PackageName,
								ElementType: types.ElementVar,
								ItemType:    types.ElementNone,
								Index:       i,
								Docs:        internal.GetDocs(spec.Doc),
								Comment:     internal.GetComment(spec.Comment),
							}

							if len(spec.Values) == len(spec.Names) {
								if a, ok := spec.Values[i].(*ast.BasicLit); ok {
									vv.Value = a.Value
								}
							}
							ps := a.findPackage(spec.Type)
							for _, p := range ps {
								if p.PkgPath != types.PackageThisPackage && p.PkgPath != types.PackageBuiltIn {
									vv.Item = a.findHandler(p.PkgPath, p.TypeName)
									vv.ItemType = vv.Item.ItemType
								}
							}
							a.file.Vars = append(a.file.Vars, vv)
						}
					}
				}
			default:
				continue
			}
		}
	}
}

func (a *AstHandler) handleTypeParam(expr ast.Expr) string {
	switch spec := expr.(type) {
	case *ast.Ident:
		return spec.Name
	case *ast.SelectorExpr:
		return spec.Sel.Name
	case *ast.IndexExpr:
		return a.handleTypeParam(spec.X)
	default:
		return ""
	}
}

// parseTypeParams
// 解析泛型类型
//
// @tParams:
//
// 1. 对于struct成员field，需要带上struct的泛型参数
//
// 2. 对于方法，需要带上其receiver的泛型参数
//
// 3. 对于函数，则在解析其参数和返回值时才需要带上函数自身定义的泛型参数
func (a *AstHandler) parseTypeParams(list *ast.FieldList, tParams []*types.Element) []*types.Element {
	log.Printf("parse type param: count: %d", len(list.List))
	result := make([]*types.Element, 0)

	for _, field := range list.List {

		for _, name := range field.Names {
			t := new(types.Element)
			switch spec := field.Type.(type) {
			case *ast.BinaryExpr:
				ss := a.parseBinaryExpr(spec)
				t.ElementString = strings.Join(ss, "|")
				if internal.IsInternalType(ss[0]) {
					t.PackagePath = types.PackageBuiltIn
				}
			case *ast.Ident:
				t.ElementString = spec.String()
				if internal.IsInternalType(t.ElementString) {
					t.PackagePath = types.PackageBuiltIn
				}
			}
			//t.PackagePath = types.PackageThisPackage
			t.PackageName = a.file.PackageName
			t.Name = name.Name
			t.ElementType = types.ElementGeneric
			ps := a.findPackage(field.Type)
			for _, p := range ps {
				if p.PkgPath != types.PackageThisPackage && p.PkgPath != types.PackageBuiltIn {
					t.Item = a.findHandler(p.PkgPath, p.TypeName)
					t.ItemType = t.Item.ItemType
					t.PackagePath = t.Item.PackagePath
					t.PackageName = t.Item.PackageName
				} else {
					//t.PackagePath += p.PkgPath + ","

					// tParams
					for _, tParam := range tParams {
						if tParam.Name == p.TypeName {
							t.Item = tParam.Clone()
							t.ItemType = tParam.ElementType
						}
					}
				}
			}

			result = append(result, t)
		}

	}
	return result
}

func (a *AstHandler) handleStructs() {
	log.Printf("parse file structs:%s\n", a.file.PackagePath)
	for _, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:

						e := &types.Element{
							Name:          spec.Name.Name,
							PackagePath:   a.file.PackagePath,
							PackageName:   a.file.PackageName,
							ElementType:   types.ElementStruct,
							TypeString:    spec.Name.Name,
							ElementString: spec.Name.Name,
							Index:         0,
							Comment:       internal.GetComment(spec.Comment),
							Elements:      make(map[types.ElementType][]*types.Element),
						}
						if spec.Doc == nil {
							e.Docs = internal.GetDocs(decl.Doc)
						} else {
							e.Docs = internal.GetDocs(spec.Doc)
						}
						if spec.TypeParams != nil {
							e.Elements[types.ElementGeneric] = a.parseTypeParams(spec.TypeParams, []*types.Element{})
						}
						switch spec1 := spec.Type.(type) {
						case *ast.StructType:
							{
								log.Printf("parse struct fields:%s\n", e.Name)
								e.Elements[types.ElementField] = a.parseFields(spec1.Fields.List, e.Elements[types.ElementGeneric])
								log.Printf("parse struct methods:%s\n", e.Name)
								methods := a.parseMethods(e)
								log.Printf("parse struct methods:count:%d\n", len(methods))
								e.Elements[types.ElementMethod] = append(e.Elements[types.ElementMethod], methods...)

								for _, field := range e.Elements[types.ElementField] {
									if field.FromParent {
										e.FromParent = true
										// TODO: only export method
										//if field.Item != nil && field.PackagePath != types.PackageThisPackage && len(field.Item.Elements[types.ElementMethod]) > 0 {
										//	// TODO: fill parent method with actual param type
										//	for _, method := range field.Item.Elements[types.ElementMethod] {
										//		for _, param := range method.Elements[types.ElementParam] {
										//			param.PackagePath = a.file.PackagePath
										//
										//		}
										//	}
										//
										//	field.Item.Elements[types.ElementMethod] = append(field.Item.Elements[types.ElementMethod], field.Item.Elements[types.ElementMethod]...)
										//	e.Docs = append(e.Docs, field.Item.Docs...)
										//}

									}
								}

								// 将结构体的方法加一份到文件的方法列表
								if a.file.Methods == nil {
									a.file.Methods = make([]*types.Element, 0)
								}
								a.file.Methods = append(a.file.Methods, methods...)

							}

						case *ast.InterfaceType:
							log.Printf("parse interface:%s\n", e.Name)
							e.ElementType = types.ElementInterface
							e.Elements[types.ElementInterface] = a.parseInterfaces(spec1.Methods.List, e.Elements[types.ElementGeneric])
						default:

						}
						a.file.Structs = append(a.file.Structs, e)

					}
				}
			}
		}
	}

}
func (a *AstHandler) handleFunctions() {
	log.Printf("parse functions: %s \n", a.file.PackagePath)
	methods := make([]*types.Element, 0)
	funcIndex := 0
	for _, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {

				method := &types.Element{
					PackagePath:   a.file.PackagePath,
					PackageName:   a.file.PackageName,
					Name:          decl.Name.Name,
					Index:         funcIndex,
					Docs:          internal.GetDocs(decl.Doc),
					ElementType:   types.ElementFunc,
					ElementString: "func " + decl.Name.Name + "()",
					Signature:     "func " + decl.Name.Name + "()",
					Elements:      make(map[types.ElementType][]*types.Element),
				}
				funcIndex++

				if decl.Type.TypeParams != nil {
					method.Elements[types.ElementGeneric] = a.parseTypeParams(decl.Type.TypeParams, []*types.Element{})
				}
				method.Elements[types.ElementParam] = a.parseParams(decl.Type.Params, method.Elements[types.ElementGeneric])
				method.Elements[types.ElementResult] = a.parseResults(decl.Type.Results, method.Elements[types.ElementGeneric])
				methods = append(methods, method)
			}

		}
	}
	a.file.Funcs = methods
}

func (a *AstHandler) parseInterfaces(list []*ast.Field, tParams []*types.Element) []*types.Element {
	log.Printf("parse interface: method count: %d \n", len(list))
	interaceFields := make([]*types.Element, 0)

	for i, field := range list {
		name := ""
		if field.Names != nil {
			name = field.Names[0].Name
		}
		item := &types.Element{
			Name:          name,
			PackagePath:   a.file.PackagePath,
			PackageName:   a.file.PackageName,
			Index:         i,
			Tag:           "",
			ElementString: "",
			Signature:     "",
			Docs:          internal.GetDocs(field.Doc),
			Comment:       internal.GetComment(field.Comment),
			Elements:      make(map[types.ElementType][]*types.Element),
		}
		switch spec := field.Type.(type) {
		case *ast.FuncType:
			item.ElementType = types.ElementMethod
			item.Elements[types.ElementParam] = a.parseParams(spec.Params, tParams)
			item.Elements[types.ElementResult] = a.parseParams(spec.Results, tParams)
			item.ElementString = "func ()"
		case *ast.BinaryExpr:
			item.ElementType = types.ElementConstrain
			vv := a.parseBinaryExpr(spec)
			item.ElementString = strings.Join(vv, "|")
		case *ast.Ident:
			item.ElementType = types.ElementConstrain
			item.ElementString = spec.Name
		default:
			item.ElementType = types.ElementConstrain
			item.ElementString = "fuck!!!! here is type constraints!"
		}
		interaceFields = append(interaceFields, item)
	}

	return interaceFields
}

func (a *AstHandler) parseBinaryExpr(expr ast.Expr) (result []string) {
	result = make([]string, 0)
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		if expr.Op == token.OR {
			xx := a.parseBinaryExpr(expr.X)
			result = append(result, xx...)
			yy := a.parseBinaryExpr(expr.Y)
			result = append(result, yy...)
			return result
		}
	case *ast.Ident:
		result = append(result, expr.Name)
		return result
	default:
		panic("unhandled expression")
	}
	return result
}
