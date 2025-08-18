package initproj

import (
	"embed"
)

//go:embed template/*
var TemplatesFS embed.FS
