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

type EnumItem struct {
	Index   int        `json:"index"`
	Name    string     `json:"name"`
	Type    string     `json:"type"`
	Value   any        `json:"value"`
	Private bool       `json:"private"`
	Doc     []*Comment `json:"doc"`
	Comment []*Comment `json:"comment"`
}
