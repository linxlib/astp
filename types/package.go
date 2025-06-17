package types

import "github.com/linxlib/astp/constants"

var _ IElem[*Package] = (*Package)(nil)

type Package struct {
	FileName string                `json:"file_name"`
	FilePath string                `json:"file_path"`
	Name     string                `json:"name"`
	Path     string                `json:"path"`
	Type     constants.PackageType `json:"type"`
}

func (p *Package) IsThis() bool {
	return p.Type == constants.PackageSamePackage
}

func (p *Package) Clone() *Package {
	if p == nil {
		return nil
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
