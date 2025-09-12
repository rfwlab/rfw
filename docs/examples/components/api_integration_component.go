//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/http"
)

//go:embed templates/api_integration_component.rtml
var apiIntegrationComponentTpl []byte

func NewAPIIntegrationComponent() *core.HTMLComponent {
	c := core.NewComponent("APIIntegrationComponent", apiIntegrationComponentTpl, nil)
	dom.RegisterHandlerFunc("load", func() {
		go func() {
			var todo struct {
				Title string `json:"title"`
			}
			if err := http.FetchJSON("https://jsonplaceholder.typicode.com/todos/1", &todo); err != nil {
				c.Store.Set("apiData", err.Error())
				return
			}
			c.Store.Set("apiData", todo.Title)
		}()
	})

	return c
}
