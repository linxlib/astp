package astp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/internal/json"
	"github.com/linxlib/astp/internal/yaml"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type Parser struct {
	lock   sync.RWMutex
	Files  map[string]*File
	modDir string //mod目录
	modPkg string //mod

	//TODO: 未来可能实现去对应的源码目录里进行解析
	// 但由于一般第三方包的源码引用繁多，可能比较麻烦
	sdkPath string //go sdk的源码根目录 eg. C:\Users\<UserName>\sdk\go1.21.0\src\builtin
	modPath string //本地mod的目录  eg. C:\Users\<UserName>\go\pkg\mod

	parseFunctions bool
	ignorePkgs     map[string]bool

	filtered map[string]*Element
}

func (p *Parser) SetParseFunctions(b bool) *Parser {
	p.parseFunctions = b
	return p
}
func (p *Parser) AddIgnorePkg(pkg string) *Parser {
	p.ignorePkgs[pkg] = true
	return p
}
func (p *Parser) IgnorePkg(pkg string) bool {
	return p.ignorePkgs[pkg]
}
func NewParser() (p *Parser) {
	p = new(Parser)
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
	p.ignorePkgs = make(map[string]bool)
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
			// 当前项目的子包
			if strings.HasPrefix(i.ImportPath, f.PackagePath) {
				dir := p.getPackageDir(i.ImportPath)
				files := p.parseDir(dir)
				p.merge(files)
			}
		}
	}
	p.filtered = make(map[string]*Element)
	p.handleThisPackage()

	// 当某个元素被标记为this时，表示需要当文件所在包中的所有类型都解析完毕之后，再去处理这部分元素的实际类型
	// 函数的泛型类型单独处理
	// 这里也需要考虑泛型类型是在当前包声明的情况，即先把对应的结构体上的泛型类型（标记为this）处理好之后，再去处理字段 参数 返回值等等
	// 如果一个结构体使用了匿名结构作为字段，表示使用了继承的写法
	// 则需要将父结构的字段、方法复制到当前结构，并修改其Actual（实际类型）
	// PS: 类型字段索引 方法索引等 在从父级合并当当前时，需要重新处理 使用 Element.Clone(newIndex)

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
			if key != "" {
				files[key] = f
			}

		}
	}
	return files
}

// parseFile 解析一个go文件
func (p *Parser) parseFile(file string) (*File, string) {
	log.Println("parse:" + file)
	if p.IgnorePkg(p.getPackage(file)) {
		log.Println("ignore:" + file)
		return nil, ""
	}
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
	h.parseFunctions = p.parseFunctions
	h.HandlePackages()
	h.HandleImports()
	h.HandleElements()

	return h.Result()
}

