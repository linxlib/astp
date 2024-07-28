package astp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/internal/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

type ISetTypeNamer interface {
	GetName() string
	SetType(namer *Struct)
	SetInnerType(b bool)
	SetIsStruct(b bool)
	SetTypeString(s string)
	SetPointer(b bool)
	SetPrivate(b bool)
	SetSlice(b bool)
	SetPackagePath(s string)
	GetType() *Struct
}

// 这里解析完之后的输出内容，可以为生成一个go文件，这个文件里包含了为当前这个对象赋值的语句
// 这样可以防止 如果生成文件，别人没拷贝文件 就尴尬了

type Parser struct {
	lock sync.RWMutex
	//所有相关文件 key为文件路径/包 的哈希
	// 如果是包则为 包路径+文件名 -> hash
	Files map[string]*File

	modDir string //mod目录
	modPkg string //mod

	sdkPath string //go sdk的源码根目录 eg. C:\Users\<UserName>\sdk\go1.21.0\src\builtin
	modPath string //本地mod的目录  eg. C:\Users\<UserName>\go\pkg\mod

	pkgWhiteList map[string]bool //包白名单，会在sdkPath或者modPath中找文件并解析

}

func NewParser() (p *Parser) {
	p = new(Parser)
	//从项目根目录下开始进行解析
	// 读取mod信息
	file, _ := os.Open("go.mod")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(m, "module") {
			m = strings.TrimPrefix(m, "module")
			m = strings.TrimSpace(m)
			p.modPkg = m
			break
		}
	}
	p.modDir, _ = os.Getwd()
	p.modPath = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	p.sdkPath = filepath.Join(os.Getenv("GOROOT"), "src")

	p.Files = make(map[string]*File)
	return p
}

func (p *Parser) Load(f string) {
	if internal.FileIsExist(f) {
		data := internal.ReadFile(f)

		var buf = bytes.NewBuffer(data)
		dec := json.NewDecoder(buf)
		_ = dec.Decode(&p.Files)
	}
}

func (p *Parser) WriteOut(filename string) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(p.Files)
	if err != nil {
		return err
	}
	internal.WriteFile(filename, buf.Bytes(), true)
	return nil
}

func (p *Parser) Parse() {
	f, key := p.parseFile("./main.go")
	p.Files[key] = f
}

// isParsed 判断一个文件是否已经缓存
func (p *Parser) isParsed(file string) bool {
	return false
}

// getPackageDir 根据包路径获得实际的目录路径
func (p *Parser) getPackageDir(pkgPath string) (dir string) {
	if strings.EqualFold(pkgPath, "main") { // if main return default path
		return p.modDir
	}
	if strings.HasPrefix(pkgPath, p.modPkg) {
		return p.modDir + strings.Replace(pkgPath[len(p.modPkg):], ".", "/", -1)
	}
	return ""
}

// getPackage 依据路径获得其对应的包路径
func (p *Parser) getPackage(file string) string {
	f, _ := filepath.Abs(file)
	f = strings.ReplaceAll(f, "\\", "/")
	n := strings.LastIndex(f, "/")
	f1 := strings.ReplaceAll(p.modDir, "\\", "/")
	f = f[:n]
	return p.modPkg + f[len(f1):]
}

// parseDir 解析一个目录
// 对于引用一个包的时候，直接解析其目录下的所有文件（不包含子目录）
func (p *Parser) parseDir(dir string) map[string]*File {
	files := make(map[string]*File)

	fs, _ := os.ReadDir(dir)
	for _, f := range fs {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".go" {
			f, key := p.parseFile(filepath.Join(dir, f.Name()))
			files[key] = f
		}

	}

	return files
}

