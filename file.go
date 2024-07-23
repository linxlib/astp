package astp

// File 代表一个已解析的go文件
type File struct {
	Name        string    //文件名
	PackageName string    //包名
	Imports     []*Import //导入
	PackagePath string    //包路径
	FilePath    string    //文件路径
	Structs     []*Struct //结构体
	Docs        []string  //文档
	Comments    []string  //注释
	Methods     []*Method //结构体方法
	Funcs       []*Method //函数
	Consts      []*Const  //常量
	Vars        []*Var    //变量
}
