//go:build js && wasm

package hostclient

import (
	"context"
	"errors"
	"strings"
	"testing"

	js "github.com/rfwlab/rfw/v2/js"
)

func installStreamBusCredentialFetch(t *testing.T, body string) js.Value {
	t.Helper()
	previous := js.Get("fetch")
	fixture := js.Get("Function").New(`
		const state = {calls: 0};
		state.fetch = (url, init) => {
			state.calls++; state.url = url; state.init = init;
			` + body + `
		};
		return state;
	`).Invoke()
	js.Set("fetch", fixture.Get("fetch"))
	t.Cleanup(func() { js.Set("fetch", previous) })
	return fixture
}

func TestStreamBusCredentialUsesAuthenticatedPostWithoutCookieAccess(t *testing.T) {
	fixture := installStreamBusCredentialFetch(t, `return Promise.resolve({status:200,text:()=>Promise.resolve(JSON.stringify({token:'a'.repeat(64)}))});`)
	token, err := loadStreamBusCredential(context.Background())
	if err != nil || len(token) != 64 {
		t.Fatalf("bootstrap result: token bytes=%d error=%v", len(token), err)
	}
	init := fixture.Get("init")
	if fixture.Get("url").String() != "/__rfw/streambus-auth" || init.Get("method").String() != "POST" {
		t.Fatal("bootstrap endpoint or method changed")
	}
	if init.Get("headers").Get("Content-Type").String() != "application/json" || init.Get("headers").Get("Cookie").Truthy() || init.Get("body").Truthy() {
		t.Fatal("bootstrap request exposed cookies or changed its empty JSON request")
	}
}

func TestStreamBusCredentialHidesRejectedResponseDetails(t *testing.T) {
	installStreamBusCredentialFetch(t, `return Promise.resolve({status:401,text:()=>Promise.resolve('private rejection detail')});`)
	if _, err := loadStreamBusCredential(context.Background()); err == nil || strings.Contains(err.Error(), "private rejection") {
		t.Fatalf("bootstrap error = %v", err)
	}
}

func TestStreamBusCredentialRefusesMalformedToken(t *testing.T) {
	installStreamBusCredentialFetch(t, `return Promise.resolve({status:200,text:()=>Promise.resolve('{"token":"short"}')});`)
	if _, err := loadStreamBusCredential(context.Background()); err == nil {
		t.Fatal("malformed bootstrap credential accepted")
	}
}

func TestStreamBusCredentialCancellationDoesNotStartFetch(t *testing.T) {
	fixture := installStreamBusCredentialFetch(t, `return Promise.reject(new Error('unexpected fetch'));`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := loadStreamBusCredential(ctx)
	if !errors.Is(err, context.Canceled) || fixture.Get("calls").Int() != 0 {
		t.Fatalf("cancelled bootstrap: error=%v calls=%d", err, fixture.Get("calls").Int())
	}
}
