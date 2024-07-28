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
	Inter       *Interface
	Docs        []string //文档
	Comment     string   //注释
	IsGeneric   bool
	TypeParams  []*TypeParam
}

func (s *Struct) SetActualType(name string, as *Struct) {
	if s.IsGeneric && len(s.TypeParams) > 0 {
		for _, param := range s.TypeParams {
			if param.Name != name {
				param.ActualType = as.Clone()
			}
		}
	}
}
func (s *Struct) IsGenericType(name string) bool {
	for _, param := range s.TypeParams {
		if param.Name == name {
			return true
		}
	}
	return false
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
func (s *Struct) SetThisPackageTypeParams(file *File) {
	for _, param := range s.TypeParams {
		if param.PackagePath == "this" {
			for _, s2 := range file.Structs {
				if s2.Name == strings.TrimLeft(param.TypeName, "*") {
					param.SetType(s2)
					param.SetPackagePath(s2.PackagePath)
				}
			}
		}
	}
}
func (s *Struct) SetThisPackageFields(file *File) {
	for _, field := range s.Fields {
		if field.PackagePath == "this" {
			if s.IsGeneric {
				for _, param := range s.TypeParams {
					if field.TypeString == param.Name {
						field.SetType(param.Type)
						field.SetPackagePath(param.PackagePath)
					}
				}

			} else {
				for _, s2 := range file.Structs {
					if s2.Name == strings.TrimLeft(field.TypeString, "*") {
						field.SetType(s2)
						field.SetPackagePath(s2.PackagePath)
					}
				}
			}

		}
	}
}

func (s *Struct) SetThisPackageMethodParams(file *File) {
	for _, method := range s.Methods {
		for _, param := range method.Params {
			if param.PackagePath == "this" {
				if method.IsGeneric && method.TypeParams != nil && len(method.TypeParams) > 0 {
					for _, typeParam := range method.TypeParams {
						if typeParam.Name == param.TypeString {
							method.SetParamType(param, typeParam.Type)
						}
					}
				} else {
					for _, s2 := range file.Structs {
						realTypeName := strings.TrimLeft(param.TypeString, "*")
						index := strings.LastIndex(realTypeName, "[")
						if index > 0 {
							realTypeName = realTypeName[:index]
						}

						if strings.EqualFold(realTypeName, s2.Name) {
							param.SetType(s2)
							param.SetPackagePath(s2.PackagePath)
						}

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
