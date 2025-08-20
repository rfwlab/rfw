//go:build js && wasm

package components

import (
	_ "embed"
	"io"
	"net/http"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
)

//go:embed templates/api_integration_component.rtml
var apiIntegrationComponentTpl []byte

func NewAPIIntegrationComponent() *core.HTMLComponent {
	c := core.NewComponent("APIIntegrationComponent", apiIntegrationComponentTpl, nil)
	dom.RegisterHandlerFunc("load", func() {
		go func() {
			data, err := FetchData("https://jsonplaceholder.typicode.com/todos/1")
			if err != nil {
				c.Store.Set("apiData", err.Error())
				return
			}
			c.Store.Set("apiData", data)
		}()
	})

	return c
}

func FetchData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
