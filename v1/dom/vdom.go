//go:build js && wasm

package dom

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"

	js "github.com/rfwlab/rfw/v1/js"
)

// VDOMNode represents a node in a virtual DOM tree derived from an HTML
// template.
type VDOMNode struct {
	Tag        string
	Attributes map[string]string
	Children   []*VDOMNode
	Text       string
}

// NewVDOM parses an HTML template into a virtual DOM tree.
func NewVDOM(htmlTemplate string) (*VDOMNode, error) {
	reader := strings.NewReader(htmlTemplate)
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	root := mapHTMLNode(doc)
	return root, nil
}

func mapHTMLNode(n *html.Node) *VDOMNode {
	if n.Type == html.TextNode {
		return &VDOMNode{
			Text: n.Data,
		}
	}

	if n.Type == html.DocumentNode || n.Type == html.ElementNode {
		node := &VDOMNode{
			Tag:        n.Data,
			Attributes: mapAttributes(n),
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			node.Children = append(node.Children, mapHTMLNode(child))
		}

		return node
	}

	return nil
}

func mapAttributes(n *html.Node) map[string]string {
	attrs := make(map[string]string)
	for _, attr := range n.Attr {
		attrs[attr.Key] = attr.Val
	}
	return attrs
}

func printVDOM(node *VDOMNode, indent string) {
	if node == nil {
		return
	}

	if node.Tag != "" {
		fmt.Printf("%s<Tag: %s, Attributes: %v>\n", indent, node.Tag, node.Attributes)
	}

	if node.Text != "" {
		fmt.Printf("%s<Text: %s>\n", indent, node.Text)
	}

	for _, child := range node.Children {
		printVDOM(child, indent+"  ")
	}
}

type eventListener struct {
	element js.Value
	event   string
	handler js.Func
}

var listeners = make(map[string][]eventListener)

func parseModifiers(attr string) map[string]bool {
	mods := make(map[string]bool)
	if attr == "" {
		return mods
	}
	for _, m := range strings.Split(attr, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			mods[m] = true
		}
	}
	return mods
}

func nodeByPath(root js.Value, path []int) js.Value {
	node := root
	for _, idx := range path {
		children := node.Get("children")
		if idx >= children.Length() {
			return js.Null()
		}
		node = children.Index(idx)
	}
	return node
}

// BindEventListeners attaches registered handlers to nodes within the component
// identified by componentID.
func BindEventListeners(componentID string, root js.Value) {
	if bs, ok := compiledBindings[componentID]; ok {
		for _, b := range bs {
			node := nodeByPath(root, b.Path)
			handler := GetHandler(b.Handler)
			if !handler.Truthy() {
				continue
			}
			mods := make(map[string]bool)
			for _, m := range b.Modifiers {
				mods[m] = true
			}
			wrapped := js.FuncOf(func(this js.Value, args []js.Value) any {
				if mods["stopPropagation"] && len(args) > 0 {
					args[0].Call("stopPropagation")
				}
				if mods["preventDefault"] && len(args) > 0 {
					args[0].Call("preventDefault")
				}
				anyArgs := make([]any, len(args))
				for i, a := range args {
					anyArgs[i] = a
				}
				handler.Invoke(anyArgs...)
				return nil
			})
			if mods["once"] {
				opts := js.NewDict()
				opts.Set("once", true)
				node.Call("addEventListener", b.Event, wrapped, opts.Value)
			} else {
				node.Call("addEventListener", b.Event, wrapped)
			}
			listeners[componentID] = append(listeners[componentID], eventListener{node, b.Event, wrapped})
		}
		return
	}

	nodes := root.Call("querySelectorAll", "*")
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		attrs := node.Call("getAttributeNames")
		for j := 0; j < attrs.Length(); j++ {
			name := attrs.Index(j).String()
			if strings.HasPrefix(name, "data-on-") && !strings.HasSuffix(name, "-modifiers") {
				event := strings.TrimPrefix(name, "data-on-")
				handlerName := node.Call("getAttribute", name).String()
				modsAttr := node.Call("getAttribute", fmt.Sprintf("data-on-%s-modifiers", event)).String()
				modifiers := parseModifiers(modsAttr)
				handler := GetHandler(handlerName)
				if handler.Truthy() {
					wrapped := js.FuncOf(func(this js.Value, args []js.Value) any {
						if modifiers["stopPropagation"] && len(args) > 0 {
							args[0].Call("stopPropagation")
						}
						if modifiers["preventDefault"] && len(args) > 0 {
							args[0].Call("preventDefault")
						}
						anyArgs := make([]any, len(args))
						for i, a := range args {
							anyArgs[i] = a
						}
						handler.Invoke(anyArgs...)
						return nil
					})
					if modifiers["once"] {
						opts := js.NewDict()
						opts.Set("once", true)
						node.Call("addEventListener", event, wrapped, opts.Value)
					} else {
						node.Call("addEventListener", event, wrapped)
					}
					listeners[componentID] = append(listeners[componentID], eventListener{node, event, wrapped})
				}
			}
		}
	}
}

// RemoveEventListeners detaches all event listeners associated with the
// specified component.
func RemoveEventListeners(componentID string) {
	if ls, ok := listeners[componentID]; ok {
		for _, l := range ls {
			l.element.Call("removeEventListener", l.event, l.handler)
			l.handler.Release()
		}
		delete(listeners, componentID)
	}
	delete(compiledBindings, componentID)
}
