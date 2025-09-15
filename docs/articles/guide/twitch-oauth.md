# Twitch OAuth

This guide shows how to integrate Twitch’s **Authorization Code Grant** flow using only **rfw** APIs. The flow consists of:

1. A **login component** that redirects users to Twitch.
2. A **callback component** that receives the authorization code.
3. A **host component** that exchanges the code for an access token.

---

## Login Component

The login page renders a button linking to Twitch’s authorization endpoint:

@include\:ExampleFrame:{code:"/examples/components/twitch\_login\_component.go", uri:"/examples/twitch/login"}

---

## Callback Component

After Twitch redirects back, the callback component:

* Extracts the `code` query parameter.
* Sends it to the host for exchange.

@include\:ExampleFrame:{code:"/examples/components/twitch\_callback\_component.go", uri:"/examples/twitch/callback"}

---

## Host Component

The host performs the secure exchange:

* Uses the official [helix](https://github.com/nicklaw5/helix) Go client.
* Requests an access token from Twitch.
* Pushes the result back to the Wasm component.

The host component is located at:

```
docs/host/components/twitch_oauth_host.go
```

It requires the following environment variables:

* `TWITCH_CLIENT_ID`
* `TWITCH_CLIENT_SECRET`

---

With this setup, you can build Twitch-authenticated features entirely in Go, combining client and server logic within rfw.
