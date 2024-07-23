package astp

// Import 导入
type Import struct {
	Name       string //名称
	Alias      string //别名
	ImportPath string //导入路径
	IsIgnore   bool   //是否 _ 隐式导入
}
