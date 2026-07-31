//go:build !js

package initproj

import (
	"embed"
)

// TemplatesFS contains the project scaffold embedded in the rfw binary.
//
//go:embed template/**
var TemplatesFS embed.FS
