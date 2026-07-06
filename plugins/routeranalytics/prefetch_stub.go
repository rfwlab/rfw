//go:build !js || !wasm

package routeranalytics

func newPrefetcher(string) prefetcher { return noopPrefetcher{} }
