package astp

// Interface 接口
type Interface struct {
	Name        string
	Methods     []*Method
	Constraints []string
}
