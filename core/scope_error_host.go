//go:build !js || !wasm

package core

import "log"

func reportScopeError(err any) {
	log.Printf("component scope cleanup: %v", err)
}
