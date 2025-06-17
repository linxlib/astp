package types

import "github.com/linxlib/astp/constants"

var _ IElem[*TypeParam] = (*TypeParam)(nil)

type TypeParam struct {
	Type          string             `json:"type"`
	TypeName      string             `json:"type_name"`
	Index         int                `json:"index"`
	ElemType      constants.ElemType `json:"elem_type,omitempty"`
	Pointer       bool               `json:"pointer,omitempty"`
	Slice         bool               `json:"slice,omitempty"`
	TypeInterface string             `json:"type_interface,omitempty"`
	Struct        *Struct            `json:"struct,omitempty"`
	Package       *Package           `json:"package,omitempty"`
}

func (t *TypeParam) String() string {
	return t.Type
}

func (t *TypeParam) Clone() *TypeParam {
	if t == nil {
		return nil
	}
	return &TypeParam{
		Index:         t.Index,
		Type:          t.Type,
		TypeName:      t.TypeName,
		ElemType:      t.ElemType,
		Pointer:       t.Pointer,
		Slice:         t.Slice,
		TypeInterface: t.TypeInterface,
		Struct:        t.Struct,
		Package:       t.Package,
	}
}