// parseFile 解析一个go文件
func (p *Parser) parseFile(file string) (*File, string) {
	name := filepath.Base(file)
	node, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	f := &File{
		Name:        name,
		PackagePath: p.getPackage(file),
		FilePath:    internal.Abs(file),
	}
	p.parsePackages(f, node)
	p.parseImports(f, node)
	p.parseVars(f, node)
	p.parseConsts(f, node)
	p.parseStructs(f, node)

	p.parseFunction(f, node)

	return f, internal.GetKey(f.PackagePath, filepath.Base(f.FilePath))
}

// parsePackages 解析包
func (p *Parser) parsePackages(file *File, af *ast.File) {
	file.PackageName = af.Name.Name
	file.Docs = internal.GetDocs(af.Doc)
	if af.Comments != nil {
		for _, comment := range af.Comments {
			if comment.List != nil {
				file.Comments = append(file.Comments, strings.TrimLeft(comment.Text(), "//"))
			}

		}
	}
}

// parseImports 解析导入区
func (p *Parser) parseImports(file *File, af *ast.File) {
	file.Imports = make([]*Import, 0)
	for _, spec := range af.Imports {
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
		file.Imports = append(file.Imports, i)
	}

}

// parseConsts 解析常量区
func (p *Parser) parseConsts(file *File, af *ast.File) {
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.CONST {
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &Const{
								Name:       v.Name,
								TypeString: "",
								Type:       nil,
								Value:      nil,
								Docs:       []string{},
								Comments:   "",
							}
							if a, ok := spec.Type.(*ast.Ident); ok {
								vv.TypeString = a.Name
							}
							if len(spec.Values) == len(spec.Names) {
								if a, ok := spec.Values[i].(*ast.Ident); ok {
									//todo: 对于const枚举需要分区处理
									if strings.Contains(a.Name, "iota") {
										vv.IsIota = true
									}
								}
							}

							vv.Docs = internal.GetDocs(spec.Doc)
							vv.Comments = internal.GetComment(spec.Comment)

							file.Consts = append(file.Consts, vv)
						}
					}
				}
			}
		}
	}
}

// findFileByPackageAndType 根据引用类型查找并解析对应的代码文件
//
// @param pkg 包名
// @param name 类型名（可以是struct func var const）
func (p *Parser) findFileByPackageAndType(pkg string, name string) ISetTypeNamer {
	var findInFile = func(f *File) ISetTypeNamer {
		for _, s := range f.Structs {
			if s.Name == name {
				return s
			}
		}
		for _, s := range f.Methods {
			if s.Name == name {
				return s
			}
		}
		for _, s := range f.Funcs {
			if s.Name == name {
				return s
			}
		}
		for _, s := range f.Vars {
			if s.Name == name {
				return s
			}
		}
		for _, s := range f.Consts {
			if s.Name == name {
				return s
			}
		}
		return nil
	}

	//http.Header{}
	//TODO: net/http 等包需要加入支持
	dir := p.getPackageDir(pkg)
	if dir == "" {
		return nil
	}
	files, _ := os.ReadDir(dir)
	//在一个package对应的目录下，遍历所有文件 找到对应的文件
	for _, file := range files {
		b := filepath.Base(file.Name())
		key := internal.GetKey(pkg, b)

		if v, ok := p.Files[key]; ok { //根据文件key，找到了对应的文件（已在其他地方解析过）
			if s := findInFile(v); s != nil {
				return s
			}
		}
	}
	// 如果之前未解析过，则对该目录进行目录解析
	filesa := p.parseDir(dir)
	p.merge(filesa)
	for _, v := range filesa {
		if s := findInFile(v); s != nil {
			return s
		}
	}

	return nil
}
func (p *Parser) merge(files map[string]*File) {
	for key, file := range files {
		if _, ok := p.Files[key]; !ok {
			p.Files[key] = file
		}
	}
}

