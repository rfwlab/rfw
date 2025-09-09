//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/http"
)

//go:embed templates/fetchjson_component.rtml
var fetchJSONComponentTpl []byte

func NewFetchJSONComponent() *core.HTMLComponent {
	c := core.NewComponent("FetchJSONComponent", fetchJSONComponentTpl, nil)
	dom.RegisterHandlerFunc("fetch", func() {
		go func() {
			title, err := FetchTodo("https://jsonplaceholder.typicode.com/todos/1")
			if err != nil {
				c.Store.Set("apiData", err.Error())
				return
			}
			c.Store.Set("apiData", title)
		}()
	})
	return c
}

func FetchTodo(url string) (string, error) {
	var todo struct {
		Title string `json:"title"`
	}
	if err := http.FetchJSON(url, &todo); err != nil {
		return "", err
	}
	return todo.Title, nil
}
