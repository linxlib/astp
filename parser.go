package astp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/internal/json"
	"github.com/linxlib/astp/internal/yaml"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

type setter interface {
	GetName() string
	SetType(s *Struct)
	SetInnerType(b bool)
	SetIsStruct(b bool)
	SetTypeString(s string)
	SetPointer(b bool)
	SetPrivate(b bool)
	SetSlice(b bool)
	SetPackagePath(s string)
	GetType() *Struct
}

type Parser struct {
	lock   sync.RWMutex
	Files  map[string]*File
	modDir string //mod目录
	modPkg string //mod

	//TODO
	sdkPath string //go sdk的源码根目录 eg. C:\Users\<UserName>\sdk\go1.21.0\src\builtin
	modPath string //本地mod的目录  eg. C:\Users\<UserName>\go\pkg\mod
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
		if filepath.Ext(f) == ".yaml" {
			yamlDec := yaml.NewDecoder(buf)
			_ = yamlDec.Decode(&p.Files)
		} else {
			dec := json.NewDecoder(buf)
			_ = dec.Decode(&p.Files)
		}

	}
}

func (p *Parser) WriteOut(filename string) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	var buf bytes.Buffer
	if filepath.Ext(filename) == ".yaml" {
		encoder := yaml.NewEncoder(&buf)
		err := encoder.Encode(p.Files)
		if err != nil {
			return err
		}
	} else {
		encoder := json.NewEncoder(&buf)
		err := encoder.Encode(p.Files)
		if err != nil {
			return err
		}
	}
	internal.WriteFile(filename, buf.Bytes(), true)
	return nil
}

func (p *Parser) Parse(fa ...string) {
	file := "./main.go"
	if len(fa) > 0 {
		file = fa[0]
	}
	f, key := p.parseFile(file)
	p.Files[key] = f

	if f.PackageName == "main" {
		for _, i := range f.Imports {
			if strings.HasPrefix(i.ImportPath, f.PackagePath) {
				dir := p.getPackageDir(i.ImportPath)
				files := p.parseDir(dir)
				p.merge(files)
			}
		}

	}

	for _, file := range p.Files {
		for _, s := range file.Structs {
			s.HandleCurrentPackageRefs(p.Files)
		}
	}

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
func (p *Parser) isGoFile(f string) bool {
	return filepath.Ext(f) == ".go"
}

// parseDir 解析一个目录
// 对于引用一个包的时候，直接解析其目录下的所有文件（不包含子目录）
func (p *Parser) parseDir(dir string) map[string]*File {
	files := make(map[string]*File)

	fs, _ := os.ReadDir(dir)
	for _, f := range fs {
		if !f.IsDir() && p.isGoFile(f.Name()) {
			f, key := p.parseFile(filepath.Join(dir, f.Name()))
			files[key] = f
		}
	}
	return files
}

// parseFile 解析一个go文件
func (p *Parser) parseFile(file string) (*File, string) {
	log.Println("parse:" + file)
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
	p.parseConst(f, node)
	p.parseStructs(f, node)
	p.parseFunctions(f, node)

	return f, internal.GetKey(f.PackagePath, filepath.Base(f.FilePath))
}

// parsePackages 解析包
func (p *Parser) parsePackages(file *File, af *ast.File) {
	log.Printf("parse file package: %s\n", file.PackagePath)
	file.PackageName = af.Name.Name
	file.Docs = internal.GetDocs(af.Doc)

	if af.Comments != nil {
		for _, comment := range af.Comments {
			file.Comments = append(file.Comments, internal.GetComments(comment)...)
		}
	}
}

// parseImports 解析导入区
func (p *Parser) parseImports(file *File, af *ast.File) {
	file.Imports = make([]*Import, len(af.Imports))
	for idx, spec := range af.Imports {
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
		file.Imports[idx] = i
	}
	log.Printf("parse file imports: count: %d", len(af.Imports))

}

// parseConst 解析常量区
func (p *Parser) parseConst(file *File, af *ast.File) {
	log.Printf("parse file const:%s\n", file.PackagePath)
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
func (p *Parser) findInFile(f *File, name string) setter {
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

// findFileByPackageAndType 根据引用类型查找并解析对应的代码文件
//
// @param pkg 包名
//
// @param name 类型名（可以是struct func var const）
func (p *Parser) findFileByPackageAndType(pkg string, name string) setter {
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
			if s := p.findInFile(v, name); s != nil {
				return s
			}
		}
	}
	// 如果之前未解析过，则对该目录进行目录解析
	filesa := p.parseDir(dir)
	p.merge(filesa)
	for _, v := range filesa {
		if s := p.findInFile(v, name); s != nil {
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
	log.Printf("parse file vars:%s\n", file.PackagePath)
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
								Name: v.Name,
								Docs: []string{},
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
	log.Printf("parse file structs:%s\n", file.PackagePath)
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
							a.TypeParams = p.parseTypeParams(spec.TypeParams)
						}
						switch spec1 := spec.Type.(type) {
						case *ast.StructType:
							{
								log.Printf("parse struct:%s\n", a.Name)
								a.Fields = p.parseFields(file, spec1.Fields.List)

								methods := p.parseMethods(a, af, file)
								a.Methods = append(a.Methods, methods...)
								for _, field := range a.Fields {
									if field.IsParent {
										a.HasParent = true
										// TODO: only export method
										if field.Type != nil && field.PackagePath != "this" && len(field.Type.Methods) > 0 {
											a.Methods = append(a.Methods, field.Type.Methods...)
											a.Docs = append(a.Docs, field.Type.Docs...)
										}

									}
								}

								if file.Methods == nil {
									file.Methods = make([]*Method, 0)
								}
								file.Methods = append(file.Methods, methods...)

							}

						case *ast.InterfaceType:
							log.Printf("parse interface:%s\n", a.Name)
							a.IsInterface = true
							interfaceStruct := p.parseInterfaces(spec1.Methods.List, file)
							a.Inter = interfaceStruct
						}
						file.Structs = append(file.Structs, a)

					}
				}
			}
		}
	}

	for _, s := range file.Structs {
		s.HandleCurrentPackageRef(file)
	}

}

