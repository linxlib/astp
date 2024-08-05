package astp

type EnumItem struct {
	Name    string
	Value   any
	Docs    []string
	Comment string
}

// Struct 结构体
type Struct struct {
	Name        string         //结构体名
	PackagePath string         //所属包名
	KeyHash     string         //所在文件hash值, 重复解析时直接获取
	Fields      []*StructField //结构体字段
	Methods     []*Method      //结构体方法
	HasParent   bool           //是否有继承结构
	IsInterface bool
	Inter       *Interface
	Docs        []string //文档
	Comment     string   //注释
	IsGeneric   bool
	TypeParams  []*TypeParam
	Enums       []*EnumItem
}

func (s *Struct) GetTypeString() string {
	return s.Name
}

func (s *Struct) GetTypeParams() []*TypeParam {
	return s.TypeParams
}

func (s *Struct) SetTypeParams(tps []*TypeParam) {
	copy(s.TypeParams, tps)
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
	n := &Struct{
		Name:        s.Name,
		PackagePath: s.PackagePath,
		KeyHash:     s.KeyHash,
		IsGeneric:   s.IsGeneric,
		HasParent:   s.HasParent,
		IsInterface: s.IsInterface,
		Docs:        s.Docs,
		Comment:     s.Comment,
	}
	n.Fields = make([]*StructField, len(s.Fields))
	n.Methods = make([]*Method, len(s.Methods))
	n.TypeParams = make([]*TypeParam, len(s.TypeParams))
	copy(n.Fields, s.Fields)
	for i, method := range s.Methods {
		n.Methods[i] = method.Clone()
	}
	//copy(n.Methods, s.Methods)
	copy(n.TypeParams, s.TypeParams)
	return n
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

//func (s *Struct) HandleCurrentPackageRefs(files map[string]*File) {
//	s.handleCurPkgTypeParam(files)
//	s.handleCurPkgFields(files)
//	s.handleCurPkgMethodParams(files)
//}
//
//func (s *Struct) HandleCurrentPackageRef(file *File) {
//	s.handleCurFileTypeParams(file)
//	s.handleCurFileFields(file)
//	s.handleCurFileMethodParams(file)
//}

//func (s *Struct) handleCurPkgTypeParam(files map[string]*File) {
//	for _, param := range s.TypeParams {
//		if param.PackagePath == "this" {
//			for _, file := range files {
//				for _, s2 := range file.Structs {
//					if s2.Name == strings.TrimLeft(param.TypeName, "*") {
//						param.SetType(s2)
//						param.SetPackagePath(s2.PackagePath)
//						return
//					}
//				}
//			}
//
//		}
//	}
//}

// TODO:

/**
type StructList []*Struct

func (l StructList) FindByName(name string) *Struct {
	for _, s := range l {
		if s.Name == name {
			return s
		}
	}
	return nil
}

*/

//func (s *Struct) handleCurFileTypeParams(file *File) {
//	for _, param := range s.TypeParams {
//		if param.PackagePath == "this" {
//			for _, s2 := range file.Structs {
//				if s2.Name == strings.TrimLeft(param.TypeName, "*") {
//					param.SetType(s2)
//					param.SetPackagePath(s2.PackagePath)
//				}
//			}
//		}
//	}
//}
//func (s *Struct) handleCurPkgFields(files map[string]*File) {
//	for _, field := range s.Fields {
//		if field.PackagePath == "this" {
//			if s.IsGeneric {
//				for _, param := range s.TypeParams {
//					if field.TypeString == param.Name {
//						field.SetType(param.Type)
//						field.SetPackagePath(param.PackagePath)
//						return
//					}
//				}
//
//			} else {
//				for _, file := range files {
//					for _, s2 := range file.Structs {
//						realTypeName := strings.TrimLeft(field.TypeString, "*")
//						index := strings.LastIndex(realTypeName, "[")
//						if index > 0 {
//							realTypeName = realTypeName[:index]
//						}
//
//						if strings.EqualFold(realTypeName, s2.Name) {
//							field.SetType(s2)
//							field.SetPackagePath(s2.PackagePath)
//							s.Methods = append(s.Methods, s2.Methods...)
//							s.Docs = append(s.Docs, s2.Docs...)
//							return
//
//						}
//
//					}
//				}
//
//			}
//
//		}
//	}
//}
//
//func (s *Struct) handleCurFileFields(file *File) {
//	for _, field := range s.Fields {
//		if field.PackagePath == "this" {
//			if s.IsGeneric {
//				for _, param := range s.TypeParams {
//					if field.TypeString == param.Name {
//						field.SetType(param.Type)
//						field.SetPackagePath(param.PackagePath)
//					}
//				}
//
//			} else {
//				for _, s2 := range file.Structs {
//					realTypeName := strings.TrimLeft(field.TypeString, "*")
//					index := strings.LastIndex(realTypeName, "[")
//					if index > 0 {
//						realTypeName = realTypeName[:index]
//					}
//
//					if strings.EqualFold(realTypeName, s2.Name) {
//						field.SetType(s2)
//						field.SetPackagePath(s2.PackagePath)
//						s.Methods = append(s.Methods, s2.Methods...)
//						s.Docs = append(s.Docs, s2.Docs...)
//
//					}
//
//				}
//			}
//
//		}
//	}
//}
//func (s *Struct) handleCurPkgMethodParams(files map[string]*File) {
//	for _, method := range s.Methods {
//		for _, param := range method.Params {
//			if param.PackagePath == "this" {
//				if method.IsGeneric && method.TypeParams != nil && len(method.TypeParams) > 0 {
//					for _, typeParam := range method.TypeParams {
//						if typeParam.Name == param.TypeString {
//							method.SetParamType(param, typeParam.Type)
//						}
//					}
//				} else {
//					for _, file := range files {
//						for _, s2 := range file.Structs {
//							realTypeName := strings.TrimLeft(param.TypeString, "*")
//							index := strings.LastIndex(realTypeName, "[")
//							if index > 0 {
//								realTypeName = realTypeName[:index]
//							}
//
//							if strings.EqualFold(realTypeName, s2.Name) {
//								param.SetType(s2)
//								param.SetPackagePath(s2.PackagePath)
//								return
//							}
//
//						}
//					}
//
//				}
//
//			}
//		}
//		for _, param := range method.Results {
//			if param.PackagePath == "this" {
//				for _, file := range files {
//					for _, s2 := range file.Structs {
//						if s2.Name == strings.TrimLeft(param.TypeString, "*") {
//							param.Type = s2
//							param.PackagePath = s2.PackagePath
//							return
//						}
//					}
//				}
//
//			}
//		}
//	}
//}
//
//func (s *Struct) handleCurFileMethodParams(file *File) {
//	for _, method := range s.Methods {
//		for _, param := range method.Params {
//			if param.PackagePath == "this" {
//				if method.IsGeneric && method.TypeParams != nil && len(method.TypeParams) > 0 {
//					for _, typeParam := range method.TypeParams {
//						if typeParam.Name == param.TypeString {
//							method.SetParamType(param, typeParam.Type)
//						}
//					}
//				} else {
//					for _, s2 := range file.Structs {
//						realTypeName := strings.TrimLeft(param.TypeString, "*")
//						index := strings.LastIndex(realTypeName, "[")
//						if index > 0 {
//							realTypeName = realTypeName[:index]
//						}
//
//						if strings.EqualFold(realTypeName, s2.Name) {
//							param.SetType(s2)
//							param.SetPackagePath(s2.PackagePath)
//						}
//
//					}
//				}
//
//			}
//		}
//		for _, param := range method.Results {
//			if param.PackagePath == "this" {
//				for _, s2 := range file.Structs {
//					if s2.Name == strings.TrimLeft(param.TypeString, "*") {
//						param.Type = s2
//						param.PackagePath = s2.PackagePath
//					}
//				}
//			}
//		}
//	}
//}
