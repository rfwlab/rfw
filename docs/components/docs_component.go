//go:build js && wasm

package components

import (
	_ "embed"
	"strconv"
	"strings"
	"time"

	"github.com/rfwlab/rfw/v2/composition"
	"github.com/rfwlab/rfw/v2/core"
	dom "github.com/rfwlab/rfw/v2/dom"
	events "github.com/rfwlab/rfw/v2/events"
	js "github.com/rfwlab/rfw/v2/js"
	markdown "github.com/rfwlab/rfw/v2/markdown"
	docplug "github.com/rfwlab/rfw/v2/plugins/docs"
	highlight "github.com/rfwlab/rfw/v2/plugins/highlight"
	"github.com/rfwlab/rfw/v2/plugins/seo"
	"github.com/rfwlab/rfw/v2/router"
	"github.com/rfwlab/rfw/v2/types"
)

//go:embed templates/docs_component.rtml
var docsTpl []byte

type DocsComponent struct {
	*core.HTMLComponent
	Page    types.String
	Search  types.String
	nav     js.Value
	order   []string
	meta    map[string]struct{ Title, Description string }
	mounted bool
	SidebarEl *types.Ref
	TocEl     *types.Ref
	ContentEl *types.Ref
	NavEl     *types.Ref
	SearchEl  *types.Ref
	ResultsEl *types.Ref
}

func NewDocsComponent() *DocsComponent {
	c := &DocsComponent{meta: make(map[string]struct{ Title, Description string })}
	c.HTMLComponent = core.NewComponent("DocsComponent", docsTpl, map[string]any{
		"Page":   &c.Page,
		"Search": &c.Search,
	}).WithLifecycle(c.mount, c.unmount)
	c.SetComponent(c)
	return c
}

func (c *DocsComponent) mount(hc *core.HTMLComponent) {
	c.mounted = true
	doc := dom.Doc()

	c.interceptNav(doc)
	c.loadAndWatchSidebar(doc)
	c.setupSearch(doc)
	c.watchDocEvents(doc)

	if c.Page.Get() == "" {
		c.Page.Set("index")
	}
	docplug.LoadArticle("/articles/" + c.Page.Get() + ".md")
}

func (c *DocsComponent) unmount(hc *core.HTMLComponent) {
	c.mounted = false
}

func (c *DocsComponent) SetRouteParams(params map[string]string) {
	switch {
	case params == nil:
		c.Page.Set("index")
	case params["section"] != "" && params["page"] != "":
		c.Page.Set(params["section"] + "/" + params["page"])
	case params["page"] != "":
		c.Page.Set(params["page"])
	default:
		c.Page.Set("index")
	}
	if c.mounted {
		docplug.LoadArticle("/articles/" + c.Page.Get() + ".md")
	}
}

func (c *DocsComponent) interceptNav(doc dom.Document) {
	for _, sel := range []string{"nav a[href='/']", "nav a[href='/index']"} {
		if el := doc.Query(sel); el.Truthy() {
			ch := events.Listen("click", el.Value)
			go func() {
				for evt := range ch {
					evt.Call("preventDefault")
					router.Navigate("/index")
				}
			}()
		}
	}
}

func (c *DocsComponent) loadAndWatchSidebar(doc dom.Document) {
	loadSidebar := func() {
		sidebarJSON := js.Get("__rfwDocsSidebar")
		if !sidebarJSON.Truthy() {
			return
		}
		c.nav = js.JSON().Call("parse", sidebarJSON)
		c.order = c.order[:0]
		c.meta = map[string]struct{ Title, Description string }{}
		sidebar := doc.ByID("sidebar")
		sidebar.SetHTML("")
		c.renderSidebar(c.nav, sidebar, 0)
	}
	loadSidebar()
	ch := events.Listen("rfwSidebar", doc.Value)
	go func() {
		for range ch {
			if !c.mounted {
				continue
			}
			loadSidebar()
		}
	}()
}

func (c *DocsComponent) setupSearch(doc dom.Document) {
	search := doc.ByID("doc-search")
	if !search.Truthy() {
		return
	}
	results := doc.ByID("search-results")

	inputCh := events.Listen("input", search.Value)
	go func() {
		for range inputCh {
			if !c.mounted {
				continue
			}
			q := strings.ToLower(search.Get("value").String())
			results.SetHTML("")
			if q == "" {
				results.Get("classList").Call("add", "hidden")
				continue
			}
			count := 0
			for _, link := range c.order {
				title := c.titleFor(link)
				if !strings.Contains(strings.ToLower(title), q) {
					continue
				}
				a := composition.A().
					Href("/"+link).
					Text(title).
					Classes("block", "px-2", "py-1", "text-gray-700", "dark:text-zinc-200", "hover:bg-gray-100", "dark:hover:bg-zinc-700")
				ch := events.Listen("mousedown", a.Element().Value)
				go func(l string) {
					for e := range ch {
						if !c.mounted {
							continue
						}
						e.Call("preventDefault")
						results.SetHTML("")
						results.Get("classList").Call("add", "hidden")
						search.Set("value", "")
						router.Navigate("/" + l)
					}
				}(link)
				results.AppendChild(a.Element())
				count++
				if count >= 5 {
					break
				}
			}
			if count > 0 {
				results.Get("classList").Call("remove", "hidden")
			} else {
				results.Get("classList").Call("add", "hidden")
			}
		}
	}()

	blurCh := events.Listen("blur", search.Value)
	go func() {
		for range blurCh {
			if !c.mounted {
				continue
			}
			go func() {
				time.Sleep(100 * time.Millisecond)
				results.Get("classList").Call("add", "hidden")
			}()
		}
	}()
}

