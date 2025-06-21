package types

var _ IElem[*File] = (*File)(nil)

type File struct {
	Name      string       `json:"name"`
	Key       string       `json:"-"`
	KeyHash   string       `json:"-"`
	Package   *Package     `json:"package,omitempty"`
	Comment   []*Comment   `json:"comment,omitempty"`
	Import    []*Import    `json:"import,omitempty"`
	Variable  []*Variable  `json:"variable,omitempty"`
	Const     []*Const     `json:"const,omitempty"`
	Function  []*Function  `json:"function,omitempty"`
	Interface []*Interface `json:"interface,omitempty"`
	Struct    []*Struct    `json:"struct,omitempty"`
}

func (f *File) String() string {
	return f.Name
}

func (f *File) Clone() *File {
	if f == nil {
		return nil
	}
	if !deepClone {
		return f
	}
	return &File{
		Name:      f.Name,
		Key:       f.Key,
		KeyHash:   f.KeyHash,
		Package:   f.Package,
		Comment:   f.Comment,
		Import:    f.Import,
		Variable:  CopySlice(f.Variable),
		Const:     CopySlice(f.Const),
		Function:  CopySlice(f.Function),
		Interface: CopySlice(f.Interface),
		Struct:    CopySlice(f.Struct),
	}
}

func (f *File) IsMainPackage() bool {
	return f.Package.Name == "main"
}

func (f *File) FindStruct(keyHash string) *Struct {
	for _, s := range f.Struct {
		if s.KeyHash == keyHash {
			return s
		}
	}

	return nil
}
