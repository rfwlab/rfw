package ssr

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rfwlab/rfw/v1/core"
)

// Render creates an HTML string from an RTML template and props using the server-side pipeline.
func Render(template []byte, props map[string]any) (string, error) {
	c := core.NewComponent("root", template, props)
	html := c.Render()
	if props != nil {
		if data, err := json.Marshal(props); err == nil {
			html += fmt.Sprintf(`<script id="__RFWDATA__" type="application/json">%s</script>`, data)
		}
	}
	return html, nil
}

// RenderFile loads an RTML template from disk and renders it.
func RenderFile(path string, props map[string]any) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return Render(data, props)
}
