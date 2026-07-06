//go:build !js || !wasm

package docs

import (
	"fmt"
	"regexp"
	"strings"
)

var slugRe = regexp.MustCompile(`[^a-z0-9\s-]`)

type slugger struct {
	seen map[string]int
}

func newSlugger() *slugger { return &slugger{seen: make(map[string]int)} }

func (s *slugger) slug(text string) string {
	slug := strings.ToLower(text)
	slug = slugRe.ReplaceAllString(slug, "")
	slug = strings.TrimSpace(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	if n, ok := s.seen[slug]; ok {
		s.seen[slug] = n + 1
		return fmt.Sprintf("%s-%d", slug, n)
	}
	s.seen[slug] = 1
	return slug
}
