package astp

// Var 变量
type Var struct {
	Name        string //变量名
	PackagePath string
	TypeString  string   //类型文本
	Type        *Struct  //类型
	InnerType   bool     //内置类型
	IsStruct    bool     //是否结构体
	Pointer     bool     //指针
	Private     bool     //私有
	Slice       bool     //数组
	Value       any      //值
	Docs        []string //文档
	Comments    string   //注释
}

func (v *Var) GetType() *Struct {
	return v.Type
}

func (v *Var) SetPackagePath(s string) {
	v.PackagePath = s
}

func (v *Var) SetInnerType(b bool) {
	v.InnerType = b
}

func (v *Var) SetIsStruct(b bool) {
	v.IsStruct = b
}

func (v *Var) SetTypeString(s string) {
	v.TypeString = s
}

func (v *Var) SetPointer(b bool) {
	v.Pointer = b
}

func (v *Var) SetPrivate(b bool) {
	v.Private = b
}

func (v *Var) SetSlice(b bool) {
	v.Slice = b
}

func (v *Var) SetType(namer *Struct) {
	v.Type = namer
}

func (v *Var) GetName() string {
	return v.Name
}
