//go:build js && wasm

// Package docs provides documentation loading and rendering support.
package docs

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/events"
	js "github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/markdown"
	"github.com/rfwlab/rfw/v2/plugins/seo"
	"github.com/rfwlab/rfw/v2/state"
)

// SidebarItem describes one documentation navigation entry.
type SidebarItem struct {
	Title       string        `json:"title"`
	Path        string        `json:"path"`
	Description string        `json:"description"`
	Children    []SidebarItem `json:"children"`
}

// ArticleData contains a loaded documentation article.
type ArticleData struct {
	Path     string
	Content  string
	Headings []Heading
}

// Heading describes one article heading.
type Heading struct {
	Text  string
	Depth int
	ID    string
}

// Plugin loads documentation navigation and article content.
type Plugin struct {
	Sidebar     string
	disableSEO  bool
	loader      js.Func
	sidebarData *state.Signal[[]SidebarItem]
	articleData *state.Signal[*ArticleData]
}

// New creates a documentation plugin using the supplied sidebar URL.
func New(sidebar string, disableSEO ...bool) *Plugin {
	sidebar = fmt.Sprintf("%s?%d", sidebar, time.Now().Unix())
	p := &Plugin{Sidebar: sidebar}
	if len(disableSEO) > 0 {
		p.disableSEO = disableSEO[0]
	}
	p.sidebarData = state.NewSignal[[]SidebarItem](nil)
	p.articleData = state.NewSignal[*ArticleData](nil)
	return p
}

// Name returns the plugin name.
func (p *Plugin) Name() string { return "docs" }

// Optional declares the SEO plugin when metadata support is enabled.
func (p *Plugin) Optional() []core.Plugin {
	if p.disableSEO {
		return nil
	}
	return []core.Plugin{seo.New()}
}

// Provide returns the values exposed to component templates.
func (p *Plugin) Provide() map[string]any {
	return map[string]any{
		"sidebar": p.sidebarData,
		"article": p.articleData,
		"loadDoc": p.loadArticle,
	}
}

// Install loads the sidebar and exposes the article loader.
func (p *Plugin) Install(_ *core.App) {
	for k, v := range p.Provide() {
		core.RegisterPluginVar("docs", k, v)
	}

	doc := js.Document()

	js.Fetch(p.Sidebar).Call("then", js.FuncOf(func(_ js.Value, args []js.Value) any {
		res := args[0]
		res.Call("text").Call("then", js.FuncOf(func(_ js.Value, args []js.Value) any {
			raw := args[0].String()
			var items []SidebarItem
			if err := json.Unmarshal([]byte(raw), &items); err == nil {
				p.sidebarData.Set(items)
			}
			js.Set("__rfwDocsSidebar", raw)
			doc.Call("dispatchEvent", js.CustomEvent().New("rfwSidebar"))
			events.EmitApp(events.EventSidebarLoaded, items)
			return nil
		}))
		return nil
	}))

	p.loader = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 1 {
			return nil
		}
		p.loadArticle(args[0].String())
		return nil
	})
	js.Set("rfwLoadDoc", p.loader)
}

func (p *Plugin) loadArticle(path string) {
	js.Fetch(path).Call("then", js.FuncOf(func(_ js.Value, args []js.Value) any {
		res := args[0]
		res.Call("text").Call("then", js.FuncOf(func(_ js.Value, args []js.Value) any {
			content := args[0].String()
			mhs := markdown.Headings(content)
			headings := make([]Heading, len(mhs))
			for i, h := range mhs {
				headings[i] = Heading{Text: h.Text, Depth: h.Depth, ID: h.ID}
			}
			data := &ArticleData{
				Path:     path,
				Content:  content,
				Headings: headings,
			}
			p.articleData.Set(data)

			doc := js.Document()
			doc.Call("dispatchEvent", js.CustomEvent().New("rfwDoc", map[string]any{
				"detail": map[string]any{
					"path":     path,
					"content":  content,
					"headings": headingsToAny(headings),
				},
			}))
			events.EmitApp(events.EventArticleLoaded, data)
			return nil
		}))
		return nil
	}))
}

func headingsToAny(headings []Heading) []any {
	result := make([]any, len(headings))
	for i, h := range headings {
		result[i] = map[string]any{
			"text":  h.Text,
			"depth": h.Depth,
			"id":    h.ID,
		}
	}
	return result
}

// Build accepts the plugin build configuration.
func (p *Plugin) Build(json.RawMessage) error { return nil }
