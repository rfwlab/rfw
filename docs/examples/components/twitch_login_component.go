//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"
	"net/url"

	rfwenv "github.com/rfwlab/rfw/docs/rfwenv"
	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/twitch_login_component.rtml
var twitchLoginTpl []byte

func NewTwitchLoginComponent() *core.HTMLComponent {
	clientID := rfwenv.Get("TWITCH_CLIENT_ID")
	redirectURI := "https://localhost:8081/examples/twitch/callback"
	scope := "user:read:email"
	authURL := fmt.Sprintf("https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", url.QueryEscape(clientID), url.QueryEscape(redirectURI), url.QueryEscape(scope))
	props := map[string]any{"authURL": authURL}
	return core.NewComponent("TwitchLoginComponent", twitchLoginTpl, props)
}
