//go:build js && wasm

package main

import (
	"embed"

	"github.com/rfwlab/rfw/demo/components"
	"github.com/rfwlab/rfw/demo/pages"
	"github.com/rfwlab/rfw/v2/composition"
	"github.com/rfwlab/rfw/v2/router"
	t "github.com/rfwlab/rfw/v2/types"
)

//go:embed pages/templates components/templates
var templates embed.FS

func init() {
	composition.RegisterFS(&templates)
}

func main() {
	composition.SetDevMode(true)

	router.SetNavItems([]router.NavItem{
		{Path: "/", Label: "Home", Meta: map[string]any{
			"activeClass":   "border-indigo-500 text-gray-900",
			"inactiveClass": "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700",
		}},
		{Path: "/about", Label: "About", Meta: map[string]any{
			"activeClass":   "border-indigo-500 text-gray-900",
			"inactiveClass": "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700",
		}},
		{Path: "/contact", Label: "Contact", Meta: map[string]any{
			"activeClass":   "border-indigo-500 text-gray-900",
			"inactiveClass": "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700",
		}},
		{Path: "/custom", Label: "Custom", Meta: map[string]any{
			"activeClass":   "border-orange-500 text-gray-900",
			"inactiveClass": "border-transparent text-gray-500 hover:border-orange-300 hover:text-orange-700",
		}},
	})

	router.Page("/", func() *t.View {
		return components.NewLayout(pages.NewHomePage())
	})
	router.Page("/about", func() *t.View {
		return components.NewLayout(pages.NewAboutPage())
	})
	router.Page("/contact", func() *t.View {
		return components.NewLayout(pages.NewContactPage())
	})
	router.Page("/custom", func() *t.View {
		return components.NewLayout(pages.NewCustomPage())
	})

	router.InitRouter()
	select {}
}