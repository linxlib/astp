package types

import (
	"fmt"
	"github.com/linxlib/astp/constants"
)

var _ IElem[*Receiver] = (*Receiver)(nil)

type Receiver struct {
	Name      string             `json:"name"`
	Pointer   bool               `json:"pointer,omitempty"`
	ElemType  constants.ElemType `json:"elem_type,omitempty"`
	Type      string             `json:"type"`
	TypeName  string             `json:"type_name"`
	Generic   bool               `json:"generic,omitempty"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Struct    *Struct            `json:"struct,omitempty"`
}

func (r *Receiver) String() string {
	return fmt.Sprintf("%s(%s)", r.Name, r.Pointer)
}

func (r *Receiver) Clone() *Receiver {
	if r == nil {
		return nil
	}
	if !deepClone {
		return r
	}
	return &Receiver{
		Name:      r.Name,
		Pointer:   r.Pointer,
		ElemType:  r.ElemType,
		Type:      r.Type,
		TypeName:  r.TypeName,
		Generic:   r.Generic,
		TypeParam: CopySlice(r.TypeParam),
		Struct:    r.Struct.Clone(),
	}
}