// parseMethods 解析结构体的方法
func (p *Parser) parseMethods(s *Struct, af *ast.File, file *File) []*Method {
	log.Printf("parse methods: %s \n", s.Name)
	methods := make([]*Method, 0)
	for idx, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv != nil {
				recv := new(Receiver)
				switch decl1 := decl.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					// 只解析当前结构体的方法
					switch spec := decl1.X.(type) {
					case *ast.Ident:
						if spec.Name == s.Name {
							recv.Name = decl.Recv.List[0].Names[0].Name
							recv.Type = s.Clone()
							recv.Pointer = true
							recv.TypeString = decl1.X.(*ast.Ident).Name
						}
					case *ast.IndexExpr:
						if spec.X.(*ast.Ident).Name == s.Name {
							recv.Name = decl.Recv.List[0].Names[0].Name
							recv.Type = s.Clone()
							recv.Pointer = true
							recv.TypeString = spec.X.(*ast.Ident).Name
						}
					case *ast.IndexListExpr:
						if spec.X.(*ast.Ident).Name == s.Name {
							recv.Name = decl.Recv.List[0].Names[0].Name
							recv.Type = s.Clone()
							recv.Pointer = true
							recv.TypeString = spec.X.(*ast.Ident).Name
						}
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
						Docs:        internal.GetDocs(decl.Doc),
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
	return methods
}

// parseParams 解析参数
func (p *Parser) parseParams(file *File, params *ast.FieldList) []*ParamField {
	if params == nil {
		return nil
	}
	log.Printf("parse method params: count: %d \n", len(params.List))
	pars := make([]*ParamField, 0)
	var pIndex int

	for _, param := range params.List {
		for _, name := range param.Names {
			par := &ParamField{
				Index:       pIndex,
				Name:        name.Name,
				PackagePath: file.PackagePath,
				Private:     true,
			}
			par.Docs = internal.GetDocs(param.Doc)
			par.Comment = internal.GetComment(param.Comment)
			p.parseOther(param.Type, name.Name, file.Imports, par)
			pars = append(pars, par)
			pIndex++
		}
	}
	return pars
}

// parseResults 解析返回值
func (p *Parser) parseResults(file *File, params *ast.FieldList) []*ParamField {
	if params == nil {
		return nil
	}
	log.Printf("parse method params: count: %d \n", len(params.List))
	pars := make([]*ParamField, 0)
	var pIndex int
	for _, param := range params.List {
		if param.Names != nil {
			for _, name := range param.Names {
				par := &ParamField{
					Index:       pIndex,
					Name:        name.Name,
					PackagePath: file.PackagePath,
					Private:     true,
				}
				par.Docs = internal.GetDocs(param.Doc)
				par.Comment = internal.GetComment(param.Comment)
				//TODO: 要考虑  （a,b string） 这样的返回值形式
				p.parseOther(param.Type, name.Name, file.Imports, par)

				pars = append(pars, par)
				pIndex++
			}
		} else { //返回值可能为隐式参数

			par := &ParamField{
				Index:       pIndex,
				Name:        "",
				PackagePath: file.PackagePath,
				Type:        nil,
				Private:     true,
			}
			p.parseOther(param.Type, "", file.Imports, par)
			pars = append(pars, par)
			pIndex++
		}

	}
	return pars
}

// parseFunctions 解析函数
func (p *Parser) parseFunctions(file *File, af *ast.File) {
	log.Printf("parse functions: %s \n", file.PackagePath)
	methods := make([]*Method, 0)
	for _, decl := range af.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				method := &Method{
					PackagePath: file.PackagePath,
					Docs:        internal.GetDocs(decl.Doc),
				}
				method.Name = decl.Name.Name
				if decl.Type.TypeParams != nil {
					method.TypeParams = p.parseTypeParams(decl.Type.TypeParams)
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
func (p *Parser) parseInterfaces(af []*ast.Field, file *File) *Interface {
	log.Printf("parse interface: method count: %d \n", len(af))
	result := new(Interface)
	result.Methods = make([]*Method, 0)
	result.Constraints = []string{}
	for _, field := range af {
		switch spec := field.Type.(type) {
		case *ast.FuncType:
			method := &Method{
				PackagePath: file.PackagePath,
				Name:        field.Names[0].Name,
				Private:     internal.IsPrivate(field.Names[0].Name),
				Docs:        internal.GetDocs(field.Doc),
				Comments:    internal.GetComment(field.Comment),
				Params:      p.parseParams(file, spec.Params),
				Results:     p.parseResults(file, spec.Results),
			}

			result.Methods = append(result.Methods, method)
		default:
			p.parseInterfaceConstraints(field, result)
		}
	}

	return result
}
func (p *Parser) parseInterfaceConstraints(expr *ast.Field, p2 *Interface) {
	p2.Constraints = append(p2.Constraints, "fuck!!!! here is type constraints!")
}

func (p *Parser) parseIdent(spec *ast.Ident, name string, snamer setter) {
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
		snamer.SetPackagePath("this") //handled later
		snamer.SetType(nil)
		snamer.SetSlice(false)
	}

}
func (p *Parser) parseSelector(spec *ast.SelectorExpr, name string, imports []*Import, snamer setter) {
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

func (p *Parser) parseStar(spec *ast.StarExpr, name string, imports []*Import, snamer setter) {
	switch spec := spec.X.(type) {
	case *ast.IndexExpr:
		switch spec1 := spec.X.(type) {
		case *ast.Ident:
			snamer.SetTypeString(spec1.Name + "[" + spec.Index.(*ast.Ident).Name + "]")
			snamer.SetSlice(false)
			snamer.SetPackagePath("this")
			snamer.SetType(nil)
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(true)
		case *ast.SelectorExpr:
			switch spec2 := spec.Index.(type) {
			case *ast.Ident:
				snamer.SetTypeString(spec1.Sel.Name + "[" + spec2.Name + "]")
			case *ast.SelectorExpr:
				snamer.SetTypeString(spec1.Sel.Name + "[" + spec2.X.(*ast.Ident).Name + "." + spec2.Sel.Name + "]")

			}
			snamer.SetSlice(false)
			snamer.SetPackagePath("")
			snamer.SetPrivate(true)
			snamer.SetPointer(internal.IsPrivate(name))
			pkgName := spec1.X.(*ast.Ident).Name
			typeName := spec1.Sel.Name
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
		case *ast.StarExpr:
		}
	case *ast.IndexListExpr:
		switch spec1 := spec.X.(type) {
		case *ast.Ident:
			var ts []string
			for _, indic := range spec.Indices {
				ts = append(ts, indic.(*ast.Ident).Name)
			}
			snamer.SetTypeString(spec1.Name + "[" + strings.Join(ts, ",") + "]")
			snamer.SetSlice(false)
			snamer.SetPackagePath("this")
			snamer.SetType(nil)
			snamer.SetPrivate(internal.IsPrivate(name))
			snamer.SetPointer(true)
		case *ast.SelectorExpr:
			switch spec.X.(type) {
			case *ast.Ident:
				var ts []string
				for _, indic := range spec.Indices {
					ts = append(ts, indic.(*ast.Ident).Name)
				}
				snamer.SetTypeString(spec1.Sel.Name + "[" + strings.Join(ts, ",") + "]")
			case *ast.SelectorExpr:
				var ts []string
				for _, indic := range spec.Indices {
					switch indic.(type) {
					case *ast.Ident:
						ts = append(ts, indic.(*ast.Ident).Name)
					case *ast.SelectorExpr:
						ts = append(ts, indic.(*ast.SelectorExpr).Sel.Name)
					}
				}
				snamer.SetTypeString(spec1.Sel.Name + "[" + strings.Join(ts, ",") + "]")

			}
			snamer.SetSlice(false)
			snamer.SetPackagePath("")
			snamer.SetPrivate(true)
			snamer.SetPointer(internal.IsPrivate(name))
			pkgName := spec1.X.(*ast.Ident).Name
			typeName := spec1.Sel.Name
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
func (p *Parser) parseArrOrEll(spec any, name string, imports []*Import, snamer setter) {
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
func (p *Parser) parseMap(spec *ast.MapType, name string, imports []*Import, snamer setter) {
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
func (p *Parser) parseOther(t ast.Expr, name string, imports []*Import, snamer setter) {

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
	case *ast.IndexExpr:
		p.parseIndex(spec, name, imports, snamer)
	default:

	}

}
func (p *Parser) parseTypeParams(list *ast.FieldList) []*TypeParam {
	log.Printf("parse type param: count: %d", len(list.List))
	result := make([]*TypeParam, 0)

	for _, field := range list.List {

		for _, name := range field.Names {
			t := new(TypeParam)
			t.Name = name.Name
			switch spec := field.Type.(type) {
			case *ast.Ident:
				t.TypeName = spec.Name

			case *ast.IndexExpr:
				switch spec := spec.X.(type) {
				case *ast.Ident:
					t.TypeName = spec.Name
				case *ast.SelectorExpr:
					t.TypeName = spec.Sel.Name
				}

			case *ast.SelectorExpr:
				t.TypeName = spec.Sel.Name
			}
			t.SetPackagePath("this")

			result = append(result, t)
		}

	}
	return result
}

// parseFields 解析字段
func (p *Parser) parseFields(file *File, fields []*ast.Field) []*StructField {
	log.Printf("parse fields: count: %d\n", len(fields))
	var sf = make([]*StructField, 0)
	for idx, field := range fields {
		a := new(StructField)
		a.Index = idx
		if field.Names != nil {
			a.Name = field.Names[0].Name
		}

		p.parseOther(field.Type, a.Name, file.Imports, a)
		if a.Name == "" || a.Name == "_" {
			a.IsParent = true
			// 将继承的字段合并到当前结构
			if f := a.GetType(); f != nil {
				sf = append(sf, f.Fields...)
			}
		}

		if field.Tag != nil {
			a.Tag = reflect.StructTag(field.Tag.Value)
			a.HasTag = true
		}
		a.Comment = internal.GetComment(field.Comment)
		a.Docs = internal.GetDocs(field.Doc)

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

func (p *Parser) parseIndex(spec *ast.IndexExpr, name string, imports []*Import, snamer setter) {
	switch spec := spec.X.(type) {
	case *ast.Ident:
		p.parseIdent(spec, name, snamer)
	case *ast.SelectorExpr:
		p.parseSelector(spec, name, imports, snamer)

	}
}
