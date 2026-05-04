package core

import (
	"fmt"
)

func LoadComponentTemplate(templateFs []byte) (string, error) {
	template := string(templateFs)
	if template == "" {
		return "", fmt.Errorf("template is empty")
	}

	return template, nil
}
