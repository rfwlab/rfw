//go:build js && wasm

package components

import (
	"embed"
	"strconv"
	"strings"
	"time"

	"github.com/rfwlab/rfw/v2/composition"
	"github.com/rfwlab/rfw/v2/core"
	dom "github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/events"
	js "github.com/rfwlab/rfw/v2/js"
	markdown "github.com/rfwlab/rfw/v2/markdown"
	highlight "github.com/rfwlab/rfw/v2/plugins/highlight"
	"github.com/rfwlab/rfw/v2/plugins/seo"
	"github.com/rfwlab/rfw/v2/router"
	"github.com/rfwlab/rfw/v2/types"
)

//go:embed templates
var templatesFS embed.FS

func init() {
	composition.RegisterFS(&templatesFS)
}

type DocsComponent struct {
	Page      types.String
	Search    types.String
	SidebarEl *types.Ref
	ContentEl *types.Ref
	TocEl     *types.Ref
	NavEl     *types.Ref
	SearchEl  *types.Ref
	ResultsEl *types.Ref

	hc    *core.HTMLComponent
	order []string
	meta  map[string]struct{ Title, Description string }
}

func NewDocsComponent() *core.HTMLComponent {
	d := &DocsComponent{}
	v, err := composition.New(d)
	if err != nil {
		panic(err)
	}
	d.hc = v
	return v
}

func (d *DocsComponent) OnMount() {
	d.meta = make(map[string]struct{ Title, Description string })
	doc := dom.Doc()

	d.interceptNav(doc)
	d.loadAndWatchSidebar(doc)
	d.setupSearch(doc)
	d.watchDocEvents(doc)

	if d.Page.Get() == "" {
		d.Page.Set("index")
	}
	d.loadArticle(d.Page.Get())
}

func (d *DocsComponent) OnParams(params map[string]string) {
	switch {
	case params == nil:
		d.Page.Set("index")
	case params["section"] != "" && params["page"] != "":
		d.Page.Set(params["section"] + "/" + params["page"])
	case params["page"] != "":
		d.Page.Set(params["page"])
	default:
		d.Page.Set("index")
	}
	d.loadArticle(d.Page.Get())
}

func (d *DocsComponent) loadArticle(page string) {
	js.Get("rfwLoadDoc").Call("call", nil, "/articles/"+page+".md")
}

func (d *DocsComponent) interceptNav(doc dom.Document) {
	for _, sel := range []string{"nav a[href='/']", "nav a[href='/index']"} {
		if el := doc.Query(sel); el.Truthy() {
			unsub := events.On("click", el.Value, func(e js.Value) {
				e.Call("preventDefault")
				router.Navigate("/index")
			})
			_ = unsub
		}
	}
}

func (d *DocsComponent) loadAndWatchSidebar(doc dom.Document) {
	d.tryLoadSidebar(doc)
	unsub := events.OnApp(events.EventSidebarLoaded, func(data any) {
		if items, ok := data.([]map[string]any); ok {
			d.processSidebar(items, doc)
		}
	})
	_ = unsub

	ch := events.Listen("rfwSidebar", doc.Value)
	go func() {
		for range ch {
			d.tryLoadSidebar(doc)
		}
	}()
}

func (d *DocsComponent) tryLoadSidebar(doc dom.Document) {
	sidebarJSON := js.Get("__rfwDocsSidebar")
	if !sidebarJSON.Truthy() {
		return
	}
	nav := js.JSON().Call("parse", sidebarJSON)
	d.order = d.order[:0]
	d.meta = map[string]struct{ Title, Description string }{}
	sidebar := doc.ByID("sidebar")
	sidebar.SetHTML("")
	d.renderSidebar(nav, sidebar, 0)
}

func (d *DocsComponent) processSidebar(items []map[string]any, doc dom.Document) {
	d.order = d.order[:0]
	d.meta = map[string]struct{ Title, Description string }{}
	sidebar := doc.ByID("sidebar")
	sidebar.SetHTML("")
	for _, item := range items {
		d.processSidebarItem(item, sidebar, 0)
	}
}

func (d *DocsComponent) processSidebarItem(item map[string]any, parent dom.Element, level int) {
	title, _ := item["title"].(string)
	desc, _ := item["description"].(string)
	if path, ok := item["path"].(string); ok && path != "" {
		link := strings.TrimSuffix(path, ".md")
		d.meta[link] = struct{ Title, Description string }{Title: title, Description: desc}
		d.order = append(d.order, link)
	}
	if children, ok := item["children"].([]any); ok {
		for _, c := range children {
			if m, ok := c.(map[string]any); ok {
				d.processSidebarItem(m, parent, level+1)
			}
		}
	}
}

func (d *DocsComponent) setupSearch(doc dom.Document) {
	search := doc.ByID("doc-search")
	if !search.Truthy() {
		return
	}
	results := doc.ByID("search-results")

	inputCh := events.Listen("input", search.Value)
	go func() {
		for range inputCh {
			if !d.hc.IsMounted() {
				continue
			}
			d.doSearch(search, results)
		}
	}()

	blurCh := events.Listen("blur", search.Value)
	go func() {
		for range blurCh {
			if !d.hc.IsMounted() {
				continue
			}
			go func() {
				time.Sleep(100 * time.Millisecond)
				results.Get("classList").Call("add", "hidden")
			}()
		}
	}()
}

