package types

import (
	"github.com/linxlib/astp/constants"
	"reflect"
)

var _ IElem[*Param] = (*Param)(nil)

type Param struct {
	Index     int                `json:"index"`
	Name      string             `json:"name"`
	TypeName  string             `json:"type_name"`
	ElemType  constants.ElemType `json:"elem_type,omitempty"`
	Package   *Package           `json:"package,omitempty"`
	Type      string             `json:"type"`
	Slice     bool               `json:"slice,omitempty"`
	Pointer   bool               `json:"pointer,omitempty"`
	Generic   bool               `json:"generic,omitempty"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Struct    *Struct            `json:"struct,omitempty"`
	rType     reflect.Type
}

func (p *Param) String() string {
	return p.Name + " " + p.TypeName
}

func (p *Param) Clone() *Param {
	if p == nil {
		return nil
	}
	if !deepClone {
		return p
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
func (p *Param) SetRType(t reflect.Type) {
	p.rType = t
}
func (p *Param) GetRType() reflect.Type {
	return p.rType
}
