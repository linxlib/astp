package types

import "github.com/linxlib/astp/constants"

var _ IElem[*Package] = (*Package)(nil)

type Package struct {
	FileName string                `json:"file_name,omitempty"`
	FilePath string                `json:"file_path,omitempty"`
	Name     string                `json:"name,omitempty"`
	Path     string                `json:"path,omitempty"`
	Type     constants.PackageType `json:"type"`
}

func (p *Package) IsThis() bool {
	return p.Type == constants.PackageSamePackage
}

func (p *Package) Clone() *Package {
	if p == nil {
		return nil
	}
	if !deepClone {
		return p
	}
	return &Package{
		FileName: p.FileName,
		FilePath: p.FilePath,
		Name:     p.Name,
		Path:     p.Path,
		Type:     p.Type,
	}
}

func (p *Package) String() string {
	return p.Path
}
