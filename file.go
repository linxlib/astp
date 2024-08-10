package astp

// File 代表一个已解析的go文件
type File struct {
	Name        string     `json:",omitempty" yaml:",omitempty"` //文件名
	FilePath    string     `json:",omitempty" yaml:",omitempty"` //文件路径
	PackageName string     `json:",omitempty" yaml:",omitempty"` //包名
	PackagePath string     `json:",omitempty" yaml:",omitempty"` //包路径
	Docs        []string   `json:",omitempty" yaml:",omitempty"` //文档
	Comments    []string   `json:",omitempty" yaml:",omitempty"` //注释
	Imports     []*Import  `json:",omitempty" yaml:",omitempty"` //导入
	Consts      []*Element `json:",omitempty" yaml:",omitempty"` //常量
	Vars        []*Element `json:",omitempty" yaml:",omitempty"` //变量
	Structs     []*Element `json:",omitempty" yaml:",omitempty"` //结构体
	Methods     []*Element `json:"-"`                            //结构体方法
	Funcs       []*Element `json:",omitempty" yaml:",omitempty"` //函数

}

func (f *File) IsMain() bool {
	return f.PackageName == "main"
}
