package constants

type ElemType = string

const (
	ElemVar       ElemType = "var"
	ElemFunc      ElemType = "func"
	ElemStruct    ElemType = "struct"
	ElemInterface ElemType = "interface"
	ElemArray     ElemType = "array"
	ElemConst     ElemType = "const"
	ElemEnum      ElemType = "enum"
	ElemGeneric   ElemType = "generic"
	ElemParam     ElemType = "param"
	ElemResult    ElemType = "result"
	ElemField     ElemType = "field"
	ElemReceiver  ElemType = "receiver"
	ElemBinary    ElemType = "binary"
	ElemConstrain ElemType = "constrain"
)
