package gob

import (
	"encoding/gob"
)

var (
	// NewDecoder is exported by gin/json package.
	NewDecoder = gob.NewDecoder
	// NewEncoder is exported by gin/json package.
	NewEncoder = gob.NewEncoder
)
