package rtml

import (
	"fmt"
	"regexp"
	"strings"
)

// Dependency represents a renderable component used for includes.
type Dependency interface {
	Render() string
}

// Context carries rendering data for templates.
type Context struct {
	Props        map[string]any
	Slots        map[string]any
	Dependencies map[string]Dependency
}

// Replace runs a minimal RTML rendering pipeline for server-side rendering.
func Replace(template string, ctx Context) string {
	rendered := replacePropPlaceholders(template, ctx)
	rendered = replaceIncludePlaceholders(ctx, rendered)
	rendered = replaceSlotPlaceholders(rendered, ctx)
	return rendered
}

func replacePropPlaceholders(template string, ctx Context) string {
	if ctx.Props == nil {
		return template
	}
	for k, v := range ctx.Props {
		pattern := fmt.Sprintf(`{{\s*%s\s*}}`, regexp.QuoteMeta(k))
		re := regexp.MustCompile(pattern)
		template = re.ReplaceAllString(template, fmt.Sprintf("%v", v))
	}
	return template
}

func replaceSlotPlaceholders(template string, ctx Context) string {
	if ctx.Slots == nil {
		return template
	}
	for name, content := range ctx.Slots {
		placeholder := fmt.Sprintf("@slot:%s", name)
		template = strings.ReplaceAll(template, placeholder, fmt.Sprintf("%v", content))
	}
	return template
}

func replaceIncludePlaceholders(ctx Context, template string) string {
	if ctx.Dependencies == nil {
		return template
	}
	for name, dep := range ctx.Dependencies {
		include := fmt.Sprintf("@include:%s", name)
		template = strings.ReplaceAll(template, include, dep.Render())
	}
	return template
}
