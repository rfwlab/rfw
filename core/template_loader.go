package core

import (
	"fmt"
)

// LoadComponentTemplate validates and returns embedded template data.
func LoadComponentTemplate(templateFs []byte) (string, error) {
	template := string(templateFs)
	if template == "" {
		return "", fmt.Errorf("template is empty")
	}

	return template, nil
}