// parseVars 解析变量
func (p *Parser) parseVars(file *File, af *ast.File) {
	file.Vars = make([]*Var, 0)
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.VAR:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						for i, v := range spec.Names {
							vv := &Var{
								Name:       v.Name,
								TypeString: "",
								Type:       nil,
								Value:      nil,
								Docs:       []string{},
								Comments:   "",
							}

							p.parseOther(spec.Type, v.Name, file.Imports, vv)

							//TODO: 解析变量的值
							if len(spec.Values) == len(spec.Names) {
								if a, ok := spec.Values[i].(*ast.BasicLit); ok {
									vv.Value = a.Value
								}
							}
							vv.Docs = internal.GetDocs(spec.Doc)
							vv.Comments = internal.GetComment(spec.Comment)

							file.Vars = append(file.Vars, vv)
						}
					}
				}
			default:
				continue
			}
		}
	}
}

// parseStructs 解析结构体
func (p *Parser) parseStructs(file *File, af *ast.File) {
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						a := &Struct{
							Name:        spec.Name.Name,
							PackagePath: file.PackagePath,
						}
						if spec.Doc == nil {
							a.Docs = internal.GetDocs(decl.Doc)
						} else {
							a.Docs = internal.GetDocs(spec.Doc)

						}
						a.Comment = internal.GetComment(spec.Comment)
						if spec.TypeParams != nil {
							a.IsGeneric = true
							a.TypeParams = p.parseTypeParams(file, spec.TypeParams)
						}
						switch spec1 := spec.Type.(type) {
						case *ast.StructType:
							{
								a.Fields = p.parseFields(file, spec1.Fields.List)
								methods := p.parseMethods(a, af, file)
								if file.Methods == nil {
									file.Methods = make([]*Method, 0)
								}
								file.Methods = append(file.Methods, methods...)

							}

						case *ast.InterfaceType:
							a.IsInterface = true

							interfaceStruct := p.parseInterfaces(a, spec1.Methods.List, file)
							a.Inter = interfaceStruct
							//if file.Methods == nil {
							//	file.Methods = make([]*Method, 0)
							//}
							//file.Methods = append(file.Methods, interfaceStruct.Methods...)
						}
						file.Structs = append(file.Structs, a)

					}
				}
			}
		}
	}
	for _, s := range file.Structs {
		s.SetThisPackageTypeParams(file)
		s.SetThisPackageFields(file)
		s.SetThisPackageMethodParams(file)
	}

}

// 解析结构体的方法
func (p *Parser) parseMethods(s *Struct, af *ast.File, file *File) []*Method {
	methods := make([]*Method, 0)
	for idx, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv != nil {

				recv := new(Receiver)
				switch decl1 := decl.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					// 只解析当前结构体的方法
					if decl1.X.(*ast.Ident).Name == s.Name {
						recv.Name = decl.Recv.List[0].Names[0].Name
						recv.Type = s.Clone()
						recv.Pointer = true
						recv.TypeString = decl1.X.(*ast.Ident).Name
					}
				case *ast.Ident:
					// 只解析当前结构体的方法
					if decl1.Name == s.Name {
						recv.Name = decl.Recv.List[0].Names[0].Name
						recv.Type = s.Clone()
						recv.Pointer = false
						recv.TypeString = decl1.Name
					}
				}
				if recv.Type != nil && recv.Type.Name == s.Name {
					method := &Method{
						Receiver:    recv,
						Index:       idx,
						PackagePath: file.PackagePath,
						Name:        decl.Name.Name,
						Private:     internal.IsPrivate(decl.Name.Name),
						Signature:   "",
						Docs:        internal.GetDocs(decl.Doc),
						Params:      nil,
						Results:     nil,
					}
					// 解析参数
					method.Params = p.parseParams(file, decl.Type.Params)
					// 解析返回
					method.Results = p.parseResults(file, decl.Type.Results)

					methods = append(methods, method)
				}

			}

		}
	}
	s.Methods = methods
	return methods
}

