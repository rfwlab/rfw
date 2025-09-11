//go:build js && wasm

package components

import (
	_ "embed"
	"strconv"
	"strings"
	"time"

	"github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/router"
)

//go:embed templates/docs_component.rtml
var docsTpl []byte

type DocsComponent struct {
	*core.HTMLComponent
	nav     js.Value
	order   []string
	titles  map[string]string
	page    string
	mounted bool
	docComp *core.HTMLComponent
}

func NewDocsComponent() *DocsComponent {
	c := &DocsComponent{titles: make(map[string]string)}
	c.HTMLComponent = core.NewComponent("DocsComponent", docsTpl, nil).WithLifecycle(c.mount, c.unmount)
	return c
}

func (c *DocsComponent) mount(hc *core.HTMLComponent) {
	c.mounted = true
	doc := js.Document()

	// intercept top nav links to use the router
	if home := dom.Query("nav a[href='/']"); home.Truthy() {
		ch := events.Listen("click", home)
		go func() {
			for evt := range ch {
				evt.Call("preventDefault")
				router.Navigate("/")
			}
		}()
	}
	if docs := dom.Query("nav a[href='/docs/index']"); docs.Truthy() {
		ch := events.Listen("click", docs)
		go func() {
			for evt := range ch {
				evt.Call("preventDefault")
				router.Navigate("/docs/index")
			}
		}()
	}

	// Sidebar data loads asynchronously; listen for the custom event
	// dispatched by the docs plugin and render when available.
	loadSidebar := func() {
		sidebarJSON := js.Get("__rfwDocsSidebar")
		if sidebarJSON.Truthy() {
			c.nav = js.JSON().Call("parse", sidebarJSON)
			c.order = c.order[:0]
			c.titles = map[string]string{}
			sidebar := dom.ByID("sidebar")
			sidebar.Set("innerHTML", "")
			c.renderSidebar(c.nav, sidebar, 0)
		}
	}
	loadSidebar()
	sidebarCh := events.Listen("rfwSidebar", doc)
	go func() {
		for range sidebarCh {
			if !c.mounted {
				continue
			}
			loadSidebar()
		}
	}()

	if search := dom.ByID("doc-search"); search.Truthy() {
		results := dom.ByID("search-results")
		inputCh := events.Listen("input", search)
		go func() {
			for range inputCh {
				if !c.mounted {
					continue
				}
				q := strings.ToLower(search.Get("value").String())
				results.Set("innerHTML", "")
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
					a := dom.CreateElement("a")
					a.Set("href", "/docs/"+link)
					a.Set("textContent", title)
					a.Set("className", "block px-2 py-1 text-gray-700 dark:text-zinc-200 hover:bg-gray-100 dark:hover:bg-zinc-700")
					ch := events.Listen("mousedown", a)
					go func(l string) {
						for e := range ch {
							if !c.mounted {
								continue
							}
							e.Call("preventDefault")
							results.Set("innerHTML", "")
							results.Get("classList").Call("add", "hidden")
							search.Set("value", "")
							router.Navigate("/docs/" + l)
						}
					}(link)
					results.Call("appendChild", a)
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

		blurCh := events.Listen("blur", search)
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

	docCh := events.Listen("rfwDoc", doc)
	go func() {
		for evt := range docCh {
			if !c.mounted {
				continue
			}
			detail := evt.Get("detail")
			path := detail.Get("path").String()
			content := detail.Get("content").String()
			html := js.Get("marked").Call("parse", content).String()
			if c.docComp != nil {
				c.docComp.Unmount()
				delete(c.HTMLComponent.Dependencies, "doc")
			}
			c.docComp = core.NewComponent("DocContent", []byte(html), nil)
			c.HTMLComponent.AddDependency("doc", c.docComp)
			dom.ByID("doc-content").Set("innerHTML", c.docComp.Render())
			c.docComp.Mount()
			if hljs := js.Get("hljs"); hljs.Truthy() {
				hljs.Call("highlightAll")
			}

			headings := detail.Get("headings")
			if headings.Truthy() {
				contentEl := dom.ByID("doc-content")
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

			if toc := dom.ByID("toc"); toc.Truthy() {
				toc.Set("innerHTML", "")
				if headings.Truthy() {
					length := headings.Length()
					for i := 0; i < length; i++ {
						h := headings.Index(i)
						id := h.Get("id").String()
						text := h.Get("text").String()
						depth := h.Get("depth").Int()
						a := dom.CreateElement("a")
						a.Set("href", "#"+id)
						a.Set("textContent", text)
						a.Set("className", "block py-1 pl-"+strconv.Itoa((depth-1)*4)+" text-gray-700 dark:text-zinc-200 dark:hover:text-white hover:text-black")
						ch := events.Listen("click", a)
						go func(i string) {
							for e := range ch {
								e.Call("preventDefault")
								if el := dom.ByID(i); el.Truthy() {
									el.Call("scrollIntoView", map[string]any{"behavior": "smooth"})
								}
							}
						}(id)
						toc.Call("appendChild", a)
					}
				}
			}

			link := strings.TrimSuffix(strings.TrimPrefix(path, "/docs/"), ".md")
			idx := -1
			for i, p := range c.order {
				if p == link {
					idx = i
					break
				}
			}
			nav := dom.ByID("doc-nav")
			nav.Set("innerHTML", "")
			if idx > 0 {
				prev := c.order[idx-1]
				a := dom.CreateElement("a")
				a.Set("className", "text-blue-600")
				a.Set("href", "/docs/"+prev)
				a.Set("textContent", "\u2190 "+c.titleFor(prev))
				ch := events.Listen("click", a)
				go func(p string) {
					for e := range ch {
						e.Call("preventDefault")
						router.Navigate("/docs/" + p)
					}
				}(prev)
				nav.Call("appendChild", a)
			}
			if idx >= 0 && idx < len(c.order)-1 {
				next := c.order[idx+1]
				a := dom.CreateElement("a")
				a.Set("className", "ml-auto text-blue-600")
				a.Set("href", "/docs/"+next)
				a.Set("textContent", c.titleFor(next)+" \u2192")
				ch := events.Listen("click", a)
				go func(n string) {
					for e := range ch {
						e.Call("preventDefault")
						router.Navigate("/docs/" + n)
					}
				}(next)
				nav.Call("appendChild", a)
			}
		}
	}()

	if c.page == "" {
		c.page = "index"
	}
	js.Call("rfwLoadDoc", "/articles/"+c.page+".md")
}

func (c *DocsComponent) SetRouteParams(params map[string]string) {
	switch {
	case params == nil:
		c.page = "index"
	case params["section"] != "" && params["page"] != "":
		c.page = params["section"] + "/" + params["page"]
	case params["page"] != "":
		c.page = params["page"]
	default:
		c.page = "index"
	}
	if c.mounted {
		js.Call("rfwLoadDoc", "/articles/"+c.page+".md")
	}
}

func (c *DocsComponent) unmount(hc *core.HTMLComponent) {
	c.mounted = false
}

func (c *DocsComponent) renderSidebar(items js.Value, parent js.Value, level int) {
	length := items.Length()
	for i := 0; i < length; i++ {
		item := items.Index(i)
		title := item.Get("title").String()
		if path := item.Get("path"); path.Truthy() {
			link := strings.TrimSuffix(path.String(), ".md")
			c.titles[link] = title
			c.order = append(c.order, link)
			a := dom.CreateElement("a")
			a.Set("href", "/docs/"+link)
			a.Set("textContent", title)
			a.Set("className", "block py-1 pl-"+strconv.Itoa(4*level)+" text-gray-700 dark:text-zinc-200 dark:hover:text-white hover:text-black")
			ch := events.Listen("click", a)
			go func(l string) {
				for evt := range ch {
					if !c.mounted {
						continue
					}
					evt.Call("preventDefault")
					router.Navigate("/docs/" + l)
				}
			}(link)
			parent.Call("appendChild", a)
		}
		if children := item.Get("children"); children.Truthy() {
			if !item.Get("path").Truthy() && title != "" {
				h := dom.CreateElement("div")
				h.Set("textContent", title)
				h.Set("className", "mt-4 mb-1 font-semibold text-gray-900 dark:text-white pl-"+strconv.Itoa(4*level))
				parent.Call("appendChild", h)
			}
			c.renderSidebar(children, parent, level+1)
		}
	}
}

func (c *DocsComponent) titleFor(path string) string {
	if t, ok := c.titles[path]; ok {
		return t
	}
	return path
}
