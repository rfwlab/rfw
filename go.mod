module github.com/rfwlab/rfw

go 1.25.0

require (
	github.com/andybalholm/brotli v1.0.6
	github.com/bmatcuk/doublestar/v4 v4.9.1
	github.com/fatih/color v1.17.0
	github.com/fsnotify/fsnotify v1.9.0
	github.com/mirkobrombin/go-cli-builder v1.0.0
	github.com/mirkobrombin/go-logger v0.2.0
	github.com/mirkobrombin/go-signal/v2 v2.0.0
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/rfwlab/rfw/v2 v2.0.0-alpha.1
	github.com/tdewolff/minify/v2 v2.24.3
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mirkobrombin/go-foundation v1.1.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/tdewolff/parse/v2 v2.8.3 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	nhooyr.io/websocket v1.8.10 // indirect
)

replace github.com/rfwlab/rfw/v2 v2.0.0-alpha.1 => ./v2