// parseParams 解析参数
func (p *Parser) parseParams(file *File, params *ast.FieldList) []*ParamField {
	if params == nil {
		return nil
	}
	pars := make([]*ParamField, len(params.List))

	for index, param := range params.List {
		for idx, name := range param.Names {
			par := &ParamField{
				Index:       idx,
				Name:        name.Name,
				PackagePath: file.PackagePath,
				Type:        nil,
				HasTag:      false,
				Tag:         "",
				TypeString:  "",
				InnerType:   false,
				Private:     true,
				Pointer:     false,
				Slice:       false,
				IsStruct:    false,
				Docs:        nil,
				Comment:     "",
			}
			par.Docs = internal.GetDocs(param.Doc)
			par.Comment = internal.GetComment(param.Comment)
			//TODO: 要考虑  （a,b string） 这样的参数形式
			p.parseOther(param.Type, name.Name, file.Imports, par)

			pars[index] = par

		}
	}
	return pars
}

// parseResults 解析返回值
func (p *Parser) parseResults(file *File, params *ast.FieldList) []*ParamField {
	if params == nil {
		return nil
	}
	pars := make([]*ParamField, len(params.List))

	for index, param := range params.List {
		if param.Names != nil {
			for idx, name := range param.Names {
				par := &ParamField{
					Index:       idx,
					Name:        name.Name,
					PackagePath: file.PackagePath,
					Type:        nil,
					HasTag:      false,
					Tag:         "",
					TypeString:  "",
					InnerType:   false,
					Private:     true,
					Pointer:     false,
					Slice:       false,
					IsStruct:    false,
					Docs:        nil,
					Comment:     "",
				}
				par.Docs = internal.GetDocs(param.Doc)
				par.Comment = internal.GetComment(param.Comment)
				//TODO: 要考虑  （a,b string） 这样的返回值形式
				p.parseOther(param.Type, name.Name, file.Imports, par)

				pars[index] = par

			}
		} else { //返回值可能为隐式参数
			par := &ParamField{
				Index:       0,
				Name:        "",
				PackagePath: "",
				Type:        nil,
				HasTag:      false,
				Tag:         "",
				TypeString:  "",
				InnerType:   false,
				Private:     true,
				Pointer:     false,
				Slice:       false,
				IsStruct:    false,
				Docs:        nil,
				Comment:     "",
			}
			p.parseOther(param.Type, "", file.Imports, par)
			pars[index] = par
		}

	}
	return pars
}

// parseFunction 解析函数
func (p *Parser) parseFunction(file *File, af *ast.File) {
	methods := make([]*Method, 0)
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				method := &Method{
					Receiver:    nil,
					PackagePath: file.PackagePath,
					Name:        "",
					Private:     false,
					Signature:   "",
					Docs:        internal.GetDocs(decl.Doc),
					Params:      nil,
					Results:     nil,
				}
				method.Name = decl.Name.Name
				if decl.Type.TypeParams != nil {
					method.TypeParams = p.parseTypeParams(file, decl.Type.TypeParams)
				}
				method.Private = internal.IsPrivate(method.Name)
				method.Params = p.parseParams(file, decl.Type.Params)

				method.Results = p.parseResults(file, decl.Type.Results)
				methods = append(methods, method)
			}

		}
	}
	file.Funcs = methods
}

// parseInterfaces 解析接口
func (p *Parser) parseInterfaces(s *Struct, af []*ast.Field, file *File) *Interface {
	result := new(Interface)
	result.Methods = make([]*Method, 0)
	result.Constraints = []string{}
	for _, field := range af {
		switch spec := field.Type.(type) {
		case *ast.FuncType:
			method := &Method{
				Receiver:    nil,
				PackagePath: file.PackagePath,
				Name:        field.Names[0].Name,
				Private:     internal.IsPrivate(field.Names[0].Name),
				Signature:   "",
				Docs:        internal.GetDocs(field.Doc),
				Comments:    internal.GetComment(field.Comment),
				Params:      p.parseParams(file, spec.Params),
				Results:     p.parseResults(file, spec.Results),
			}

			result.Methods = append(result.Methods, method)
		default:
			p.parseInterfaceContraints(field, result)
		}
	}

	return result
}
func (p *Parser) parseInterfaceContraints(expr *ast.Field, p2 *Interface) {
	p2.Constraints = append(p2.Constraints, "fuck!!!! here is type constraints!")
}

