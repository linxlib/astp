package types

import (
	"github.com/linxlib/astp/constants"
)

var _ IElem[*Function] = (*Function)(nil)

type Function struct {
	Name      string             `json:"name"`
	Key       string             `json:"key"`
	KeyHash   string             `json:"key_hash"`
	ElemType  constants.ElemType `json:"elem_type"`
	TypeName  string             `json:"type_name"`
	Doc       []*Comment         `json:"doc,omitempty"`
	Private   bool               `json:"private"`
	Index     int                `json:"index"`
	Package   *Package           `json:"package,omitempty"`
	Generic   bool               `json:"generic"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Param     []*Param           `json:"param,omitempty"`
	Result    []*Param           `json:"result,omitempty"`
	Receiver  *Receiver          `json:"receiver,omitempty"`
}

func (f *Function) IsOp() bool {
	if f.Doc == nil {
		return false
	}
	for _, comment := range f.Doc {
		if comment.Op {
			return true
		}
	}
	return false
}

func (f *Function) String() string {
	return f.Name
}

func (f *Function) Clone() *Function {
	if f == nil {
		return nil
	}
	return &Function{
		Name:      f.Name,
		Key:       f.Key,
		KeyHash:   f.KeyHash,
		ElemType:  f.ElemType,
		TypeName:  f.TypeName,
		Doc:       CopySlice(f.Doc),
		Generic:   f.Generic,
		Private:   f.Private,
		Index:     f.Index,
		Package:   f.Package.Clone(),
		TypeParam: CopySlice(f.TypeParam),
		Param:     CopySlice(f.Param),
		Result:    CopySlice(f.Result),
		Receiver:  f.Receiver.Clone(),
	}
}
