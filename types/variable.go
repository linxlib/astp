package types

import (
	"fmt"
	"github.com/linxlib/astp/constants"
)

var _ IElem[*Variable] = (*Variable)(nil)

type Variable struct {
	Name     string             `json:"name"`
	ElemType constants.ElemType `json:"elem_type"`
	Index    int                `json:"index"`
	Value    any                `json:"value"`
	Type     string             `json:"type"`
	TypeName string             `json:"type_name"`
	Iota     bool               `json:"iota,omitempty"`
	Package  *Package           `json:"package,omitempty"`
	Struct   *Struct            `json:"struct,omitempty"`
	Doc      []*Comment         `json:"doc,omitempty"`
	Comment  []*Comment         `json:"comment,omitempty"`
}

func (v *Variable) String() string {
	return fmt.Sprintf("%s(%s)", v.Name, v.Value)
}

func (v *Variable) Clone() *Variable {
	if v == nil {
		return nil
	}
	return &Variable{
		Name:     v.Name,
		ElemType: v.ElemType,
		Type:     v.Type,
		Value:    v.Value,
		TypeName: v.TypeName,
		Package:  v.Package.Clone(),
		Struct:   v.Struct.Clone(),
		Doc:      CopySlice(v.Doc),
		Comment:  CopySlice(v.Comment),
	}
}

type Const = Variable
