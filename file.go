package astp

import "github.com/linxlib/astp/types"

// File 代表一个已解析的go文件
type File struct {
	Name        string           `json:",omitempty"` //文件名
	FilePath    string           `json:",omitempty"` //文件路径
	PackageName string           `json:",omitempty"` //包名
	PackagePath string           `json:",omitempty"` //包路径
	Docs        []string         `json:"-"`          //文档
	Comments    []string         `json:"-"`          //注释
	Imports     []*Import        `json:"-"`          //导入
	Consts      []*types.Element `json:",omitempty"` //常量
	Vars        []*types.Element `json:",omitempty"` //变量
	Structs     []*types.Element `json:",omitempty"` //结构体
	Methods     []*types.Element `json:"-"`          //结构体方法
	Funcs       []*types.Element `json:",omitempty"` //函数

}

func (f *File) IsMain() bool {
	return f.PackageName == "main"
}