func (p *Parser) parseIdent(spec *ast.Ident, name string, snamer ISetTypeNamer) {
	if internal.IsInternalType(spec.Name) {
		snamer.SetIsStruct(false)
		snamer.SetInnerType(true)
		snamer.SetTypeString(spec.Name)
		snamer.SetPrivate(internal.IsPrivate(name))
		snamer.SetSlice(false)
		snamer.SetType(nil)
	} else {
		snamer.SetIsStruct(true)
		snamer.SetInnerType(false)
		snamer.SetTypeString(spec.Name)
		snamer.SetPrivate(internal.IsPrivate(name))
		snamer.SetPackagePath("this") //标记为this 在整个文件都解析完的情况下再去处理
		snamer.SetType(nil)
		snamer.SetSlice(false)
	}

}
func (p *Parser) parseSelector(spec *ast.SelectorExpr, name string, imports []*Import, snamer ISetTypeNamer) {
	snamer.SetIsStruct(true)
	snamer.SetInnerType(false)
	pkgName := spec.X.(*ast.Ident).Name
	typeName := spec.Sel.Name
	snamer.SetTypeString(pkgName + "." + typeName)
	for _, i3 := range imports {
		if i3.Name == pkgName {
			snamer.SetPackagePath(i3.ImportPath)
			namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
			if namer != nil {
				snamer.SetType(namer.GetType())
			} else {
				snamer.SetInnerType(true)
			}

		}
	}
	snamer.SetPrivate(internal.IsPrivate(name))
	snamer.SetPointer(false)
	snamer.SetSlice(false)
}
func (p *Parser) parseStar(spec *ast.StarExpr, name string, imports []*Import, snamer ISetTypeNamer) {
	switch spec := spec.X.(type) {
	case *ast.IndexExpr:
		switch spec1 := spec.X.(type) {
		case *ast.Ident:
			//fmt.Println(spec1.Name)
			snamer.SetTypeString(spec1.Name + "[" + spec.Index.(*ast.Ident).Name + "]")
			snamer.SetSlice(false)
			snamer.SetPackagePath("this")
			snamer.SetType(nil)
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(true)
		case *ast.SelectorExpr:
		case *ast.StarExpr:
		}
	case *ast.Ident: //指针的内置类型
		if internal.IsInternalType(spec.Name) {
			snamer.SetIsStruct(false)
			snamer.SetInnerType(true)
			snamer.SetType(nil)
			snamer.SetTypeString("*" + spec.Name)
			snamer.SetPointer(true)
			snamer.SetSlice(false)
		} else {

			snamer.SetIsStruct(true)
			snamer.SetInnerType(false)
			snamer.SetTypeString("*" + spec.Name)
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPackagePath("this")
			snamer.SetType(nil)
			snamer.SetSlice(false)
			snamer.SetPointer(true)
		}

	case *ast.SelectorExpr: //指针 带包类型
		pkgName := spec.X.(*ast.Ident).Name
		typeName := spec.Sel.Name
		snamer.SetPointer(true)
		snamer.SetInnerType(false)
		snamer.SetIsStruct(false)
		snamer.SetIsStruct(true)
		snamer.SetSlice(false)
		for _, i3 := range imports {
			if i3.Name == pkgName {
				snamer.SetPackagePath(i3.ImportPath)
				namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
				if namer != nil {
					snamer.SetType(namer.GetType())
				} else {
					snamer.SetInnerType(true)
				}

				snamer.SetTypeString("*" + pkgName + "." + typeName)
			}
		}
	case *ast.ArrayType: //指针的数组
		switch spec := spec.Elt.(type) {
		case *ast.Ident: //指针的数组 内置类型
			snamer.SetIsStruct(false)
			snamer.SetInnerType(true)
			snamer.SetTypeString("*[]" + spec.Name)
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(true)
			snamer.SetType(nil)
			snamer.SetSlice(true)
		case *ast.SelectorExpr: //指针的数组 带包类型
			snamer.SetIsStruct(true)
			snamer.SetInnerType(false)
			snamer.SetPrivate(internal.IsPrivate(name))
			pkgName := spec.X.(*ast.Ident).Name
			typeName := spec.Sel.Name
			snamer.SetTypeString("*[]" + pkgName + "." + typeName)
			for _, i3 := range imports {
				if i3.Name == pkgName {
					snamer.SetPackagePath(i3.ImportPath)
					namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
					if namer != nil {
						snamer.SetType(namer.GetType())
					} else {
						snamer.SetInnerType(true)
					}

				}
			}
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(false)
			snamer.SetSlice(true)

		case *ast.StarExpr: //指针的数组 指针的类型
			switch spec := spec.X.(type) {
			case *ast.Ident: //指针的数组 指针的内置类型
				snamer.SetIsStruct(false)
				snamer.SetInnerType(true)
				snamer.SetType(nil)

				snamer.SetSlice(true)
				snamer.SetTypeString("*[]*" + spec.Name)
				snamer.SetPointer(true)
			case *ast.SelectorExpr: //指针的数组 指针的带包类型
				pkgName := spec.X.(*ast.Ident).Name
				typeName := spec.Sel.Name
				snamer.SetPointer(true)
				snamer.SetInnerType(false)
				snamer.SetIsStruct(false)
				snamer.SetIsStruct(true)
				snamer.SetSlice(true)
				for _, i3 := range imports {
					if i3.Name == pkgName {
						snamer.SetPackagePath(i3.ImportPath)
						namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
						if namer != nil {
							snamer.SetType(namer.GetType())
						} else {
							snamer.SetInnerType(true)
						}

						snamer.SetTypeString("*[]*" + pkgName + "." + typeName)
					}
				}
			}
		}

	}
}
func (p *Parser) parseArrOrEll(spec any, name string, imports []*Import, snamer ISetTypeNamer) {
	switch spec := spec.(type) {
	case *ast.ArrayType:
		switch spec := spec.Elt.(type) {
		case *ast.Ident: //数组的内置类型
			snamer.SetIsStruct(false)
			snamer.SetInnerType(true)
			snamer.SetTypeString("[]" + spec.Name)
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(true)
			snamer.SetType(nil)
			snamer.SetSlice(true)
		case *ast.SelectorExpr: //数组的带包类型
			snamer.SetIsStruct(true)
			snamer.SetInnerType(false)
			snamer.SetPrivate(internal.IsPrivate(name))
			pkgName := spec.X.(*ast.Ident).Name
			typeName := spec.Sel.Name
			snamer.SetTypeString("[]" + pkgName + "." + typeName)
			for _, i3 := range imports {
				if i3.Name == pkgName {
					snamer.SetPackagePath(i3.ImportPath)
					namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
					if namer != nil {
						snamer.SetType(namer.GetType())
					} else {
						snamer.SetInnerType(true)
					}

				}
			}
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(false)
			snamer.SetSlice(true)

		case *ast.StarExpr: //数组的指针
			switch spec := spec.X.(type) {
			case *ast.Ident: //数组的指针内置类型
				snamer.SetIsStruct(false)
				snamer.SetInnerType(true)
				snamer.SetTypeString("[]*" + spec.Name)
				snamer.SetPointer(true)
				snamer.SetType(nil)
				snamer.SetSlice(true)
			case *ast.SelectorExpr: //数组的指针带包类型
				pkgName := spec.X.(*ast.Ident).Name
				typeName := spec.Sel.Name
				snamer.SetPointer(true)
				snamer.SetInnerType(false)
				snamer.SetIsStruct(false)
				snamer.SetIsStruct(true)
				snamer.SetSlice(true)
				for _, i3 := range imports {
					if i3.Name == pkgName {
						snamer.SetPackagePath(i3.ImportPath)
						namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
						if namer != nil {
							snamer.SetType(namer.GetType())
						} else {
							snamer.SetInnerType(true)
						}

						snamer.SetTypeString("[]*" + pkgName + "." + typeName)
					}
				}
			}
		}
	case *ast.Ellipsis:
		switch spec := spec.Elt.(type) {
		case *ast.Ident: //数组的内置类型
			snamer.SetIsStruct(false)
			snamer.SetInnerType(true)
			snamer.SetTypeString("[]" + spec.Name)
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(true)
			snamer.SetType(nil)
			snamer.SetSlice(true)
		case *ast.SelectorExpr: //数组的带包类型
			snamer.SetIsStruct(true)
			snamer.SetInnerType(false)
			snamer.SetPrivate(internal.IsPrivate(name))
			pkgName := spec.X.(*ast.Ident).Name
			typeName := spec.Sel.Name
			snamer.SetTypeString("[]" + pkgName + "." + typeName)
			for _, i3 := range imports {
				if i3.Name == pkgName {
					snamer.SetPackagePath(i3.ImportPath)
					namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
					if namer != nil {
						snamer.SetType(namer.GetType())
					} else {
						snamer.SetInnerType(true)
					}

				}
			}
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(false)
			snamer.SetSlice(true)

		case *ast.StarExpr: //数组的指针
			switch spec := spec.X.(type) {
			case *ast.Ident: //数组的指针内置类型
				snamer.SetIsStruct(false)
				snamer.SetInnerType(true)
				snamer.SetTypeString("[]*" + spec.Name)
				snamer.SetPointer(true)
				snamer.SetType(nil)
				snamer.SetSlice(true)
			case *ast.SelectorExpr: //数组的指针带包类型
				pkgName := spec.X.(*ast.Ident).Name
				typeName := spec.Sel.Name
				snamer.SetPointer(true)
				snamer.SetInnerType(false)
				snamer.SetIsStruct(false)
				snamer.SetIsStruct(true)
				snamer.SetSlice(true)
				for _, i3 := range imports {
					if i3.Name == pkgName {
						snamer.SetPackagePath(i3.ImportPath)
						namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
						if namer != nil {
							snamer.SetType(namer.GetType())
						} else {
							snamer.SetInnerType(true)
						}

						snamer.SetTypeString("[]*" + pkgName + "." + typeName)
					}
				}
			}
		}
	}

}
func (p *Parser) parseMap(spec *ast.MapType, name string, imports []*Import, snamer ISetTypeNamer) {
	snamer.SetPrivate(internal.IsPrivate(name))
	snamer.SetIsStruct(false)
	snamer.SetInnerType(true)
	snamer.SetSlice(false)

	switch spec1 := spec.Value.(type) {
	case *ast.Ident: //map[string]string
		snamer.SetTypeString(fmt.Sprintf("map[%s]%s", spec.Key.(*ast.Ident).Name, spec1.Name))
		snamer.SetPointer(false)
		snamer.SetPackagePath("builtin")
	case *ast.SelectorExpr: //map[string]http.Header

		pkgName := spec1.X.(*ast.Ident).Name
		typeName := spec1.Sel.Name
		snamer.SetTypeString(fmt.Sprintf("map[%s]%s", spec.Key.(*ast.Ident).Name, pkgName+"."+typeName))
		for _, i3 := range imports {
			if i3.Name == pkgName {
				snamer.SetPackagePath(i3.ImportPath)
				namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
				if namer != nil {
					snamer.SetType(namer.GetType())
				} else {
					snamer.SetInnerType(true)
				}

			}
		}
	case *ast.StarExpr: //map[string]*http.Header
		switch spec2 := spec1.X.(type) {
		case *ast.Ident: //数组的指针内置类型
			snamer.SetIsStruct(false)
			snamer.SetInnerType(true)
			snamer.SetTypeString(fmt.Sprintf("map[%s]*%s", spec.Key.(*ast.Ident).Name, spec2.Name))
			snamer.SetPointer(true)
			snamer.SetType(nil)
			snamer.SetSlice(true)
		case *ast.SelectorExpr: //数组的指针带包类型
			pkgName := spec2.X.(*ast.Ident).Name
			typeName := spec2.Sel.Name
			snamer.SetPointer(true)
			snamer.SetInnerType(false)
			snamer.SetIsStruct(false)
			snamer.SetIsStruct(true)
			snamer.SetSlice(true)
			for _, i3 := range imports {
				if i3.Name == pkgName {
					snamer.SetPackagePath(i3.ImportPath)
					namer := p.findFileByPackageAndType(i3.ImportPath, typeName)
					if namer != nil {
						snamer.SetType(namer.GetType())
					} else {
						snamer.SetInnerType(true)
					}

					snamer.SetTypeString("map[" + spec.Key.(*ast.Ident).Name + "]*" + pkgName + "." + typeName)
				}
			}
		}
	}
}

