package types

import (
	"github.com/linxlib/astp/constants"
	"strings"
)

var _ IElem[*Struct] = (*Struct)(nil)

type Struct struct {
	Name      string             `json:"name"`
	Index     int                `json:"index"`
	Key       string             `json:"key"`
	KeyHash   string             `json:"key_hash"`
	TypeName  string             `json:"type_name"`
	Type      string             `json:"type"`
	Private   bool               `json:"private"`
	Generic   bool               `json:"generic"`
	ElemType  constants.ElemType `json:"elem_type"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Field     []*Field           `json:"field,omitempty"`
	Doc       []*Comment         `json:"doc,omitempty"`
	Comment   []*Comment         `json:"comment,omitempty"`
	Method    []*Function        `json:"method,omitempty"`
	Package   *Package           `json:"package,omitempty"`
	Enum      *Enum              `json:"enum,omitempty"`
}

func (s *Struct) IsEnum() bool {
	return s.Enum != nil
}
func (s *Struct) IsTop() bool {

	b := false
	for _, field := range s.Field {
		if !field.IsTop() {
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
