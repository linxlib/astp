package types

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

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

func (p *Project) Write(fileName string) error {
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
