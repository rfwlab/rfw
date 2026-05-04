package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	blackfriday "github.com/russross/blackfriday/v2"
)

// Parse converts Markdown source to HTML.
func Parse(src string) string {
	flags := blackfriday.CommonHTMLFlags &^ blackfriday.Smartypants &^ blackfriday.SmartypantsFractions &^ blackfriday.SmartypantsDashes &^ blackfriday.SmartypantsLatexDashes
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{Flags: flags})
	out := blackfriday.Run(
		[]byte(src),
		blackfriday.WithExtensions(blackfriday.CommonExtensions),
		blackfriday.WithRenderer(renderer),
	)
	return string(out)
}

// Heading represents a parsed heading.
type Heading struct {
	Text  string
	Depth int
	ID    string
}

var slugRe = regexp.MustCompile(`[^a-z0-9\s-]`)

type slugger struct{ seen map[string]int }

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

// Headings extracts headings from Markdown and generates ids.
func Headings(src string) []Heading {
	parser := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions))
	root := parser.Parse([]byte(src))

	headings := []Heading{}
	slug := newSlugger()

	var buf bytes.Buffer
	var collectText func(n *blackfriday.Node)
	collectText = func(n *blackfriday.Node) {
		switch n.Type {
		case blackfriday.Text, blackfriday.Code:
			buf.Write(n.Literal)
		}
		for c := n.FirstChild; c != nil; c = c.Next {
			collectText(c)
		}
	}

	root.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if !entering {
			return blackfriday.GoToNext
		}
		if n.Type == blackfriday.Heading {
			buf.Reset()
			for c := n.FirstChild; c != nil; c = c.Next {
				collectText(c)
				if c.Next != nil {
					buf.WriteByte(' ')
				}
			}
			text := strings.TrimSpace(buf.String())
			level := n.HeadingData.Level
			headings = append(headings, Heading{
				Text:  text,
				Depth: int(level),
				ID:    slug.slug(text),
			})
		}
		return blackfriday.GoToNext
	})

	return headings
}