// parseOther 解析引入的类型
func (p *Parser) parseOther(t ast.Expr, name string, imports []*Import, snamer ISetTypeNamer) {

	snamer.SetPrivate(internal.IsPrivate(name))
	snamer.SetPackagePath("builtin")
	switch spec := t.(type) {
	case *ast.Ident: //直接一个类型
		p.parseIdent(spec, name, snamer)
	case *ast.SelectorExpr: //带包的类型
		p.parseSelector(spec, name, imports, snamer)
	case *ast.StarExpr: //指针
		p.parseStar(spec, name, imports, snamer)
	case *ast.ArrayType: //数组
		p.parseArrOrEll(spec, name, imports, snamer)
	case *ast.Ellipsis: // ...
		p.parseArrOrEll(spec, name, imports, snamer)
	case *ast.MapType:
		p.parseMap(spec, name, imports, snamer)
	default:

	}

}
func (p *Parser) parseTypeParams(file *File, list *ast.FieldList) []*TypeParam {
	result := make([]*TypeParam, 0)

	for _, field := range list.List {

		for _, name := range field.Names {
			t := new(TypeParam)
			t.Name = name.Name
			t.TypeName = field.Type.(*ast.Ident).Name
			t.SetPackagePath("this")

			result = append(result, t)
		}

	}
	return result
}

// parseFields 解析字段
func (p *Parser) parseFields(file *File, fields []*ast.Field) []*StructField {
	var sf = make([]*StructField, 0)
	for idx, field := range fields {
		a := new(StructField)
		a.Index = idx
		if field.Names != nil {
			a.Name = field.Names[0].Name
		}

		p.parseOther(field.Type, a.Name, file.Imports, a)
		if a.Name == "" {
			// 将继承的字段合并到当前结构
			if f := a.GetType(); f != nil {
				sf = append(sf, f.Fields...)
			}
		}

		if field.Tag != nil {
			a.Tag = reflect.StructTag(field.Tag.Value)
			a.HasTag = true
		}
		if field.Comment != nil {
			a.Comment = internal.GetComment(field.Comment)
		}
		if field.Doc != nil {
			a.Docs = internal.GetDocs(field.Doc)
		}

		sf = append(sf, a)
	}
	return sf
}

func (p *Parser) VisitAllStructs(name string, f func(s *Struct) bool) {
	for _, file := range p.Files {
		for _, s := range file.Structs {
			if s.Name == name {
				if f(s) == true {
					return
				}
			}
		}

	}
}
