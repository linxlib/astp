package types

import (
	"github.com/linxlib/astp/constants"
	"reflect"
)

var _ IElem[*Function] = (*Function)(nil)

type Function struct {
	Name      string             `json:"name"`
	Key       string             `json:"key"`
	KeyHash   string             `json:"key_hash"`
	ElemType  constants.ElemType `json:"elem_type,omitempty"`
	TypeName  string             `json:"type_name"`
	Doc       []*Comment         `json:"doc,omitempty"`
	Private   bool               `json:"private,omitempty"`
	Index     int                `json:"index"`
	Package   *Package           `json:"package,omitempty"`
	Generic   bool               `json:"generic,omitempty"`
	TypeParam []*TypeParam       `json:"type_param,omitempty"`
	Param     []*Param           `json:"param,omitempty"`
	Result    []*Param           `json:"result,omitempty"`
	Receiver  *Receiver          `json:"receiver,omitempty"`
	rValue    reflect.Value
	value     any
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
	if !deepClone {
		return f
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

func (f *Function) VisitParams(handler func(param *Param)) {
	for _, param := range f.Param {
		handler(param)
	}
}
func (f *Function) VisitResults(handler func(param *Param)) {
	for _, param := range f.Result {
		handler(param)
	}
}

func (f *Function) SetRValue(t reflect.Value) {
	f.rValue = t
}
func (f *Function) SetValue(t any) {
	f.value = t
}

func (f *Function) GetRValue() reflect.Value {
	return f.rValue
}
func (f *Function) GetValue() any {
	return f.value
}

func (f *Function) GetAttrValue(attr constants.AttrType) string {
	for _, comment := range f.Doc {
		if comment.Op && comment.AttrType == attr {
			return comment.AttrValue
		}
	}
	return ""
}
func (f *Function) GetAttrs() []*Comment {
	return CopySliceWithFilter(f.Doc, func(comment *Comment) bool {
		return comment.Op
	})
}

func (f *Function) GetCustomAttrs() []*Comment {
	var result = make([]*Comment, 0)
	for _, comment := range f.GetAttrs() {
		if comment.AttrType == constants.AT_CUSTOM {
			result = append(result, comment)
		}
	}
	return result
}
func (f *Function) HasAttr(attr constants.AttrType) bool {
	for _, comment := range f.Doc {
		if comment.Op && comment.AttrType == attr {
			return true
		}
	}
	return false
}
func (f *Function) HasAttrs() bool {
	tmp := f.GetAttrs()
	return tmp != nil && len(tmp) > 0
}

func (f *Function) GetHttpMethodAttrs() []*Comment {
	return CopySliceWithFilter(f.Doc, func(comment *Comment) bool {
		return comment.Op && (comment.AttrType == constants.AT_POST ||
			comment.AttrType == constants.AT_PUT ||
			comment.AttrType == constants.AT_DELETE ||
			comment.AttrType == constants.AT_GET || comment.AttrType == constants.AT_PATCH || comment.AttrType == constants.AT_OPTIONS ||
			comment.AttrType == constants.AT_ANY)
	})
}
