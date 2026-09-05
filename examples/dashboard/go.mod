module github.com/rfwlab/examples/dashboard

go 1.26.0

require github.com/rfwlab/rfw/v2 v2.0.0-beta.19

require (
	github.com/mirkobrombin/go-foundation/v2 v2.1.0 // indirect
	github.com/tdewolff/minify/v2 v2.24.3 // indirect
	github.com/tdewolff/parse/v2 v2.8.3 // indirect
	golang.org/x/net v0.57.0 // indirect
	nhooyr.io/websocket v1.8.10 // indirect
)

// Always build against the checkout this example ships with.
replace github.com/rfwlab/rfw/v2 => ../..
