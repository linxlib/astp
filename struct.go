package astp

import (
	"strings"
)

// Struct 结构体
type Struct struct {
	Name        string         //结构体名
	PackagePath string         //所属包名
	KeyHash     string         //所在文件hash值, 重复解析时直接获取
	Fields      []*StructField //结构体字段
	Methods     []*Method      //结构体方法
	IsInterface bool
	Docs        []string //文档
	Comment     string   //注释
}

func (s *Struct) Clone() *Struct {
	if s == nil {
		return nil
	}
	return &Struct{
		Name:        s.Name,
		PackagePath: s.PackagePath,
		KeyHash:     s.KeyHash,
		//Fields:      s.Fields,
		//Methods:     s.Methods,
		IsInterface: s.IsInterface,
		Docs:        s.Docs,
		Comment:     s.Comment,
	}
}
func (s *Struct) GetType() *Struct {
	return s
}

func (s *Struct) SetType(namer *Struct) {

}

func (s *Struct) SetInnerType(b bool) {

}

func (s *Struct) SetIsStruct(b bool) {

}

func (s *Struct) SetTypeString(str string) {

}

func (s *Struct) SetPointer(b bool) {

}

func (s *Struct) SetPrivate(b bool) {

}

func (s *Struct) SetSlice(b bool) {

}

func (s *Struct) SetPackagePath(str string) {
	s.PackagePath = str
}

func (s *Struct) AppendFields(fields []*StructField) {
	s.Fields = append(s.Fields, fields...)
}

func (s *Struct) GetName() string {
	return s.Name
}

func (s *Struct) GetAllFieldsByTag(tag string) []*StructField {
	rtn := make([]*StructField, 0)
	for _, field := range s.Fields {
		if _, ok := field.Tag.Lookup(tag); ok {
			rtn = append(rtn, field)
		}
	}
	return rtn
}
func (s *Struct) SetThisPackageFields(file *File) {
	for _, field := range s.Fields {
		if field.PackagePath == "this" {
			for _, s2 := range file.Structs {
				if s2.Name == strings.TrimLeft(field.TypeString, "*") {
					field.Type = s2
					field.PackagePath = s2.PackagePath
				}
			}
		}
	}
}
func (s *Struct) SetThisPackageMethodParams(file *File) {
	for _, method := range s.Methods {
		for _, param := range method.Params {
			if param.PackagePath == "this" {
				for _, s2 := range file.Structs {
					if s2.Name == strings.TrimLeft(param.TypeString, "*") {
						param.Type = s2
						param.PackagePath = s2.PackagePath
					}
				}
			}
		}
		for _, param := range method.Results {
			if param.PackagePath == "this" {
				for _, s2 := range file.Structs {
					if s2.Name == strings.TrimLeft(param.TypeString, "*") {
						param.Type = s2
						param.PackagePath = s2.PackagePath
					}
				}
			}
		}
	}
}
