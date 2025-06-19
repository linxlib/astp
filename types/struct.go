package types

import (
	"github.com/linxlib/astp/constants"
	"reflect"
	"strings"
)

var _ IElem[*Struct] = (*Struct)(nil)

type Struct struct {
	Name      string             `json:"name"`
	Index     int                `json:"index"`
	Key       string             `json:"-"`
	KeyHash   string             `json:"-"`
	TypeName  string             `json:"type_name"`
	Type      string             `json:"type"`
	Private   bool               `json:"private,omitempty"`
	Generic   bool               `json:"generic,omitempty"`
	ElemType  constants.ElemType `json:"elem_type"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Field     []*Field           `json:"field,omitempty"`
	Doc       []*Comment         `json:"doc,omitempty"`
	Comment   []*Comment         `json:"comment,omitempty"`
	Method    []*Function        `json:"method,omitempty"`
	Package   *Package           `json:"package,omitempty"`
	Enum      *Enum              `json:"enum,omitempty"`

	rValue reflect.Value
	value  any
	t      reflect.Type
}

func (s *Struct) IsEnum() bool {
	if s==nil {
		return false
	}
	return s.Enum != nil
}
func (s *Struct) IsTop() bool {

	b := false
	for _, field := range s.Field {

		if !field.IsTop() || field.Name == constants.EmptyName {
			return false
		}
		b = true
	}
	if b {
		if len(s.TypeParam) > 0 {
			return false
		}
	}

	if s.Field == nil || len(s.Field) == 0 {
		//fmt.Printf("struct: %s is top struct\n", s.Name)
		return true
	}
	if b {
		//fmt.Printf("struct: %s is top struct\n", s.Name)
	}
	return b
}

func (s *Struct) String() string {
	return s.Key
}
func (s *Struct) Clone() *Struct {
	if s == nil {
		return nil
	}
	return &Struct{
		Index:     s.Index,
		Name:      s.Name,
		Key:       s.Key,
		KeyHash:   s.KeyHash,
		TypeName:  s.TypeName,
		Type:      s.Type,
		Private:   s.Private,
		Generic:   s.Generic,
		TypeParam: CopySlice(s.TypeParam),
		Field:     CopySlice(s.Field),
		Doc:       CopySlice(s.Doc),
		ElemType:  s.ElemType,
		Enum:      s.Enum.Clone(),
		Comment:   CopySlice(s.Comment),
		//Method:    CopySlice(s.Method),
		Package: s.Package.Clone(),
	}
}

func (s *Struct) CloneFull() *Struct {
	if s == nil {
		return nil
	}
	return &Struct{
		Index:     s.Index,
		Name:      s.Name,
		Key:       s.Key,
		KeyHash:   s.KeyHash,
		TypeName:  s.TypeName,
		Type:      s.Type,
		Private:   s.Private,
		Generic:   s.Generic,
		TypeParam: CopySlice(s.TypeParam),
		Field:     CopySlice(s.Field),
		Doc:       CopySlice(s.Doc),
		Enum:      s.Enum.Clone(),
		Comment:   CopySlice(s.Comment),
		Method:    CopySlice(s.Method),
		Package:   s.Package.Clone(),
	}
}

func (s *Struct) HasAttr(attr constants.AttrType) bool {
	for _, comment := range s.Doc {
		if comment.Op && comment.AttrType == attr {
			return true
		}
	}
	return false
}
func (s *Struct) GetAttr() constants.AttrType {

	for _, comment := range s.Doc {
		if comment.Op {
			return comment.AttrType
		}
	}
	return 0
}
func (s *Struct) HasParamAttr() bool {
	if strings.Contains(s.Name, "Body") ||
		strings.Contains(s.Name, "Query") ||
		strings.Contains(s.Name, "Header") ||
		strings.Contains(s.Name, "Form") ||
		strings.Contains(s.Name, "Json") ||
		strings.Contains(s.Name, "Cookie") ||
		strings.Contains(s.Name, "Multipart") ||
		strings.Contains(s.Name, "Path") {
		return true
	}
	if s.HasAttr(constants.AT_BODY) ||
		s.HasAttr(constants.AT_QUERY) ||
		s.HasAttr(constants.AT_HEADER) ||
		s.HasAttr(constants.AT_FORM) ||
		s.HasAttr(constants.AT_PLAIN) ||
		s.HasAttr(constants.AT_JSON) ||
		s.HasAttr(constants.AT_COOKIE) ||
		s.HasAttr(constants.AT_XML) ||
		s.HasAttr(constants.AT_MULTIPART) ||
		s.HasAttr(constants.AT_PATH) {
		return true
	}
	return false
}

func (s *Struct) HasCustomAttr(attr string) bool {
	upper := strings.ToUpper(attr)
	if v, ok := constants.AttrTypes[upper]; ok {
		return s.HasAttr(v)
	}
	for _, comment := range s.Doc {
		if comment.Op {
			if comment.CustomAttr == upper {
				return true
			}
		}
	}
	return false
}
func (s *Struct) GetAttrs() []*Comment {
	return CopySliceWithFilter(s.Doc, func(comment *Comment) bool {
		return comment.Op
	})
}

func (s *Struct) SetRValue(r reflect.Value) {
	s.rValue = r
}

func (s *Struct) SetValue(v any) {
	s.value = v
}
func (s *Struct) SetRType(t reflect.Type) {
	s.t = t
}
func (s *Struct) GetRValue() reflect.Value {
	return s.rValue
}
func (s *Struct) GetValue() any {
	return s.value
}
func (s *Struct) GetRType() reflect.Type {
	return s.t
}

func (s *Struct) VisitMethods(filter func(f *Function) bool, handler func(f *Function)) {
	for _, method := range s.Method {
		if filter(method) {
			handler(method)
		}
	}
}
func (s *Struct) VisitFields(filter func(f *Field) bool, handler func(f *Field)) {
	if s == nil {
		return
	}
	if s.Field == nil || len(s.Field) == 0 {
		return
	}
	for _, field := range s.Field {
		if filter(field) {
			handler(field)
		}
	}
}

func (s *Struct) GetAttrValue(attr constants.AttrType) string {
	for _, comment := range s.Doc {
		if comment.Op && comment.AttrType == attr {
			return comment.AttrValue
		}
	}
	return ""
}

func (s *Struct) GetCustomAttrs() []*Comment {
	var result = make([]*Comment, 0)
	for _, comment := range s.GetAttrs() {
		if comment.AttrType == constants.AT_CUSTOM {
			result = append(result, comment)
		}
	}
	return result
}
