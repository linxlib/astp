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
	IsGeneric   bool
	TypeParams  []*TypeParam
	method      any
}

func (m *Method) Clone() *Method {
	m1 := &Method{
		Receiver:    m.Receiver.Clone(),
		Index:       m.Index,
		PackagePath: m.PackagePath,
		Name:        m.Name,
		Private:     m.Private,
		Signature:   m.Signature,
		Docs:        m.Docs,
		Comments:    m.Comments,
		//Params:      m.Params,
		//Results:     m.Results,
		IsGeneric: m.IsGeneric,
		//TypeParams:  m.TypeParams,
		//method:      m.method,
	}
	m1.Params = make([]*ParamField, len(m.Params))
	m1.Results = make([]*ParamField, len(m.Results))
	m1.TypeParams = make([]*TypeParam, len(m.TypeParams))
	for i, param := range m.Params {
		m1.Params[i] = param.Clone()
	}
	for i, param := range m.Results {
		m1.Results[i] = param.Clone()
	}
	//copy(m1.Params, m.Params)
	//copy(m1.Results, m.Results)
	copy(m1.TypeParams, m.TypeParams)
	return m1
}
func (m *Method) SetActualType(name string, as *Struct) {
	if m.IsGeneric && len(m.TypeParams) > 0 {
		for _, param := range m.TypeParams {
			if param.Name != name {
				param.ActualType = as.Clone()
			}
		}
	}
}
func (m *Method) IsGenericType(name string) bool {
	for _, param := range m.TypeParams {
		if param.Name == name {
			return true
		}
	}
	return false
}
func (m *Method) SetParamType(field *ParamField, as *Struct) {
	if as != nil {
		field.Type = as.Clone()
		field.PackagePath = field.Type.PackagePath
	}

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
