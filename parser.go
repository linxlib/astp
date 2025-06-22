package astp

import (
	"bufio"
	"errors"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/parsers"
	"github.com/linxlib/astp/types"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Parser struct {
	*types.Project
	startTime time.Time
}

func (p *Parser) Parse() error {
	p.startTime = time.Now()
	if !internal.FileIsExist("main.go") {
		return errors.New("main.go not exist")
	}
	if !internal.FileIsExist("go.mod") {
		return errors.New("go.mod not exist")
	}
	modFile := "go.mod"
	modDir, _ := os.Getwd()
	modPkg := ""
	modPath := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	sdkPath := filepath.Join(os.Getenv("GOROOT"), "src")
	modVersion := ""
	file, _ := os.Open(modFile)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(m, "module") {
			m = strings.TrimPrefix(m, "module")
			m = strings.TrimSpace(m)
			modPkg = m
		}
		if strings.HasPrefix(m, "go") {
			m = strings.TrimPrefix(m, "go")
			m = strings.TrimSpace(m)
			modVersion = m
			break
		}
	}
	_ = file.Close()
	p.Project = &types.Project{
		ModPkg:     modPkg,
		BaseDir:    modDir,
		ModName:    modPkg,
		ModVersion: modVersion,
		ModPath:    modPath,
		SdkPath:    sdkPath,
		Timestamp:  time.Now().Unix(),
		Generator:  "github.com/linxlib/astp",
		Version:    "v0.4",
	}
	slog.Info("parsing project...", "mod", modPkg, "go version", modVersion)
	parsers.ParseFile("main.go", p.Project)
	p.AfterParseProj()
	slog.Info("project parsed.", "elapsed", time.Since(p.startTime).String())
	return nil
}

func (p *Parser) VisitStructByName(name string, filter func(s *types.Struct) bool, handler func(s *types.Struct)) {
	for _, file := range p.FileMap {
		for _, s := range file.Struct {
			if s.Name == name && filter(s) {
				handler(s)
			}
		}
	}
}
