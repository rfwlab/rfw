//go:build js && wasm

package dom

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type VDOMNode struct {
	Tag        string
	Attributes map[string]string
	Children   []*VDOMNode
	Text       string
}

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
