package astp

// Method 结构体方法
type Method struct {
	Receiver    *Receiver
	Index       int
	PackagePath string
	Name        string
	Private     bool
	Signature   string
	Docs        []string
	Comments    string
	Params      []*ParamField
	Results     []*ParamField
	method      any
}

func (m *Method) SetType(namer *Struct) {

}

func (m *Method) SetInnerType(b bool) {

}

func (m *Method) SetIsStruct(b bool) {

}

func (m *Method) SetTypeString(s string) {

}

func (m *Method) SetPointer(b bool) {

}

func (m *Method) SetPrivate(b bool) {

}

func (m *Method) SetSlice(b bool) {

}

func (m *Method) SetPackagePath(s string) {

}

func (m *Method) GetType() *Struct {
	return nil
}

func (m *Method) GetMethod() any {
	return m.method
}
func (m *Method) SetMethod(method any) {
	m.method = method
}

func (m *Method) GetName() string {
	return m.Name
}
