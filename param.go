package astp

import "reflect"

// ParamField 结构体字段
type ParamField struct {
	Index       int     //索引
	Name        string  //字段名
	PackagePath string  //包名
	Type        *Struct //类型
	rType       reflect.Type
	HasTag      bool              //是否有tag
	Tag         reflect.StructTag `json:"-"` //tag
	TypeString  string            //类型文本
	InnerType   bool              //内置类型
	Private     bool              //私有
	Pointer     bool              //指针字段
	Slice       bool              //数组字段
	IsStruct    bool              //是否结构体
	Docs        []string          //文档
	Comment     string            //注释
	IsGeneric   bool
}

func (p *ParamField) Clone() *ParamField {
	return &ParamField{
		Index:       p.Index,
		Name:        p.Name,
		PackagePath: p.PackagePath,
		Type:        p.Type.Clone(),
		InnerType:   p.InnerType,
		Docs:        p.Docs,
		Comment:     p.Comment,
		HasTag:      p.HasTag,
		Tag:         p.Tag,
		TypeString:  p.TypeString,
		IsStruct:    p.IsStruct,
		Private:     p.Private,
		Slice:       p.Slice,
		IsGeneric:   p.IsGeneric,
	}
}
func (p *ParamField) GetName() string {
	return p.Name
}

func (p *ParamField) SetType(namer *Struct) {
	p.Type = namer.Clone()
}

func (p *ParamField) GetRType() reflect.Type {
	return p.rType
}
func (p *ParamField) SetRType(t reflect.Type) {
	p.rType = t
}

func (p *ParamField) SetInnerType(b bool) {
	p.InnerType = b
}

func (p *ParamField) SetIsStruct(b bool) {
	p.IsStruct = b
}

func (p *ParamField) SetTypeString(s string) {
	p.TypeString = s
}

func (p *ParamField) SetPointer(b bool) {
	p.Pointer = b
}

func (p *ParamField) SetPrivate(b bool) {
	p.Private = b
}

func (p *ParamField) SetSlice(b bool) {
	p.Slice = b
}

func (p *ParamField) SetPackagePath(s string) {
	p.PackagePath = s
}

func (p *ParamField) GetType() *Struct {
	return p.Type
}
