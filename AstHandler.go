package astp

import (
	"github.com/linxlib/astp/internal"
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

type FindHandler func(pkg string, name string) *Element

func NewAstHandler(f *File, modPkg string, astFile *ast.File, findHandler FindHandler) *AstHandler {
	return &AstHandler{
		file:        f,
		af:          astFile,
		modPkg:      modPkg,
		findHandler: findHandler,
	}
}

// AstHandler ast树解析类，一个代表一个go文件
type AstHandler struct {
	file *File
	af   *ast.File
	// 用于在项目的另外一个包中查找一个类型
	findHandler FindHandler
	// 项目的mod名
	modPkg         string
	parseFunctions bool
}

// Result 返回 File 对象
func (a *AstHandler) Result() (*File, string) {
	return a.file, internal.GetKey(a.file.PackagePath, filepath.Base(a.file.FilePath))
}

func (a *AstHandler) HandlePackages() *AstHandler {
	log.Printf("[%s] 解析文件头\n", a.file.Name)
	a.file.PackageName = a.af.Name.Name

	// 由于其他结构体 函数等上面的注释也会在 a.af 中被找到
	// 只处理main包的文档和注释，其他散落在外的注释无需解析.
	// 其他文件中的注释 可由对应模块进行解析
	if a.file.IsMain() {
		a.file.Docs = internal.GetDocs(a.af.Doc)

		if a.af.Comments != nil {
			for _, comment := range a.af.Comments {
				a.file.Comments = append(a.file.Comments, internal.GetComments(comment)...)
			}
		}
	}

	return a
}
func (a *AstHandler) HandleImports() *AstHandler {
	log.Printf("[%s] 解析导入项 \n", a.file.Name)
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
	return a
}

func (a *AstHandler) HandleElements() *AstHandler {
	a.handleConstArea()
	a.handleVarArea()
	if a.parseFunctions {
		a.handleFunctions()
	}
	a.handleStructs()
	return a
}

func (a *AstHandler) handleConstArea() {
	// 枚举写法说明：
	// 1. 在一个const区域内有多个常量，1个的时候不符合，即必须为 const () 的声明方式
	// 2. 区域内第一个常量的类型被指定了，并且不是内置类型。 其他常量若未指定类型，则需要第一个常量采用了iota（一般这个枚举类型为 int，如果是string，则每个常量都需指定类型）
	// 3. 被指定的类型需要在当前包下（不能分开不同包，即 *ast.Ident）

	// 当判定为枚举时， ElementType 被置为 types.ElementEnum

	log.Printf("[%s] 解析常量区块\n", a.file.Name)
	for _, decl := range a.af.Decls {
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
				for _, spec := range decl.Specs {

					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &Element{
								Name:        v.Name,
								PackagePath: a.file.PackagePath,
								PackageName: a.file.PackageName,
								ElementType: ElementConst,
								ItemType:    ElementNone,
								Index:       i,
								Docs:        internal.GetDocs(spec.Doc),
								Comment:     internal.GetComment(spec.Comment),
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
										vv.Value = vvv.Value
										hasIota = false // 直接赋值
										isEnum = true
										vv.ElementType = ElementEnum
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
									vv.TypeString = a.Name
									vv.ElementString = a.Name
									constAreaType = a.Name
									isEnum = true
								} else {
									isEnum = false
									vv.TypeString = "ignore"
									vv.ElementString = "ignore"
								}
							} else {
								if hasIota || constAreaType != "" {
									isEnum = true
									vv.TypeString = constAreaType
									curValue++
									vv.Value = curValue

								}
							}

							if isEnum {
								vv.ElementType = ElementEnum
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
	IsGeneric bool
	PkgPath   string
	TypeName  string
}

// findPackage 返回一个声明的类型的包地址和类型名
// 如果是泛型，则第一个为类型的，其他元素则为泛型声明的
func (a *AstHandler) findPackage(expr ast.Expr) []*PkgType {
	if expr == nil {
		return []*PkgType{}
	}
	result := make([]*PkgType, 0)
	switch spec := expr.(type) {
	case *ast.Ident: //直接一个类型

		return []*PkgType{
			&PkgType{
				IsGeneric: false,
				PkgPath:   PackagePath("", spec.Name),
				TypeName:  spec.Name,
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
		pp := PackagePath(pkgName, typeName)
		if pp != "" {
			pkgPath = pp
		}
		return []*PkgType{
			&PkgType{
				IsGeneric: false,
				PkgPath:   pkgPath,
				TypeName:  typeName,
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
		for _, pkgType := range aa {
			pkgType.IsGeneric = true
		}
		result = append(result, aa...)

		return result
	case *ast.IndexListExpr:
		bb := a.findPackage(spec.X)
		result = append(result, bb...)
		for _, indic := range spec.Indices {
			aa := a.findPackage(indic)
			for _, pkgType := range aa {
				pkgType.IsGeneric = true
			}
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
				IsGeneric: false,
				PkgPath:   PackageBuiltIn,
				TypeName:  "interface{}",
			},
		}
	case *ast.ChanType:
		return []*PkgType{
			&PkgType{
				IsGeneric: false,
				PkgPath:   PackageBuiltIn,
				TypeName:  "chan",
			},
		}
	default:
		panic("unhandled expr")
	}
	panic("unreachable")
}

func (a *AstHandler) parseResults(params *ast.FieldList, tParams []*Element) []*Element {
	if params == nil {
		return nil
	}
	log.Printf("        解析返回值: 数量: %d \n", len(params.List))
	pars := make([]*Element, 0)
	var pIndex int
	for _, param := range params.List {
		if param.Names != nil {
			for _, name := range param.Names {
				par := &Element{
					Index:       pIndex,
					Name:        name.Name,
					PackagePath: a.file.PackagePath,
					PackageName: a.file.PackageName,
					ElementType: ElementField,
					Docs:        internal.GetDocs(param.Doc),
					Comment:     internal.GetComment(param.Comment),
				}

				ps := a.findPackage(param.Type)
				for _, p := range ps {
					tmp := CheckPackage(a.modPkg, p.PkgPath)
					if tmp != PackageThisPackage && tmp != PackageBuiltIn && tmp != PackageThird {
						par.Item = a.findHandler(p.PkgPath, p.TypeName)
						par.ItemType = par.Item.ItemType
						par.PackagePath = par.Item.PackagePath
						par.PackageName = par.Item.PackageName
						par.TypeString = par.Item.TypeString
					} else {
						par.PackagePath = tmp
						par.TypeString = p.TypeName
						// tParams
						for _, tParam := range tParams {
							if tParam.Name == p.TypeName {
								par.Item = tParam.Clone()
								par.TypeString = tParam.TypeString
								par.PackagePath = tmp
								par.ItemType = tParam.ElementType

							}
						}
					}
				}

				pars = append(pars, par)
				pIndex++
			}
		} else { //返回值可能为隐式参数

			par := &Element{
				Index:       pIndex,
				Name:        "",
				PackagePath: a.file.PackagePath,
				PackageName: a.file.PackageName,
				ElementType: ElementField,
				Docs:        internal.GetDocs(param.Doc),
				Comment:     internal.GetComment(param.Comment),
			}
			ps := a.findPackage(param.Type)
			for _, p := range ps {
				tmp := CheckPackage(a.modPkg, p.PkgPath)
				if tmp != PackageThisPackage && tmp != PackageBuiltIn && tmp != PackageThird {
					par.Item = a.findHandler(p.PkgPath, p.TypeName)
					par.ItemType = par.Item.ItemType
					par.PackagePath = par.Item.PackagePath
					par.PackageName = par.Item.PackageName
				} else {
					par.PackagePath = tmp
					par.TypeString = p.TypeName
					// tParams
					for _, tParam := range tParams {
						if tParam.Name == p.TypeName {
							par.Item = tParam.Clone()
							par.PackagePath = tParam.PackagePath
							par.TypeString = tParam.TypeString
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

func (a *AstHandler) parseParams(params *ast.FieldList, tParams []*Element) []*Element {
	if params == nil {
		return nil
	}
	log.Printf("        解析参数: 数量: %d \n", len(params.List))
	pars := make([]*Element, 0)
	var pIndex int

	for _, param := range params.List {
		for _, name := range param.Names {
			par := &Element{
				Index:       pIndex,
				Name:        name.Name,
				PackagePath: a.file.PackagePath,
				PackageName: a.file.PackageName,
				ElementType: ElementField,
				Docs:        internal.GetDocs(param.Doc),
				Comment:     internal.GetComment(param.Comment),
			}
			ps := a.findPackage(param.Type)
			for _, p := range ps {
				tmp := CheckPackage(a.modPkg, p.PkgPath)
				if tmp != PackageThisPackage && tmp != PackageBuiltIn && tmp != PackageThird {
					par.Item = a.findHandler(p.PkgPath, p.TypeName)
					par.ItemType = par.Item.ItemType
					par.PackagePath = par.Item.PackagePath
					par.PackageName = par.Item.PackageName
				} else {
					par.PackagePath = tmp
					par.TypeString = p.TypeName
					// tParams
					for _, tParam := range tParams {
						if tParam.Name == p.TypeName {
							par.Item = tParam.Clone()
							par.PackagePath = tParam.PackagePath
							par.TypeString = tParam.TypeString
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

func (a *AstHandler) parseFields(fields []*ast.Field, tParams []*Element) []*Element {
	log.Printf("    解析结构体字段: %d\n", len(fields))
	var sf = make([]*Element, 0)
	for idx, field := range fields {
		af1 := new(Element)
		af1.Index = idx
		af1.Comment = internal.GetComment(field.Comment)
		af1.Docs = internal.GetDocs(field.Doc)
		af1.PackagePath = a.file.PackagePath
		af1.PackageName = a.file.PackageName
		af1.ElementType = ElementField
		if field.Names != nil {
			af1.Name = field.Names[0].Name
		}

		ps := a.findPackage(field.Type)
		idx1 := 0
		for idx2, p := range ps {
			tmp := CheckPackage(a.modPkg, p.PkgPath)
			if tmp != PackageThisPackage && tmp != PackageBuiltIn && tmp != PackageThird && idx2 == 0 {

				af1.Item = a.findHandler(p.PkgPath, p.TypeName)
				af1.ItemType = af1.Item.ElementType
				af1.PackagePath = af1.Item.PackagePath
				af1.TypeString = af1.Item.TypeString
				af1.PackageName = af1.Item.PackageName

				if p.IsGeneric {
					if af1.Elements == nil {
						af1.Elements = make(map[ElementType][]*Element)
					}
					af1.Elements[ElementGeneric] = append(af1.Elements[ElementGeneric], &Element{
						Name:          p.TypeName,
						ElementType:   ElementGeneric,
						Index:         idx1,
						TypeString:    p.TypeName,
						ElementString: p.TypeName,

						FromParent: true,
					})
					idx1++
				}

			} else {
				af1.TypeString = p.TypeName

				if p.IsGeneric {
					if af1.Elements == nil {
						af1.Elements = make(map[ElementType][]*Element)
					}
					tmp2 := CheckPackage(a.modPkg, p.PkgPath)
					if tmp2 == PackageOther || tmp2 == PackageThird {
						tmp1 := a.findHandler(p.PkgPath, p.TypeName)
						genericType := tmp1.Clone(idx1)
						genericType.ElementType = ElementGeneric
						genericType.PackagePath = p.PkgPath
						genericType.FromParent = true
						af1.Elements[ElementGeneric] = append(af1.Elements[ElementGeneric], genericType)
					} else {

						af1.Elements[ElementGeneric] = append(af1.Elements[ElementGeneric], &Element{
							Name:          p.TypeName,
							ElementType:   ElementGeneric,
							Index:         idx1,
							TypeString:    p.TypeName,
							ElementString: p.TypeName,
							PackagePath:   p.PkgPath,
							FromParent:    true,
						})
					}

					idx1++
				}
				// tParams
				for _, tParam := range tParams {
					if tParam.Name == p.TypeName {
						af1.Item = tParam.Clone()

						af1.PackagePath = tParam.PackagePath
						af1.TypeString = tParam.TypeString
						af1.ItemType = tParam.ElementType
					}
				}
			}
		}

		if af1.Name == "" || af1.Name == "_" {
			af1.FromParent = true
		}

		if field.Tag != nil {
			af1.TagString = field.Tag.Value
		}

		sf = append(sf, af1)
	}
	return sf
}

func (a *AstHandler) parseReceiver(fieldList *ast.FieldList, s *Element) *Element {

	receiver := fieldList.List[0]
	ps := a.findPackage(receiver.Type)
	typeString := ""
	if len(ps) > 0 {
		typeString = ps[0].TypeName
	}
	if typeString != s.Name {
		return nil
	}

	result := &Element{
		PackagePath: a.file.PackagePath,
		PackageName: a.file.PackageName,
		ElementType: ElementReceiver,
		Index:       0,

		Elements: make(map[ElementType][]*Element),
	}
	name := fieldList.List[0].Names[0].Name
	result.Name = name
	result.ItemType = ElementStruct
	result.Item = s.Clone()
	result.TypeString = typeString
	return result
}

func (a *AstHandler) parseMethods(s *Element) []*Element {
	log.Printf("    解析结构体方法: %s\n", s.Name)
	methods := make([]*Element, 0)
	for idx, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				continue
			}
			recv := a.parseReceiver(decl.Recv, s)
			//TODO: 此处的当前结构的方法的判断应该由 parseReceiver 处理
			if recv != nil && recv.Item != nil && recv.Item.Name == s.Name {
				log.Printf("      解析方法: %s \n", decl.Name.Name)
				method := &Element{
					Index:       idx,
					PackagePath: a.file.PackagePath,
					Name:        decl.Name.Name,
					Docs:        internal.GetDocs(decl.Doc),
					ElementType: ElementMethod,
					Elements:    make(map[ElementType][]*Element),
				}
				method.Elements[ElementReceiver] = []*Element{recv}

				method.Elements[ElementParam] = a.parseParams(decl.Type.Params, s.Elements[ElementGeneric])
				method.Elements[ElementResult] = a.parseResults(decl.Type.Results, s.Elements[ElementGeneric])

				methods = append(methods, method)
			}

		}
	}
	return methods
}

func (a *AstHandler) handleVarArea() {
	log.Printf("[%s] 解析变量区块\n", a.file.Name)
	a.file.Vars = make([]*Element, 0)
	for _, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.VAR:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &Element{
								Name:        v.Name,
								PackagePath: a.file.PackagePath,
								PackageName: a.file.PackageName,
								ElementType: ElementVar,
								ItemType:    ElementNone,
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
								tmp := CheckPackage(a.modPkg, p.PkgPath)
								if tmp != PackageThisPackage && tmp != PackageBuiltIn && tmp != PackageThird {
									vv.Item = a.findHandler(p.PkgPath, p.TypeName)
									vv.ItemType = vv.Item.ItemType
									vv.TypeString = vv.Item.TypeString
									vv.PackagePath = vv.Item.PackagePath
									vv.PackageName = vv.Item.PackageName
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
func (a *AstHandler) parseTypeParams(list *ast.FieldList, tParams []*Element) []*Element {
	log.Printf("parse type param: count: %d", len(list.List))
	result := make([]*Element, 0)
	tpIndex := 0
	for _, field := range list.List {

		for _, name := range field.Names {
			t := new(Element)
			t.Index = tpIndex
			tpIndex++
			switch spec := field.Type.(type) {
			case *ast.BinaryExpr:
				ss := a.parseBinaryExpr(spec)
				t.TypeString = strings.Join(ss, "|")
				if internal.IsInternalType(ss[0]) {
					t.PackagePath = PackageBuiltIn
				}
			case *ast.Ident:
				t.TypeString = spec.String()
				if internal.IsInternalType(t.TypeString) {
					t.PackagePath = PackageBuiltIn
				}
			}
			//t.PackagePath = types.PackageThisPackage
			t.PackageName = a.file.PackageName
			t.Name = name.Name
			t.ElementType = ElementGeneric
			ps := a.findPackage(field.Type)
			for _, p := range ps {
				tmp := CheckPackage(a.modPkg, p.PkgPath)
				if tmp != PackageThisPackage && tmp != PackageBuiltIn && tmp != PackageThird {

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
	log.Printf("[%s] 解析结构体\n", a.file.Name)
	for _, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						e := &Element{
							Name:          spec.Name.Name,
							PackagePath:   a.file.PackagePath,
							PackageName:   a.file.PackageName,
							ElementType:   ElementStruct,
							TypeString:    spec.Name.Name,
							ElementString: spec.Name.Name,
							Index:         0,
							Comment:       internal.GetComment(spec.Comment),
							Elements:      make(map[ElementType][]*Element),
						}
						if spec.Doc == nil {
							e.Docs = internal.GetDocs(decl.Doc)
						} else {
							e.Docs = internal.GetDocs(spec.Doc)
						}
						if spec.TypeParams != nil {
							e.Elements[ElementGeneric] = a.parseTypeParams(spec.TypeParams, []*Element{})
						}
						switch spec1 := spec.Type.(type) {
						case *ast.StructType:
							{
								log.Printf("  解析结构体字段: %s\n", e.Name)
								e.Elements[ElementField] = a.parseFields(spec1.Fields.List, e.Elements[ElementGeneric])
								for _, field := range e.Elements[ElementField] {
									e.Elements[ElementGeneric] = append(e.Elements[ElementGeneric], field.Elements[ElementGeneric]...)
								}
								log.Printf("  解析结构体方法:%s\n", e.Name)
								methods := a.parseMethods(e)
								e.Elements[ElementMethod] = append(e.Elements[ElementMethod], methods...)

								for _, field := range e.Elements[ElementField] {
									if field.FromParent {
										e.FromParent = true
										break
									}
								}

								// 将结构体的方法加一份到文件的方法列表
								if a.file.Methods == nil {
									a.file.Methods = make([]*Element, 0)
								}
								a.file.Methods = append(a.file.Methods, methods...)

							}

						case *ast.InterfaceType:
							log.Printf("  解析接口类型:%s\n", e.Name)
							e.ElementType = ElementInterface
							e.Elements[ElementInterface] = a.parseInterfaces(spec1.Methods.List, e.Elements[ElementGeneric])
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
	log.Printf("[%s] 解析函数\n", a.file.Name)
	methods := make([]*Element, 0)
	funcIndex := 0
	for _, decl := range a.af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {

				method := &Element{
					PackagePath: a.file.PackagePath,
					PackageName: a.file.PackageName,
					Name:        decl.Name.Name,
					Index:       funcIndex,
					Docs:        internal.GetDocs(decl.Doc),
					ElementType: ElementFunc,
					TypeString:  decl.Name.Name,
					Elements:    make(map[ElementType][]*Element),
				}
				funcIndex++

				if decl.Type.TypeParams != nil {
					method.Elements[ElementGeneric] = a.parseTypeParams(decl.Type.TypeParams, []*Element{})
				}
				method.Elements[ElementParam] = a.parseParams(decl.Type.Params, method.Elements[ElementGeneric])
				method.Elements[ElementResult] = a.parseResults(decl.Type.Results, method.Elements[ElementGeneric])
				methods = append(methods, method)
			}

		}
	}
	a.file.Funcs = methods
}

func (a *AstHandler) parseInterfaces(list []*ast.Field, tParams []*Element) []*Element {
	log.Printf("parse interface: method count: %d \n", len(list))
	interaceFields := make([]*Element, 0)

	for i, field := range list {
		name := ""
		if field.Names != nil {
			name = field.Names[0].Name
		}
		item := &Element{
			Name:        name,
			PackagePath: a.file.PackagePath,
			PackageName: a.file.PackageName,
			Index:       i,
			Docs:        internal.GetDocs(field.Doc),
			Comment:     internal.GetComment(field.Comment),
			Elements:    make(map[ElementType][]*Element),
		}
		switch spec := field.Type.(type) {
		case *ast.FuncType:
			item.ElementType = ElementMethod
			item.Elements[ElementParam] = a.parseParams(spec.Params, tParams)
			item.Elements[ElementResult] = a.parseParams(spec.Results, tParams)
			item.TypeString = name
		case *ast.BinaryExpr:
			item.ElementType = ElementConstrain
			vv := a.parseBinaryExpr(spec)
			item.TypeString = strings.Join(vv, "|")
			item.ElementString = strings.Join(vv, "|")
		case *ast.Ident:
			item.ElementType = ElementConstrain
			item.ElementString = spec.Name
			item.TypeString = spec.Name
		default:
			item.ElementType = ElementConstrain
			item.ElementString = "fuck!!!! here is type constraints!"
			item.TypeString = ""
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
	case *ast.IndexExpr:
		xx := a.parseBinaryExpr(expr.X)
		result = append(result, xx...)
		idx := a.parseBinaryExpr(expr.Index)
		result = append(result, idx...)
		return result
	case *ast.IndexListExpr:
		xx := a.parseBinaryExpr(expr.X)
		result = append(result, xx...)
		for _, indic := range expr.Indices {
			idx := a.parseBinaryExpr(indic)
			result = append(result, idx...)
		}
		return result
	default:
		panic("unhandled expression")
	}
	return result
}