func (c *DocsComponent) watchDocEvents(doc dom.Document) {
	docCh := events.Listen("rfwDoc", doc.Value)
	go func() {
		for evt := range docCh {
			if !c.mounted {
				continue
			}
		detail := evt.Get("detail")
		path := detail.Get("path").String()
		content := detail.Get("content").String()
		html := markdown.Parse(content)

		docContent := doc.ByID("doc-content")
		docContent.SetHTML(html)
		highlight.HighlightAll()

			c.applyHeadings(doc, detail)
			c.buildTOC(doc, detail)
			c.updateSEO(detail, path)
			c.buildNav(doc, path)
		}
	}()
}

func (c *DocsComponent) applyHeadings(doc dom.Document, detail js.Value) {
	headings := detail.Get("headings")
	if !headings.Truthy() {
		return
	}
	contentEl := doc.ByID("doc-content")
	idxByDepth := map[int]int{}
	length := headings.Length()
	for i := 0; i < length; i++ {
		h := headings.Index(i)
		depth := h.Get("depth").Int()
		id := h.Get("id").String()
		tag := "h" + strconv.Itoa(depth)
		nodes := contentEl.Call("getElementsByTagName", tag)
		idx := idxByDepth[depth]
		if el := nodes.Index(idx); el.Truthy() {
			el.Set("id", id)
		}
		idxByDepth[depth] = idx + 1
	}
}

func (c *DocsComponent) buildTOC(doc dom.Document, detail js.Value) {
	toc := doc.ByID("toc")
	if !toc.Truthy() {
		return
	}
	toc.SetHTML("")
	headings := detail.Get("headings")
	if !headings.Truthy() {
		return
	}
	length := headings.Length()
	for i := 0; i < length; i++ {
		h := headings.Index(i)
		id := h.Get("id").String()
		text := h.Get("text").String()
		depth := h.Get("depth").Int()
		a := composition.A().
			Href("#"+id).
			Text(text).
			Classes("block", "py-1", "pl-"+strconv.Itoa((depth-1)*4), "text-gray-700", "dark:text-zinc-200", "dark:hover:text-white", "hover:text-black")
		ch := events.Listen("click", a.Element().Value)
		go func(i string) {
			for e := range ch {
				e.Call("preventDefault")
				if el := doc.ByID(i); el.Truthy() {
					el.Call("scrollIntoView", map[string]any{"behavior": "smooth"})
				}
			}
		}(id)
		toc.AppendChild(a.Element())
	}
}

func (c *DocsComponent) updateSEO(detail js.Value, path string) {
	link := strings.TrimSuffix(strings.TrimPrefix(path, "/articles/"), ".md")
	meta := c.meta[link]
	if meta.Title != "" {
		seo.SetTitle(meta.Title)
	} else {
		seo.SetTitle(link)
	}
	seo.SetMeta("description", meta.Description)
}

func (c *DocsComponent) buildNav(doc dom.Document, path string) {
	link := strings.TrimSuffix(strings.TrimPrefix(path, "/articles/"), ".md")
	idx := -1
	for i, p := range c.order {
		if p == link {
			idx = i
			break
		}
	}
	nav := doc.ByID("doc-nav")
	nav.SetHTML("")
	if idx > 0 {
		prev := c.order[idx-1]
		a := composition.A().Classes("text-white").Href("/"+prev).Text("\u2190 " + c.titleFor(prev))
		ch := events.Listen("click", a.Element().Value)
		go func(p string) {
			for e := range ch {
				e.Call("preventDefault")
				router.Navigate("/" + p)
			}
		}(prev)
		nav.AppendChild(a.Element())
	}
	if idx >= 0 && idx < len(c.order)-1 {
		next := c.order[idx+1]
		a := composition.A().Classes("ml-auto", "text-white").Href("/"+next).Text(c.titleFor(next) + " \u2192")
		ch := events.Listen("click", a.Element().Value)
		go func(n string) {
			for e := range ch {
				e.Call("preventDefault")
				router.Navigate("/" + n)
			}
		}(next)
		nav.AppendChild(a.Element())
	}
}

func (c *DocsComponent) renderSidebar(items js.Value, parent dom.Element, level int) {
	length := items.Length()
	for i := 0; i < length; i++ {
		item := items.Index(i)
		title := item.Get("title").String()
		desc := ""
		if d := item.Get("description"); d.Truthy() {
			desc = d.String()
		}
		if path := item.Get("path"); path.Truthy() {
			link := strings.TrimSuffix(path.String(), ".md")
			c.meta[link] = struct{ Title, Description string }{Title: title, Description: desc}
			c.order = append(c.order, link)
			a := composition.A().
				Href("/"+link).
				Text(title).
				Classes("block", "py-1", "pl-"+strconv.Itoa(4*level), "text-gray-700", "dark:text-zinc-200", "dark:hover:text-white", "hover:text-black")
			ch := events.Listen("click", a.Element().Value)
			go func(l string) {
				for evt := range ch {
					if !c.mounted {
						continue
					}
					evt.Call("preventDefault")
					router.Navigate("/" + l)
				}
			}(link)
			parent.AppendChild(a.Element())
		}
		if children := item.Get("children"); children.Truthy() {
			if !item.Get("path").Truthy() && title != "" {
				h := composition.Div().
					Text(title).
					Classes("mt-4", "mb-1", "font-semibold", "text-gray-900", "dark:text-white", "pl-"+strconv.Itoa(4*level))
				parent.AppendChild(h.Element())
			}
			c.renderSidebar(children, parent, level+1)
		}
	}
}

func (c *DocsComponent) titleFor(path string) string {
	if m, ok := c.meta[path]; ok && m.Title != "" {
		return m.Title
	}
	return path
}