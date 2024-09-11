package framework

import (
	"fmt"
	"io/fs"
)

type TemplateLoader struct {
	TemplateDir fs.FS
}

func NewTemplateLoader(templateDir fs.FS) *TemplateLoader {
	return &TemplateLoader{
		TemplateDir: templateDir,
	}
}

func (tl *TemplateLoader) LoadComponentTemplate(componentName string) (string, error) {
	filename := fmt.Sprintf("templates/%s.html", componentName)
	content, err := fs.ReadFile(tl.TemplateDir, filename)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", filename, err)
	}
	return string(content), nil
}
