package types

type Project struct {
	ModPkg     string           `json:"mod_pkg"`
	BaseDir    string           `json:"base_dir"`
	ModName    string           `json:"mod_name"`
	ModVersion string           `json:"mod_version"`
	ModPath    string           `json:"mod_path"`
	SdkPath    string           `json:"sdk_path"`
	Timestamp  int64            `json:"timestamp"`
	Generator  string           `json:"generator"`
	Version    string           `json:"version"`
	File       map[string]*File `json:"file"`
}

func (p *Project) AddFile(f *File) {
	if p.File == nil {
		p.File = make(map[string]*File)
	}
	p.File[f.KeyHash] = f
}

func (p *Project) Merge(files map[string]*File) {
	for key, file := range files {
		if _, ok := p.File[key]; !ok {
			p.File[key] = file.Clone()
		}
	}
}
