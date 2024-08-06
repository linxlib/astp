package astp

import (
	"bufio"
	"bytes"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/internal/json"
	"github.com/linxlib/astp/internal/yaml"
	"github.com/linxlib/astp/types"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

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
		encoder.SetIndent("", "  ")
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

	// 处理main包，让parser可以开始去解析其他包
	if f.IsMain() {
		for _, i := range f.Imports {
			if strings.HasPrefix(i.ImportPath, f.PackagePath) {
				dir := p.getPackageDir(i.ImportPath)
				files := p.parseDir(dir)
				p.merge(files)
			}
		}
	}
	p.handleThisPackage()

	// 当某个元素被标记为this时，表示需要当文件所在包中的所有类型都解析完毕之后，再去处理这部分元素的实际类型
	// 函数的泛型类型单独处理
	// 这里也需要考虑泛型类型是在当前包声明的情况，即先把对应的结构体上的泛型类型（标记为this）处理好之后，再去处理字段 参数 返回值等等
	// 如果一个结构体使用了匿名结构作为字段，表示使用了继承的写法
	// 则需要将父结构的字段、方法复制到当前结构，并修改其Actual（实际类型）
	// PS: 类型字段索引 方法索引等 在从父级合并当当前时，需要重新处理 使用 Element.Clone(newIndex)
	//
	// 1. handleFunctionThisPackage
	// 2. handleConstThisPackage
	// 3. handleVarThisPackage
	// 4. handleStructTypeParamThisPackage
	// 5. handleStructMethodThisPackage
	//     5.1 handleStructMethodRecvThisPackage
	//     5.2 handleStructMethodParamsThisPackage
	//     5.2 handleStructMethodResultsThisPackage
	// 6. handleActual

	//for _, file := range p.Files {
	//常量使用了iota的，到当前文件中找到结构体，作为对应结构体的枚举元素
	// 暂不考虑类型和枚举声明在不同文件的情况
	//for _, c := range file.Consts {
	//	for _, element := range file.Structs {
	//		if element.Name == c.ElementString {
	//			c.Item = element.Clone()
	//			c.ItemType = types.ElementStruct
	//			element.Elements[types.ElementConst] = append(element.Elements[types.ElementConst],
	//				&types.Element{
	//					Name:        c.Name,
	//					ElementType: types.ElementConst,
	//					Docs:        c.Docs,
	//					Comment:     c.Comment,
	//					Value:       c.Value,
	//				})
	//		}
	//	}
	//}
	//类型被标记为this的
	//for _, s := range file.Structs {
	//	// 类型的泛型类型声明在本包的
	//	for _, typeParam := range s.Elements[types.ElementGeneric] {
	//		if typeParam.PackagePath == types.PackageThisPackage {
	//			for _, element := range file.Structs {
	//				if element.Name == typeParam.ElementString {
	//					typeParam.Item = element.Clone()
	//					typeParam.PackagePath = element.PackagePath
	//				}
	//			}
	//		}
	//	}
	//	//类型的字段的类型声明在本包的
	//	for _, field := range s.Elements[types.ElementField] {
	//		if field.PackagePath == types.PackageThisPackage {
	//
	//			for _, elementS := range file.Structs {
	//				if field.ElementString == elementS.ElementString {
	//					field.Item = elementS.Clone()
	//					field.PackagePath = elementS.PackagePath
	//				}
	//			}
	//		}
	//	}
	//
	//	// 类型方法的参数和返回值类型声明在本包的
	//	for _, method := range s.Elements[types.ElementMethod] {
	//		for _, element := range method.Elements[types.ElementParam] {
	//			if element.PackagePath == types.PackageThisPackage {
	//				for _, elementS := range file.Structs {
	//					if element.ElementString == elementS.ElementString {
	//						element.Item = element.Clone()
	//						element.PackagePath = elementS.PackagePath
	//					}
	//				}
	//			}
	//		}
	//		for _, element := range method.Elements[types.ElementResult] {
	//			for _, elementS := range file.Structs {
	//				if element.ElementString == elementS.ElementString {
	//					element.Item = element.Clone()
	//					element.PackagePath = elementS.PackagePath
	//				}
	//			}
	//		}
	//	}
	//
	//}
	//}

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
	h := NewAstHandler(f, p.modPkg, node, p.findFileByPackageAndType)
	h.HandlePackages()
	h.HandleImports()
	h.HandleElements()

	return h.Result()
}

