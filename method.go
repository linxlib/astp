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
	Params      []*StructField
	Results     []*StructField
	method      any
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
