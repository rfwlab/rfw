//go:build js && wasm

package components

import (
	_ "embed"
	"net/url"
	jst "syscall/js"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	hostclient "github.com/rfwlab/rfw/v1/hostclient"
)

//go:embed templates/twitch_callback_component.rtml
var twitchCallbackTpl []byte

func NewTwitchCallbackComponent() *core.HTMLComponent {
	c := core.NewComponent("TwitchCallbackComponent", twitchCallbackTpl, nil)
	c.AddHostComponent("TwitchOAuthHost")

	dom.RegisterHandlerFunc("load", func() {
		loc := jst.Global().Get("location")
		search := loc.Get("search").String()
		if len(search) > 1 {
			vals, err := url.ParseQuery(search[1:])
			if err == nil {
				code := vals.Get("code")
				if code != "" {
					hostclient.Send("TwitchOAuthHost", map[string]any{"code": code})
				}
			}
		}
	})

	return c
}
