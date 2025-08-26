//go:build js && wasm

package components

import (
	_ "embed"
	"net/url"

	core "github.com/rfwlab/rfw/v1/core"
	hostclient "github.com/rfwlab/rfw/v1/hostclient"
	js "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/twitch_callback_component.rtml
var twitchCallbackTpl []byte

func NewTwitchCallbackComponent() *core.HTMLComponent {
	c := core.NewComponent("TwitchCallbackComponent", twitchCallbackTpl, nil)
	c.AddHostComponent("TwitchOAuthHost")

	c.SetOnMount(func(_ *core.HTMLComponent) {
		search := js.Location().Get("search").String()
		if len(search) > 1 {
			if vals, err := url.ParseQuery(search[1:]); err == nil {
				if code := vals.Get("code"); code != "" {
					hostclient.Send("TwitchOAuthHost", map[string]any{"code": code})
				}
			}
		}
	})

	return c
}
