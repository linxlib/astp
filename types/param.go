package types

import (
	"github.com/linxlib/astp/constants"
)

var _ IElem[*Param] = (*Param)(nil)

type Param struct {
	Index     int                `json:"index"`
	Name      string             `json:"name"`
	TypeName  string             `json:"type_name"`
	ElemType  constants.ElemType `json:"elem_type"`
	Package   *Package           `json:"package,omitempty"`
	Type      string             `json:"type"`
	Slice     bool               `json:"slice"`
	Pointer   bool               `json:"pointer"`
	Generic   bool               `json:"generic"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Struct    *Struct            `json:"struct,omitempty"`
}

func (p *Param) String() string {
	return p.Name + " " + p.TypeName
}

func (p *Param) Clone() *Param {
	if p == nil {
		return nil
	}
	return &Param{
		Index:     p.Index,
		Name:      p.Name,
		TypeName:  p.TypeName,
		ElemType:  p.ElemType,
		Package:   p.Package.Clone(),
		Type:      p.Type,
		Struct:    p.Struct.Clone(),
		Slice:     p.Slice,
		Pointer:   p.Pointer,
		Generic:   p.Generic,
		TypeParam: CopySlice(p.TypeParam),
	}
}