func (p *Parser) findInFile(f *File, name string) *Element {
	// TODO: 这里可能只需要查找结构体，其他的暂时还用不到
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
func (p *Parser) findFileByPackageAndType(pkg string, name string) *Element {
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

// filterFilesByPackage 根据包路径，返回在该包路径下的解析过的文件列表
func (p *Parser) filterFilesByPackage(pkg string) []*File {
	files := make([]*File, 0)
	for _, file := range p.Files {
		if file.PackagePath == pkg {
			files = append(files, file)
		}
	}
	return files
}

// filterElementByName 在包下的所有文件中找到名称为name的结构
func (p *Parser) filterElementByName(files []*File, name string) *Element {
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
		p.handleFunctionThisPackage(file)
		p.handleStructThisPackage(file)

	}
	for _, file := range p.Files {
		p.handleActual(file)
	}
	for _, file := range p.Files {
		p.handleActual2(file)
	}
	for _, file := range p.Files {
		p.handleActual3(file)
	}

}
func (p *Parser) handleStructThisPackage(file *File) {

	for _, element := range file.Structs {
		log.Printf("处理结构体: %s \n", element.Name)
		files := p.filterFilesByPackage(element.PackagePath)

		log.Printf("  处理结构体字段: %s \n", element.Name)
		for _, field := range element.Elements[ElementField] {
			if field.PackagePath != PackageThisPackage {
				continue
			}
			findTypeName := field.TypeString
			if field.IsItemSlice || strings.HasPrefix(field.TypeString, "[]") {
				findTypeName = strings.TrimPrefix(field.TypeString, "[]")
			}
			var tmp *Element
			if tmp1, ok := p.filtered[findTypeName]; ok {
				tmp = tmp1
			} else {
				tmp = p.filterElementByName(files, findTypeName)
				if tmp == nil {
					continue
				}
				p.filtered[findTypeName] = tmp
			}

			//TODO: 这个字段对应的结构有可能还未处理过
			filesa := p.filterFilesByPackage(field.PackagePath)
			for _, f := range filesa {
				p.handleStructThisPackage(f)
			}

			field.Item = tmp.Clone()
			field.ElementType = tmp.ElementType
			field.PackagePath = tmp.PackagePath
			field.ItemType = tmp.ElementType

		}

		log.Printf("  处理结构体方法: %s \n", element.Name)
		for _, method := range element.Elements[ElementMethod] {
			log.Printf("    处理方法参数: %s \n", method.Name)
			for _, param := range method.Elements[ElementParam] {
				if param.PackagePath != PackageThisPackage {
					continue
				}
				findTypeName := param.TypeString
				if param.IsItemSlice || strings.HasPrefix(param.TypeString, "[]") {
					findTypeName = strings.TrimPrefix(param.TypeString, "[]")
				}
				var tmp *Element
				if tmp1, ok := p.filtered[findTypeName]; ok {
					tmp = tmp1
				} else {
					tmp = p.filterElementByName(files, findTypeName)
					if tmp == nil {
						continue
					}
					p.filtered[findTypeName] = tmp
				}
				param.Item = tmp.Clone()
				param.Docs = tmp.Docs
				param.Comment = tmp.Comment
				param.PackagePath = tmp.PackagePath
				param.ItemType = tmp.ElementType
			}
			for _, param := range method.Elements[ElementParam] {
				if param.Item == nil {
					continue
				}
				for _, field := range param.Item.Elements[ElementField] {
					if field.PackagePath != PackageThisPackage {
						continue
					}
					findTypeName := field.TypeString
					if field.IsItemSlice || strings.HasPrefix(field.TypeString, "[]") {
						findTypeName = strings.TrimPrefix(field.TypeString, "[]")
					}
					var tmp *Element
					if tmp1, ok := p.filtered[findTypeName]; ok {
						tmp = tmp1
					} else {
						tmp = p.filterElementByName(files, findTypeName)
						if tmp == nil {
							continue
						}
						p.filtered[findTypeName] = tmp
					}
					field.Item = tmp.Clone()
					field.Docs = tmp.Docs
					field.Comment = tmp.Comment
					field.PackagePath = tmp.PackagePath
					field.ItemType = tmp.ElementType
				}
			}

			log.Printf("    处理方法返回值: %s \n", method.Name)
			for _, param := range method.Elements[ElementResult] {
				if param.PackagePath != PackageThisPackage {
					continue
				}
				findTypeName := param.TypeString
				if param.IsItemSlice || strings.HasPrefix(param.TypeString, "[]") {
					findTypeName = strings.TrimPrefix(param.TypeString, "[]")
				}
				var tmp *Element
				if tmp1, ok := p.filtered[findTypeName]; ok {
					tmp = tmp1
				} else {
					tmp = p.filterElementByName(files, findTypeName)
					if tmp == nil {
						continue
					}
					p.filtered[findTypeName] = tmp
				}
				param.Item = tmp.Clone()
				param.PackagePath = tmp.PackagePath
				param.ItemType = tmp.ElementType

			}
			for _, param := range method.Elements[ElementResult] {
				if param.Item == nil {
					continue
				}
				for _, field := range param.Item.Elements[ElementField] {
					if field.PackagePath != PackageThisPackage {
						continue
					}
					findTypeName := field.TypeString
					if field.IsItemSlice || strings.HasPrefix(field.TypeString, "[]") {
						findTypeName = strings.TrimPrefix(field.TypeString, "[]")
					}
					var tmp *Element
					if tmp1, ok := p.filtered[findTypeName]; ok {
						tmp = tmp1
					} else {
						tmp = p.filterElementByName(files, findTypeName)
						if tmp == nil {
							continue
						}
						p.filtered[findTypeName] = tmp
					}
					field.Item = tmp.Clone()
					field.Docs = tmp.Docs
					field.Comment = tmp.Comment
					field.PackagePath = tmp.PackagePath
					field.ItemType = tmp.ElementType
				}
			}

		}
	}

}

func (p *Parser) handleFunctionThisPackage(file *File) {

	for _, element := range file.Funcs {
		log.Printf("处理函数： %s \n", element.Name)
		files := p.filterFilesByPackage(element.PackagePath)
		log.Printf("  处理函数参数： %s \n", element.Name)
		for _, param := range element.Elements[ElementParam] {
			if param.PackagePath != PackageThisPackage {
				continue
			}
			tmp := p.filterElementByName(files, param.TypeString)
			if tmp == nil {
				continue
			}
			param.Item = tmp
			param.PackagePath = tmp.PackagePath
			param.ItemType = tmp.ElementType

		}
		log.Printf("  处理函数返回值： %s \n", element.Name)
		for _, param := range element.Elements[ElementResult] {
			if param.PackagePath != PackageThisPackage {
				continue
			}
			tmp := p.filterElementByName(files, param.TypeString)
			if tmp == nil {
				continue
			}
			param.Item = tmp
			param.PackagePath = tmp.PackagePath
			param.ItemType = tmp.ElementType

		}

	}

}
func (p *Parser) handleConstThisPackage(file *File) {
	// 对该文件的常量进行处理
	log.Printf("处理常量，合并枚举: %s \n", file.Name)
	for _, element := range file.Consts {
		if element.ElementType == ElementConst {
			continue
		}
		// 不是标记为this的不需处理
		//if element.PackagePath != types.PackageThisPackage {
		//	continue
		//}

		files := p.filterFilesByPackage(element.PackagePath)
		tmp := p.filterElementByName(files, element.TypeString)
		if tmp == nil {
			continue
		}
		// 为枚举类型添加枚举成员
		if tmp.Elements == nil {
			tmp.Elements = make(map[ElementType][]*Element)
		}
		tmp.ElementType = ElementEnum
		tmp.Elements[ElementEnum] = append(tmp.Elements[ElementEnum], element.Clone())

	}

}
func (p *Parser) handleVarThisPackage(file *File) {
	log.Printf("处理本包变量: %s \n", file.Name)
	for _, element := range file.Vars {
		if element.PackagePath != PackageThisPackage {
			continue
		}
		files := p.filterFilesByPackage(element.PackagePath)
		tmp := p.filterElementByName(files, element.TypeString)
		if tmp == nil {
			continue
		}
		element.Item = tmp
		element.ItemType = tmp.ElementType

	}
}
func (p *Parser) handleActual2(file *File) {
	log.Println("处理已更新的结构，更新其引用")

	for _, eleStruct := range file.Structs {
		if !eleStruct.FromParent {
			continue
		}
		log.Printf("  处理结构体: %s \n", eleStruct.Name)
		for _, eleMethod := range eleStruct.Elements[ElementMethod] {
			for _, param := range eleMethod.Elements[ElementParam] {
				tmp := p.findFileByPackageAndType(param.PackagePath, param.TypeString)
				if tmp != nil {
					param.Item = tmp
				}

			}
			for _, param := range eleMethod.Elements[ElementResult] {
				tmp := p.findFileByPackageAndType(param.PackagePath, param.TypeString)
				if tmp != nil {
					param.Item = tmp
				}

			}

		}

	}
}
func (p *Parser) handleActual(file *File) {
	log.Println("处理泛型的实际映射类型，合并继承")

	for _, eleStruct := range file.Structs {
		//if eleStruct.Name != "UserController" {
		//	continue
		//}
		if !eleStruct.FromParent {
			continue
		}
		log.Printf("  处理结构体: %s \n", eleStruct.Name)
		needDel := -1
		for idx, eleField := range eleStruct.Elements[ElementField] {
			if !eleField.FromParent || eleField.Name != "" {
				continue
			}
			needDel = idx
			// 处理需要继承的字段时，将该字段从当前结构中删除
			// 然后已经继承过来的字段，由于Name不是空，下次执行handleActual时不会进入此循环
			log.Printf("    处理字段 %s \n", eleField.TypeString)

			//继承父级时，将当前结构中声明的实际类型拉取出来
			typeParams := eleField.Elements[ElementGeneric]
			log.Printf("      字段声明了 %d 个泛型参数 \n", len(typeParams))
			// 赋值它的原始类型名 比如实际类型是 int 原始是声明为 T 的
			// 方便后面方法处理的时候进行匹配
			for _, param := range typeParams {
				for _, originTypeParam := range eleField.Item.Elements[ElementGeneric] {
					if originTypeParam.Index != param.Index {
						continue
					}
					param.Name = originTypeParam.Name
				}
			}
			if eleField.Item == nil {
				continue
			}
			// TODO: 继承父级时需要将父级的导出字段也拉出来, 需要先去父级的那个Struct里先处理好

			eleFieldType := eleField.Item
			log.Printf("    处理结构 %s 的字段继承\n", eleStruct.Name)
			for _, eleFieldTypeField := range eleFieldType.Elements[ElementField] {
				if eleFieldTypeField.Private() {
					continue
				}
				if eleFieldTypeField == nil {
					continue
				}
				//if eleFieldTypeField.Item == nil {
				//	continue
				//}
				newField := eleFieldTypeField.Clone()

				for _, e3 := range typeParams {
					// 根据泛型的索引位置来确定实际类型
					if newField.Item != nil && newField.Item.Name != e3.Name {
						continue
					}
					newField.Item = e3.Clone()
					newField.ItemType = e3.ElementType
				}
				newField.FromParent = true
				eleStruct.Elements[ElementField] = append(eleStruct.Elements[ElementField], newField)

			}

			log.Printf("    处理结构 %s 的方法继承\n", eleStruct.Name)
			for _, e2 := range eleFieldType.Elements[ElementMethod] {
				if e2.Private() {
					continue
				}
				newMethodFromParent := e2.Clone()
				for _, e3 := range newMethodFromParent.Elements[ElementParam] {
					if !e3.Generic() {
						continue
					}
					for _, e4 := range typeParams {
						if e3.Item.Name != e4.Name {
							continue
						}
						e3.Item = e4.Clone()
						e3.ItemType = e4.ElementType

					}
				}

				for _, e3 := range newMethodFromParent.Elements[ElementResult] {
					if !e3.Generic() {
						continue
					}
					for _, e4 := range typeParams {
						if e3.Item.Name != e4.Name {
							continue
						}
						e3.Item = e4.Clone()
						e3.ItemType = e4.ElementType

					}

				}
				newMethodFromParent.FromParent = true
				receiver := newMethodFromParent.MustGetElement(ElementReceiver)
				receiver.TypeString = eleStruct.Name

				eleStruct.Elements[ElementMethod] = append(eleStruct.Elements[ElementMethod], newMethodFromParent)

			}

		}
		if needDel != -1 {
			fmt.Println("pre count:", len(eleStruct.Elements[ElementField]))
			eleStruct.Elements[ElementField] = slices.Delete(eleStruct.Elements[ElementField], needDel, needDel+1)
			fmt.Println("after count:", len(eleStruct.Elements[ElementField]))
		}

	}

}

func (p *Parser) VisitStruct(check func(element *Element) bool, f func(element *Element)) {
	for _, file := range p.Files {
		for _, e := range file.Structs {
			if e.ElementType != ElementStruct {
				continue
			}
			if !check(e) {
				continue
			}
			f(e)
		}
	}
}

func (p *Parser) handleActual3(file *File) {
	log.Println("处理方法参数和返回值中的泛型参数, 合并继承")

	for _, eleStruct := range file.Structs {
		for _, eleMethod := range eleStruct.Elements[ElementMethod] {
			log.Println("处理方法:", eleMethod.Name)
			// 遍历参数
			for _, eleParam := range eleMethod.Elements[ElementParam] {
				if !eleParam.Generic() {
					continue
				}
				log.Println("处理参数:", eleParam.Name, " ", eleParam.TypeString)
				tParams := make([]*Element, 0)
				for _, element := range eleParam.Elements[ElementGeneric] {
					tParams = append(tParams, element.Clone())
				}
				for _, element := range eleParam.Item.Elements[ElementField] {
					if element.ItemType == ElementGeneric {
						for _, param := range tParams {
							if element.Item.Index == param.Index {
								element.Name = param.Name
								element.PackagePath = param.PackagePath
								element.PackageName = param.PackageName
								element.ItemType = param.ItemType
								element.Item = param.Item.Clone()
								element.TypeString = param.TypeString
								element.Docs = param.Docs
								element.Comment = param.Comment
								element.FromParent = true
								if element.Elements == nil {
									element.Elements = make(map[ElementType][]*Element)
								}
								if param.Elements == nil {
									param.Elements = make(map[ElementType][]*Element)
								}
								element.Elements[ElementField] = copySlice(param.Elements[ElementField])

							}
						}
					}
				}
			}

			for _, eleResult := range eleMethod.Elements[ElementResult] {
				if !eleResult.Generic() {
					continue
				}
				log.Println("处理返回值:", eleResult.Name, " ", eleResult.TypeString)
				// 找到泛型类型中的泛型字段, 使用实际类型进行替换
				// 先找到真实类型
				tParams := make([]*Element, 0)
				for _, element := range eleResult.Elements[ElementGeneric] {
					tParams = append(tParams, element.Clone())
				}

				for _, element := range eleResult.Item.Elements[ElementField] {
					if element.ItemType == ElementGeneric {
						for _, param := range tParams {
							if element.Item.Index == param.Index {
								element.Name = param.Name
								element.PackagePath = param.PackagePath
								element.PackageName = param.PackageName
								element.ItemType = param.ItemType
								element.Item = param.Item.Clone()
								element.TypeString = param.TypeString
								element.Docs = param.Docs
								element.Comment = param.Comment
								element.FromParent = true
								eleResult.ElementString = eleResult.TypeString + element.TypeString
								if element.Elements == nil {
									element.Elements = make(map[ElementType][]*Element)
								}
								if param.Elements == nil {
									param.Elements = make(map[ElementType][]*Element)
								}
								element.Elements[ElementField] = copySlice(param.Elements[ElementField])

							}
						}
					}
				}

			}
		}

	}

}

func (p *Parser) handleActual4(file *File) {
	log.Println("处理方法参数和返回值中的泛型参数, 合并继承")
}
