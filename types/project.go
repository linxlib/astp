package types

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

type Project struct {
	ModPkg     string           `json:"mod_pkg,omitempty"`
	BaseDir    string           `json:"base_dir,omitempty"`
	ModName    string           `json:"mod_name,omitempty"`
	ModVersion string           `json:"mod_version,omitempty"`
	ModPath    string           `json:"mod_path,omitempty"`
	SdkPath    string           `json:"sdk_path,omitempty"`
	Timestamp  int64            `json:"timestamp,omitempty"`
	Generator  string           `json:"generator,omitempty"`
	Version    string           `json:"version,omitempty"`
	FileMap    map[string]*File `json:"-"`
	File       []*File          `json:"file,omitempty"`
}

func (p *Project) AddFile(f *File) {
	if p.FileMap == nil {
		p.FileMap = make(map[string]*File)
	}
	p.FileMap[f.KeyHash] = f
}

func (p *Project) Merge(files map[string]*File) {
	for key, file := range files {
		if _, ok := p.FileMap[key]; !ok {
			p.FileMap[key] = file.Clone()
		}
	}
}

func (p *Project) Write(fileName string) error {
	p.File = make([]*File, 0)
	for _, file := range p.FileMap {
		p.File = append(p.File, file)
	}
	// Serialize project to JSON with indentation
	jsonData, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	//if os.Getenv("ASTP_DEBUG") == "true" {
	//
	//}
	err = os.WriteFile(fileName+".json", jsonData, 0644)
	if err != nil {
		return err
	}

	// Create output file
	f, err := os.Create(fileName + ".gz")
	if err != nil {
		return err
	}
	defer f.Close()

	// Create gzip writer
	gz := gzip.NewWriter(f)
	defer gz.Close()

	// Write compressed data
	_, err = gz.Write(jsonData)
	return err
}
func (p *Project) Read(path string) error {
	// Open the gzipped file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a gzip reader
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	// Decode the JSON data into the Project struct
	if err := json.NewDecoder(gz).Decode(p); err != nil {
		return err
	}

	return nil
}
