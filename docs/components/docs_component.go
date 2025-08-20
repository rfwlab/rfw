//go:build js && wasm

package components

import (
	_ "embed"
	"strconv"
	"strings"

	"github.com/rfwlab/rfw/v1/core"
	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/router"
	jst "syscall/js"
)

//go:embed templates/docs_component.rtml
var docsTpl []byte

type DocsComponent struct {
	*core.HTMLComponent
	nav     jst.Value
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
	if home := doc.Call("querySelector", "nav a[href='/']"); home.Truthy() {
		ch := events.Listen("click", home)
		go func() {
			for evt := range ch {
				evt.Call("preventDefault")
				router.Navigate("/")
			}
		}()
	}
	if docs := doc.Call("querySelector", "nav a[href='/docs/index']"); docs.Truthy() {
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
			sidebar := doc.Call("getElementById", "sidebar")
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
			}
			c.docComp = core.NewComponent("DocContent", []byte(html), nil)
			doc.Call("getElementById", "doc-content").Set("innerHTML", c.docComp.Render())
			c.docComp.Mount()
			if hljs := js.Get("hljs"); hljs.Truthy() {
				hljs.Call("highlightAll")
			}

			link := strings.TrimSuffix(strings.TrimPrefix(path, "/docs/"), ".md")
			idx := -1
			for i, p := range c.order {
				if p == link {
					idx = i
					break
				}
			}
			nav := doc.Call("getElementById", "doc-nav")
			nav.Set("innerHTML", "")
			if idx > 0 {
				prev := c.order[idx-1]
				a := doc.Call("createElement", "a")
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
				a := doc.Call("createElement", "a")
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

func (c *DocsComponent) renderSidebar(items jst.Value, parent jst.Value, level int) {
	doc := js.Document()
	length := items.Length()
	for i := 0; i < length; i++ {
		item := items.Index(i)
		title := item.Get("title").String()
		if path := item.Get("path"); path.Truthy() {
			link := strings.TrimSuffix(path.String(), ".md")
			c.titles[link] = title
			c.order = append(c.order, link)
			a := doc.Call("createElement", "a")
			a.Set("href", "/docs/"+link)
			a.Set("textContent", title)
			a.Set("className", "block py-1 pl-"+strconv.Itoa(4*level)+" text-gray-700 hover:text-blue-600")
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
				h := doc.Call("createElement", "div")
				h.Set("textContent", title)
				h.Set("className", "mt-4 mb-1 font-semibold text-gray-900 pl-"+strconv.Itoa(4*level))
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
