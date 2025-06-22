package types

import (
	"reflect"
	"strings"
)

var _ IElem[*Field] = (*Field)(nil)

type Field struct {
	Index     int          `json:"index"`
	Name      string       `json:"name"`
	TypeName  string       `json:"type_name"`
	Type      string       `json:"type"`
	Parent    bool         `json:"parent,omitempty"`
	Private   bool         `json:"private,omitempty"`
	Generic   bool         `json:"generic,omitempty"`
	Slice     bool         `json:"slice,omitempty"`
	Pointer   bool         `json:"pointer,omitempty"`
	TypeParam []*TypeParam `json:"type_param,omitempty"`
	Tag       string       `json:"tag,omitempty"`
	Doc       []*Comment   `json:"doc,omitempty"`
	Comment   []*Comment   `json:"comment,omitempty"`
	Struct    *Struct      `json:"struct,omitempty"`
	Package   *Package     `json:"package,omitempty"`
}

func (f *Field) IsTop() bool {
	return f.Struct == nil
}
func (f *Field) String() string {
	return f.Name + " " + f.TypeName
}

func (f *Field) Clone() *Field {
	if f == nil {
		return nil
	}
	if !deepClone {
		return f
	}
	return &Field{
		Index:     f.Index,
		Name:      f.Name,
		TypeName:  f.TypeName,
		Type:      f.Type,
		Parent:    f.Parent,
		Private:   f.Private,
		Generic:   f.Generic,
		Pointer:   f.Pointer,
		Slice:     f.Slice,
		TypeParam: CopySlice(f.TypeParam),
		Tag:       f.Tag,
		Package:   f.Package.Clone(),
		Doc:       f.Doc,
		Comment:   f.Comment,
		Struct:    f.Struct.Clone(),
	}
}
func (f *Field) HasTag() bool {
	return f.Tag != ""
}
func (f *Field) GetTag() reflect.StructTag {
	return reflect.StructTag(strings.Trim(f.Tag, "`"))
}

func (f *Field) GetTagByName(name string) string {
	if v, ok := f.GetTag().Lookup(name); ok {
		return v
	} else {
		return ""
	}
}
