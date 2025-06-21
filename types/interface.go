package types

import "github.com/linxlib/astp/constants"

var _ IElem[*Interface] = (*Interface)(nil)

type Interface struct {
	Index     int                `json:"index"`
	Name      string             `json:"name"`
	ElemType  constants.ElemType `json:"elem_type"`
	TypeName  string             `json:"type_name"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Function  []*Function        `json:"function,omitempty"`
	Doc       []*Comment         `json:"doc,omitempty"`
	Param     []*Param           `json:"param,omitempty"`
	Result    []*Param           `json:"result,omitempty"`
}

func (i *Interface) String() string {
	return i.Name
}

func (i *Interface) Clone() *Interface {
	if i == nil {
		return nil
	}
	if !deepClone {
		return i
	}
	return &Interface{
		Index:     i.Index,
		Name:      i.Name,
		ElemType:  i.ElemType,
		TypeName:  i.TypeName,
		TypeParam: CopySlice(i.TypeParam),
		Function:  CopySlice(i.Function),
		Doc:       CopySlice(i.Doc),
		Param:     CopySlice(i.Param),
		Result:    CopySlice(i.Result),
	}
}
