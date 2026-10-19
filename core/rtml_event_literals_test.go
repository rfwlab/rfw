//go:build js && wasm

package core

import (
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

func TestRenderedEventAttributesDoNotRewriteStoreOrPropData(t *testing.T) {
	if dom.ByID("app").IsNull() {
		host := dom.CreateElement("div")
		host.SetAttr("id", "app")
		dom.Doc().Body().AppendChild(host)
	}
	store := state.NewStore("eventliteral", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "eventliteral")
	const url = "ssh://git@127.0.0.1:42601/team/repo.git"
	store.Set("url", url)
	c := NewHTMLComponent("EventLiteral", []byte(`<root><span id="literal-url">@store:app.eventliteral.url</span><a id="literal-link" href="@prop:url">@prop:label</a><button id="literal-button" @on:click.stop:save>Save</button></root>`), map[string]any{"url": url, "label": "@click:fake "})
	c.SetComponent(c)
	c.Init(nil)
	dom.UpdateDOM(c.ID, c.Render())
	c.Mount()
	defer c.Unmount()
	if got := dom.ByID("literal-url").Get("textContent").String(); got != url {
		t.Fatalf("store URL: %q", got)
	}
	if got := dom.ByID("literal-link").Attr("href"); got != url {
		t.Fatalf("prop URL: %q", got)
	}
	if got := dom.ByID("literal-link").Get("textContent").String(); got != "@click:fake " {
		t.Fatalf("prop label: %q", got)
	}
	if got := dom.ByID("literal-button").Attr("data-on-click"); got != "save" {
		t.Fatalf("event handler: %q", got)
	}
	if got := dom.ByID("literal-button").Attr("data-on-click-modifiers"); got != "stop" {
		t.Fatalf("event modifiers: %q", got)
	}
	const next = "ssh://deploy@10.20.30.40:2222/team/repo.git"
	store.Set("url", next)
	waitForRenderFlush()
	if got := dom.ByID("literal-url").Get("textContent").String(); got != next {
		t.Fatalf("updated store URL: %q", got)
	}
}
