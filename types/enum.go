package types

import "github.com/linxlib/astp/constants"

type Enum struct {
	Index    int                `json:"index"`
	Type     string             `json:"type"`
	TypeName string             `json:"type_name"`
	ElemType constants.ElemType `json:"elem_type"`
	Name     string             `json:"name"`
	Doc      []*Comment         `json:"doc,omitempty"`
	Comment  []*Comment         `json:"comment,omitempty"`
	Private  bool               `json:"private,omitempty"`
	Iota     bool               `json:"iota"`
	Enums    []*EnumItem        `json:"enums,omitempty"`
}

func (e *Enum) String() string {
	return e.Name
}

func (e *Enum) Clone() *Enum {
	if e == nil {
		return nil
	}
	return &Enum{
		Index:    e.Index,
		Type:     e.Type,
		TypeName: e.TypeName,
		ElemType: e.ElemType,
		Name:     e.Name,
		Doc:      e.Doc,
		Comment:  e.Comment,
		Private:  e.Private,
		Iota:     e.Iota,
		Enums:    CopySlice(e.Enums),
	}
}

type EnumItem struct {
	Index   int        `json:"index"`
	Name    string     `json:"name"`
	Type    string     `json:"type"`
	Value   any        `json:"value"`
	Private bool       `json:"private"`
	Doc     []*Comment `json:"doc,omitempty"`
	Comment []*Comment `json:"comment,omitempty"`
}

func (e *EnumItem) String() string {
	return e.Name
}

func (e *EnumItem) Clone() *EnumItem {
	if e == nil {
		return nil
	}
	if !deepClone {
		return e
	}
	return &EnumItem{
		Index:   e.Index,
		Name:    e.Name,
		Type:    e.Type,
		Value:   e.Value,
		Private: e.Private,
		Doc:     CopySlice(e.Doc),
		Comment: CopySlice(e.Comment),
	}
}
