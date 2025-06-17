package main

import (
	"github.com/linxlib/astp/internal/json"
	"github.com/linxlib/astp/parsers"
	"os"
)

func main() {
	proj := parsers.ParseProj()
	bytes, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile("gen1.json", bytes, 0666)
}
