package astp

import (
	"reflect"
)

// StructField 结构体字段
type StructField struct {
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
}

func (f *StructField) GetType() *Struct {
	return f.Type
}

func (f *StructField) GetName() string {
	return f.Name
}

func (f *StructField) GetRType() reflect.Type {
	return f.rType
}
func (f *StructField) SetRType(t reflect.Type) {
	f.rType = t
}

func (f *StructField) SetPackagePath(s string) {
	f.PackagePath = s
}

func (f *StructField) SetInnerType(b bool) {
	f.InnerType = b
}

func (f *StructField) SetIsStruct(b bool) {
	f.IsStruct = b
}

func (f *StructField) SetTypeString(s string) {
	f.TypeString = s
}

func (f *StructField) SetPointer(b bool) {
	f.Pointer = b
}

func (f *StructField) SetPrivate(b bool) {
	f.Private = b
}

func (f *StructField) SetSlice(b bool) {
	f.Slice = b
}

func (f *StructField) SetType(namer *Struct) {
	f.Type = namer
}

func (f *StructField) GetTag(tag string) string {
	return f.Tag.Get(tag)
}
