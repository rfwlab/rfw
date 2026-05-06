//go:build js && wasm

package docs

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/events"
	js "github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/markdown"
	"github.com/rfwlab/rfw/v2/plugins/seo"
	"github.com/rfwlab/rfw/v2/state"
)

type SidebarItem struct {
	Title       string         `json:"title"`
	Path        string         `json:"path"`
	Description string         `json:"description"`
	Children    []SidebarItem  `json:"children"`
}

type ArticleData struct {
	Path      string
	Content   string
	Headings  []Heading
}

type Heading struct {
	Text  string
	Depth int
	ID    string
}

type Plugin struct {
	Sidebar      string
	disableSEO  bool
	loader       js.Func
	sidebarData  *state.Signal[[]SidebarItem]
	articleData  *state.Signal[*ArticleData]
	sidebarOnce  sync.Once
}

func New(sidebar string, disableSEO ...bool) *Plugin {
	sidebar = fmt.Sprintf("%s?%s", sidebar, time.Now().Unix())
	p := &Plugin{Sidebar: sidebar}
	if len(disableSEO) > 0 {
		p.disableSEO = disableSEO[0]
	}
	p.sidebarData = state.NewSignal[[]SidebarItem](nil)
	p.articleData = state.NewSignal[*ArticleData](nil)
	return p
}

func (p *Plugin) Name() string { return "docs" }

func (p *Plugin) Optional() []core.Plugin {
	if p.disableSEO {
		return nil
	}
	return []core.Plugin{seo.New()}
}

func (p *Plugin) Provide() map[string]any {
	return map[string]any{
		"sidebar":  p.sidebarData,
		"article":  p.articleData,
		"loadDoc":  p.loadArticle,
	}
}

func (p *Plugin) Install(a *core.App) {
	for k, v := range p.Provide() {
		core.RegisterPluginVar("docs", k, v)
	}

	doc := js.Document()

	js.Fetch(p.Sidebar).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		res := args[0]
		res.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
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

	p.loader = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return nil
		}
		p.loadArticle(args[0].String())
		return nil
	})
	js.Set("rfwLoadDoc", p.loader)
}

func (p *Plugin) loadArticle(path string) {
	js.Fetch(path).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		res := args[0]
		res.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
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

func (p *Plugin) Build(json.RawMessage) error { return nil }