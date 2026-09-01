# RFW Capacitor SSC transport (local preview)

This unpublished package supplies the iOS-native SSC transport selected by
`build.sscTransport: "capacitor"`. It is intentionally local while the physical
device and security gates are being completed.

Install it from this checkout into the Capacitor wrapper, then synchronize the
native project:

```bash
npm install --save /absolute/path/to/rfw/mobile/capacitor-ssc
npx cap sync ios
```

Run the iOS package tests with a full Xcode installation and at least one
available simulator:

```bash
npm run verify:ios
```

Set `RFW_IOS_SIMULATOR_ID` to select a specific simulator; otherwise the script
uses the first available iOS simulator reported by Xcode.

Configure an exact native endpoint allowlist. The plugin fails closed when the
allowlist is absent, when the endpoint is not WSS, or when the requested origin
does not match exactly:

```json
{
  "plugins": {
    "RFWSSC": {
      "allowedOrigins": ["wss://api.example.com"],
      "origin": "capacitor://localhost",
      "outboundQueueSize": 256,
      "maxMessageBytes": 8388608,
      "connectTimeoutMilliseconds": 10000
    }
  }
}
```

The plugin uses `HTTPCookieStorage.shared`, the same native cookie jar used by
Capacitor HTTP on iOS. Cookie values are never returned over the JavaScript
bridge. Only SSC frames and connection lifecycle events cross the bridge.

Changing the authenticated user or exact host auth-session ID is a connection
ownership boundary. After the authoritative login response has installed the
new cookie, unmount the previous user's protected owners and call
`hostclient.ResetSession()` before mounting the replacement identity. The reset
closes this native socket and discards the old resume/outbox state; allowing the
old connection to survive would keep receiving data for its original user until
the next background or network reconnect.

This preview contains no Android implementation yet. The Go/WASM selector and
bridge event contract are platform-neutral; Android remains a separate native
implementation and physical-device gate.
