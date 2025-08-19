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

type APIIntegrationComponent struct {
	*core.HTMLComponent
}

func NewAPIIntegrationComponent() *APIIntegrationComponent {
	c := &APIIntegrationComponent{}
	c.HTMLComponent = core.NewComponentWith("APIIntegrationComponent", apiIntegrationComponentTpl, nil, c)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "API Integration",
	})
	c.AddDependency("header", headerComponent)

	dom.RegisterHandlerFunc("load", c.Load)

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

func (c *APIIntegrationComponent) Load() {
        go func() {
                data, err := FetchData("https://jsonplaceholder.typicode.com/todos/1")
                if err != nil {
                        c.Store.Set("apiData", err.Error())
                        return
                }
                c.Store.Set("apiData", data)
        }()
}