func (p *Parser) findInFile(f *File, name string) *types.Element {
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
func (p *Parser) findFileByPackageAndType(pkg string, name string) *types.Element {
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

func (p *Parser) filterFilesByPackage(pkg string) []*File {
	files := make([]*File, 0)
	for _, file := range p.Files {
		if file.PackagePath == pkg {
			files = append(files, file)
		}
	}
	return files
}

func (p *Parser) filterElementByName(files []*File, name string) *types.Element {
	for _, file := range files {
		for _, element := range file.Structs {
			if name == element.Name {
				return element
			}
		}
	}
	return nil
}

func (p *Parser) handleThisPackage() {
	for _, file := range p.Files {
		p.handleConstThisPackage(file)
		p.handleVarThisPackage(file)

	}
	p.handleFunctionThisPackage()
	p.handleStructThisPackage()

	p.handleActual()
}
func (p *Parser) handleStructThisPackage() {
	for _, file := range p.Files {
		for _, element := range file.Structs {
			files := p.filterFilesByPackage(element.PackagePath)
			for _, field := range element.Elements[types.ElementField] {
				if field.PackagePath != types.PackageThisPackage {
					continue
				}
				tmp := p.filterElementByName(files, field.TypeString)
				if tmp != nil {
					field.Item = tmp.Clone()
					field.PackagePath = tmp.PackagePath
					field.ItemType = tmp.ElementType
					field.Signature = field.GetSignature()
				}
			}
			for _, method := range element.Elements[types.ElementMethod] {
				for _, param := range method.Elements[types.ElementParam] {
					if param.PackagePath != types.PackageThisPackage {
						continue
					}
					tmp := p.filterElementByName(files, param.TypeString)
					if tmp != nil {
						param.Item = tmp.Clone()
						param.PackagePath = tmp.PackagePath
						param.ItemType = tmp.ElementType
						param.Signature = param.GetSignature()
					}
				}
				for _, param := range method.Elements[types.ElementResult] {
					if param.PackagePath != types.PackageThisPackage {
						continue
					}
					tmp := p.filterElementByName(files, param.TypeString)
					if tmp != nil {
						param.Item = tmp.Clone()
						param.PackagePath = tmp.PackagePath
						param.ItemType = tmp.ElementType
						param.Signature = param.GetSignature()
					}
				}

			}
		}
	}
}
func (p *Parser) handleFields() {
	// handle this package
	// handle type param
}
func (p *Parser) handleFunctionThisPackage() {
	for _, file := range p.Files {
		for _, element := range file.Funcs {
			files := p.filterFilesByPackage(element.PackagePath)
			for _, param := range element.Elements[types.ElementParam] {
				if param.PackagePath != types.PackageThisPackage {
					continue
				}
				tmp := p.filterElementByName(files, param.TypeString)
				if tmp != nil {
					param.Item = tmp.Clone()
					param.PackagePath = tmp.PackagePath
					param.ItemType = tmp.ElementType
					param.Signature = param.GetSignature()
				}
			}
			for _, param := range element.Elements[types.ElementResult] {
				if param.PackagePath != types.PackageThisPackage {
					continue
				}
				tmp := p.filterElementByName(files, param.TypeString)
				if tmp != nil {
					param.Item = tmp.Clone()
					param.PackagePath = tmp.PackagePath
					param.ItemType = tmp.ElementType
					param.Signature = param.GetSignature()
				}
			}

		}

	}

}
func (p *Parser) handleConstThisPackage(file *File) {
	// 对该文件的常量进行处理
	for _, element := range file.Consts {
		if element.ElementType == types.ElementConst {
			continue
		}
		files := p.filterFilesByPackage(element.PackagePath)
		tmp := p.filterElementByName(files, element.TypeString)
		if tmp != nil {
			//element.Item = tmp.Clone()
			//element.ItemType = tmp.ElementType
			if tmp.Elements == nil {
				tmp.Elements = make(map[types.ElementType][]*types.Element)
			}
			tmp.Elements[types.ElementEnum] = append(tmp.Elements[types.ElementEnum], element.Clone())
			element.Signature = element.GetSignature()
		}

	}

}
func (p *Parser) handleVarThisPackage(file *File) {
	for _, element := range file.Vars {
		if element.PackagePath != types.PackageThisPackage {
			continue
		}
		files := p.filterFilesByPackage(element.PackagePath)
		tmp := p.filterElementByName(files, element.TypeString)
		if tmp != nil {
			element.Item = tmp.Clone()
			element.ItemType = tmp.ElementType
			element.Signature = element.GetSignature()
		}
	}
}
func (p *Parser) handleReceiver() {

}
func (p *Parser) handleTypeParamThisPackage() {

}
func (p *Parser) handleStructMethodThisPackage() {
	p.handleTypeParamThisPackage()
	p.handleParamsThisPackage()
	p.handleResultsThisPackage()
}
func (p *Parser) handleStructMethodReceiverThisPackage() {}
func (p *Parser) handleParamsThisPackage()               {}
func (p *Parser) handleResultsThisPackage()              {}
func (p *Parser) handleActual() {
	log.Println("处理泛型的实际映射类型，合并继承")
	for _, file := range p.Files {
		for _, eleStruct := range file.Structs {
			if eleStruct.Name != "UserController" {
				continue
			}
			if eleStruct.FromParent {
				log.Printf("处理结构 %s \n", eleStruct.Name)
				for _, eleField := range eleStruct.Elements[types.ElementField] {
					if eleField.FromParent && eleField.Name == "" {
						log.Printf("  处理字段 %s \n", eleField.TypeString)
						//继承父级时，将当前结构中声明的实际类型拉取出来
						typeParams := eleField.Elements[types.ElementGeneric]
						log.Printf("    字段声明了 %d 个泛型参数 \n", len(typeParams))
						// 赋值它的原始类型名 比如实际类型是 int 原始是声明为 T 的
						// 方便后面方法处理的时候进行匹配
						for _, param := range typeParams {
							for _, originTypeParam := range eleField.Item.Elements[types.ElementGeneric] {
								if originTypeParam.Index == param.Index {
									param.Name = originTypeParam.Name
								}
							}
						}
						log.Printf("处理结构 %s 的字段\n", eleStruct.Name)
						eleFieldType := eleField.Item
						if eleFieldType != nil {
							for _, eleFieldTypeField := range eleFieldType.Elements[types.ElementField] {
								if !eleFieldTypeField.Private() {

									newEle := eleFieldTypeField.Clone()
									for _, e3 := range typeParams {
										// 根据泛型的索引位置来确定实际类型
										if newEle.Item != nil && newEle.Item.Name == e3.Name {
											newEle.Actual = e3.Clone()

											//e3.Name = newEle.Item.Name
										}
									}

									eleStruct.Elements[types.ElementField] = append(eleStruct.Elements[types.ElementField], newEle)
								}
							}

							for _, e2 := range eleFieldType.Elements[types.ElementMethod] {
								if !e2.Private() {
									newEle := e2.Clone()
									for _, e3 := range newEle.Elements[types.ElementParam] {
										if e3.Generic() {
											for _, e4 := range typeParams {
												if e3.Item.Name == e4.Name {
													e3.Actual = e4.Clone()
												}
											}
										}
									}

									for _, e3 := range newEle.Elements[types.ElementResult] {
										if e3.Generic() {
											for _, e4 := range typeParams {
												if e3.Item.Name == e4.Name {
													e3.Actual = e4.Clone()
												}
											}
										}
									}

									eleStruct.Elements[types.ElementMethod] = append(eleStruct.Elements[types.ElementMethod], newEle)
								}
							}
						}

					}
				}
			}

		}

	}
}
