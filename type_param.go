package astp

type TypeParam struct {
	Name        string //泛型名
	PackagePath string
	TypeName    string  //泛型约束名
	Type        *Struct //泛型约束
	ActualType  *Struct //实际类型
}

func (t *TypeParam) Clone() *TypeParam {
	return &TypeParam{
		Name:        t.Name,
		PackagePath: t.PackagePath,
		TypeName:    t.TypeName,
		Type:        t.Type.Clone(),
		ActualType:  t.ActualType.Clone(),
	}
}
func (t *TypeParam) GetName() string {
	return t.Name
}

func (t *TypeParam) SetType(namer *Struct) {
	t.Type = namer.Clone()
}

func (t *TypeParam) SetInnerType(b bool) {

}

func (t *TypeParam) SetIsStruct(b bool) {

}

func (t *TypeParam) SetTypeString(s string) {
	t.TypeName = s
}

func (t *TypeParam) SetPointer(b bool) {

}

func (t *TypeParam) SetPrivate(b bool) {

}

func (t *TypeParam) SetSlice(b bool) {

}

func (t *TypeParam) SetPackagePath(s string) {
	t.PackagePath = s
}

func (t *TypeParam) GetType() *Struct {
	return t.Type
}