func (d *DocsComponent) doSearch(search, results dom.Element) {
	q := strings.ToLower(search.Get("value").String())
	results.SetHTML("")
	if q == "" {
		results.Get("classList").Call("add", "hidden")
		return
	}
	count := 0
	for _, link := range d.order {
		title := d.titleFor(link)
		if !strings.Contains(strings.ToLower(title), q) {
			continue
		}
		a := composition.A().
			Href("/"+link).
			Text(title).
			Classes("block", "px-2", "py-1", "text-gray-700", "dark:text-zinc-200", "hover:bg-gray-100", "dark:hover:bg-zinc-700")
		unsub := events.On("mousedown", a.Element().Value, func(e js.Value) {
			e.Call("preventDefault")
			results.SetHTML("")
			results.Get("classList").Call("add", "hidden")
			search.Set("value", "")
			router.Navigate("/" + link)
		})
		_ = unsub
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

func (d *DocsComponent) watchDocEvents(doc dom.Document) {
	ch := events.Listen("rfwDoc", doc.Value)
	go func() {
		for evt := range ch {
			if !d.hc.IsMounted() {
				continue
			}
			d.onArticleLoaded(doc, evt)
		}
	}()
}

func (d *DocsComponent) onArticleLoaded(doc dom.Document, evt js.Value) {
	detail := evt.Get("detail")
	path := detail.Get("path").String()
	content := detail.Get("content").String()
	html := markdown.Parse(content)

	docContent := doc.ByID("doc-content")
	docContent.SetHTML(html)
	highlight.HighlightAll()

	d.applyHeadings(doc, detail)
	d.buildTOC(doc, detail)
	d.updateSEO(detail, path)
	d.buildNav(doc, path)
}

func (d *DocsComponent) applyHeadings(doc dom.Document, detail js.Value) {
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

func (d *DocsComponent) buildTOC(doc dom.Document, detail js.Value) {
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
		events.On("click", a.Element().Value, func(e js.Value) {
			e.Call("preventDefault")
			if el := doc.ByID(id); el.Truthy() {
				el.Call("scrollIntoView", map[string]any{"behavior": "smooth"})
			}
		})
		toc.AppendChild(a.Element())
	}
}

func (d *DocsComponent) updateSEO(detail js.Value, path string) {
	link := strings.TrimSuffix(strings.TrimPrefix(path, "/articles/"), ".md")
	meta := d.meta[link]
	if meta.Title != "" {
		seo.SetTitle(meta.Title)
	} else {
		seo.SetTitle(link)
	}
	seo.SetMeta("description", meta.Description)
}

func (d *DocsComponent) buildNav(doc dom.Document, path string) {
	link := strings.TrimSuffix(strings.TrimPrefix(path, "/articles/"), ".md")
	idx := -1
	for i, p := range d.order {
		if p == link {
			idx = i
			break
		}
	}
	nav := doc.ByID("doc-nav")
	nav.SetHTML("")
	if idx > 0 {
		prev := d.order[idx-1]
		a := composition.A().Classes("text-white").Href("/"+prev).Text("\u2190 " + d.titleFor(prev))
		events.On("click", a.Element().Value, func(e js.Value) {
			e.Call("preventDefault")
			router.Navigate("/" + prev)
		})
		nav.AppendChild(a.Element())
	}
	if idx >= 0 && idx < len(d.order)-1 {
		next := d.order[idx+1]
		a := composition.A().Classes("ml-auto", "text-white").Href("/"+next).Text(d.titleFor(next) + " \u2192")
		events.On("click", a.Element().Value, func(e js.Value) {
			e.Call("preventDefault")
			router.Navigate("/" + next)
		})
		nav.AppendChild(a.Element())
	}
}

func (d *DocsComponent) renderSidebar(items js.Value, parent dom.Element, level int) {
	length := items.Length()
	for i := 0; i < length; i++ {
		item := items.Index(i)
		title := item.Get("title").String()
		desc := ""
		if dd := item.Get("description"); dd.Truthy() {
			desc = dd.String()
		}
		if path := item.Get("path"); path.Truthy() {
			link := strings.TrimSuffix(path.String(), ".md")
			d.meta[link] = struct{ Title, Description string }{Title: title, Description: desc}
			d.order = append(d.order, link)
			a := composition.A().
				Href("/"+link).
				Text(title).
				Classes("block", "py-1", "pl-"+strconv.Itoa(4*level), "text-gray-700", "dark:text-zinc-200", "dark:hover:text-white", "hover:text-black")
			events.On("click", a.Element().Value, func(evt js.Value) {
				evt.Call("preventDefault")
				router.Navigate("/" + link)
			})
			parent.AppendChild(a.Element())
		}
		if children := item.Get("children"); children.Truthy() {
			if !item.Get("path").Truthy() && title != "" {
				h := composition.Div().
					Text(title).
					Classes("mt-4", "mb-1", "font-semibold", "text-gray-900", "dark:text-white", "pl-"+strconv.Itoa(4*level))
				parent.AppendChild(h.Element())
			}
			d.renderSidebar(children, parent, level+1)
		}
	}
}

func (d *DocsComponent) titleFor(path string) string {
	if m, ok := d.meta[path]; ok && m.Title != "" {
		return m.Title
	}
	return path
}