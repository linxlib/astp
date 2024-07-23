package astp

// Const 常量
type Const struct {
	Name        string // 名称
	PackagePath string
	TypeString  string   //类型文本
	Type        *Struct  //类型
	Private     bool     //私有
	Slice       bool     //数组
	Value       any      //值
	Docs        []string //文档
	Comments    string   //注释
	IsIota      bool     //iota自动判定值（枚举）
}

func (c *Const) SetType(namer *Struct) {
	c.Type = namer
}

func (c *Const) GetType() *Struct {
	return c.Type
}

func (c *Const) SetPackagePath(s string) {
	c.PackagePath = s
}

func (c *Const) SetInnerType(b bool) {
}

func (c *Const) SetIsStruct(b bool) {

}

func (c *Const) SetTypeString(s string) {
	c.TypeString = s
}

func (c *Const) SetPointer(b bool) {
}

func (c *Const) SetPrivate(b bool) {
	c.Private = b
}

func (c *Const) SetSlice(b bool) {
	c.Slice = b
}

func (c *Const) GetName() string {
	return c.Name
}

type ConstSection = []*Const
