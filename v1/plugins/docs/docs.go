//go:build js && wasm

package docs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
	marked "github.com/rfwlab/rfw/v1/js/shim/marked"
)

type Plugin struct {
	Sidebar string
	loader  js.Func
}

func New(sidebar string) *Plugin {
	sidebar = fmt.Sprintf("%s?%s", sidebar, time.Now().Unix())
	return &Plugin{Sidebar: sidebar}
}

func (p *Plugin) Install(a *core.App) {
	doc := js.Document()

	js.Fetch(p.Sidebar).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		res := args[0]
		res.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			js.Set("__rfwDocsSidebar", args[0].String())
			doc.Call("dispatchEvent", js.CustomEvent().New("rfwSidebar"))
			return nil
		}))
		return nil
	}))

	p.loader = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return nil
		}
		path := args[0].String()
		js.Fetch(path).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			res := args[0]
			res.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				content := args[0].String()
				tokens := marked.Lexer(content)
				slug := newSlugger()
				headings := make([]any, 0)
				length := tokens.Length()
				for i := 0; i < length; i++ {
					tok := tokens.Index(i)
					if tok.Get("type").String() == "heading" {
						text := tok.Get("text").String()
						depth := tok.Get("depth").Int()
						id := slug.slug(text)
						headings = append(headings, map[string]any{"text": text, "depth": depth, "id": id})
					}
				}
				doc.Call("dispatchEvent", js.CustomEvent().New("rfwDoc", map[string]any{"detail": map[string]any{"path": path, "content": content, "headings": headings}}))
				return nil
			}))
			return nil
		}))
		return nil
	})
	js.Set("rfwLoadDoc", p.loader)
}

func (p *Plugin) Build(json.RawMessage) error { return nil }

var slugRe = regexp.MustCompile(`[^a-z0-9\s-]`)

type slugger struct {
	seen map[string]int
}

func newSlugger() *slugger {
	return &slugger{seen: make(map[string]int)}
}

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
