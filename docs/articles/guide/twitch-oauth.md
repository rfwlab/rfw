# Twitch OAuth

This example demonstrates how to integrate Twitch's Authorization Code Grant flow using only **rfw** APIs. It consists of a login component that redirects the user to Twitch and a callback component connected to a host component that exchanges the authorization code for an access token.

The login page renders a button linking to the Twitch authorization endpoint:

@include:ExampleFrame:{code:"/examples/components/twitch_login_component.go", uri:"/examples/twitch/login"}

After Twitch redirects back to the application, the callback component extracts the `code` query parameter and sends it to the host. The host uses the official [helix](https://github.com/nicklaw5/helix) Go client to request an access token and pushes the result back to the Wasm component:

@include:ExampleFrame:{code:"/examples/components/twitch_callback_component.go", uri:"/examples/twitch/callback"}

The host component can be found in `docs/host/main.go` and expects the environment variables `TWITCH_CLIENT_ID` and `TWITCH_CLIENT_SECRET` to be set.
