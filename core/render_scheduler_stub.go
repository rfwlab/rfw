//go:build !js || !wasm

package core

func (c *HTMLComponent) requestRender() {}

func cancelComponentRender(string) {}
