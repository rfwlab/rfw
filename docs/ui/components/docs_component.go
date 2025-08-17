//go:build js && wasm

package components

import (
	_ "embed"
	"strconv"
	"strings"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/router"
	jst "syscall/js"
)

//go:embed templates/docs_component.rtml
var docsTpl []byte

type DocsComponent struct {
	*core.HTMLComponent
	listener        jst.Func
	sidebarListener jst.Func
	nav             jst.Value
	order           []string
	titles          map[string]string
	links           []jst.Func
	page            string
	mounted         bool
}

func NewDocsComponent() *DocsComponent {
	c := &DocsComponent{HTMLComponent: core.NewHTMLComponent("DocsComponent", docsTpl, nil), titles: make(map[string]string)}
	c.SetComponent(c)
	c.Init(nil)
	return c
}

func (c *DocsComponent) OnMount() {
	c.mounted = true
	doc := jst.Global().Get("document")

	// intercept top nav links to use the router
	if home := doc.Call("querySelector", "nav a[href='/']"); home.Truthy() {
		handler := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
			args[0].Call("preventDefault")
			router.Navigate("/")
			return nil
		})
		home.Call("addEventListener", "click", handler)
		c.links = append(c.links, handler)
	}
	if docs := doc.Call("querySelector", "nav a[href='/docs/index']"); docs.Truthy() {
		handler := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
			args[0].Call("preventDefault")
			router.Navigate("/docs/index")
			return nil
		})
		docs.Call("addEventListener", "click", handler)
		c.links = append(c.links, handler)
	}

	// Sidebar data loads asynchronously; listen for the custom event
	// dispatched by the docs plugin and render when available.
	c.sidebarListener = jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		sidebarJSON := jst.Global().Get("__rfwDocsSidebar")
		if sidebarJSON.Truthy() {
			c.nav = jst.Global().Get("JSON").Call("parse", sidebarJSON)
			c.order = c.order[:0]
			c.titles = map[string]string{}
			sidebar := doc.Call("getElementById", "sidebar")
			sidebar.Set("innerHTML", "")
			c.renderSidebar(c.nav, sidebar, 0)
		}
		return nil
	})
	jst.Global().Call("addEventListener", "rfwSidebar", c.sidebarListener)
	if jst.Global().Get("__rfwDocsSidebar").Truthy() {
		c.sidebarListener.Invoke()
	}

	c.listener = jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		detail := args[0].Get("detail")
		path := detail.Get("path").String()
		content := detail.Get("content").String()
		doc.Call("getElementById", "doc-content").Set("innerHTML", jst.Global().Get("marked").Call("parse", content))
		if hljs := jst.Global().Get("hljs"); hljs.Truthy() {
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
			handler := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
				args[0].Call("preventDefault")
				router.Navigate("/docs/" + prev)
				return nil
			})
			a.Call("addEventListener", "click", handler)
			nav.Call("appendChild", a)
			c.links = append(c.links, handler)
		}
		if idx >= 0 && idx < len(c.order)-1 {
			next := c.order[idx+1]
			a := doc.Call("createElement", "a")
			a.Set("className", "ml-auto text-blue-600")
			a.Set("href", "/docs/"+next)
			a.Set("textContent", c.titleFor(next)+" \u2192")
			handler := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
				args[0].Call("preventDefault")
				router.Navigate("/docs/" + next)
				return nil
			})
			a.Call("addEventListener", "click", handler)
			nav.Call("appendChild", a)
			c.links = append(c.links, handler)
		}
		return nil
	})
	jst.Global().Call("addEventListener", "rfwDoc", c.listener)

	if c.page == "" {
		c.page = "index"
	}
	jst.Global().Call("rfwLoadDoc", "/docs/"+c.page+".md")
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
		jst.Global().Call("rfwLoadDoc", "/docs/"+c.page+".md")
	}
}

func (c *DocsComponent) OnUnmount() {
	c.mounted = false
	jst.Global().Call("removeEventListener", "rfwDoc", c.listener)
	jst.Global().Call("removeEventListener", "rfwSidebar", c.sidebarListener)
	c.listener.Release()
	c.sidebarListener.Release()
	for _, fn := range c.links {
		fn.Release()
	}
}

func (c *DocsComponent) renderSidebar(items jst.Value, parent jst.Value, level int) {
	doc := jst.Global().Get("document")
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
			handler := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
				args[0].Call("preventDefault")
				router.Navigate("/docs/" + link)
				return nil
			})
			a.Call("addEventListener", "click", handler)
			parent.Call("appendChild", a)
			c.links = append(c.links, handler)
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
