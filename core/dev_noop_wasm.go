//go:build js && wasm

package core

func devOverrideTemplate(c *HTMLComponent, template string) string { return template }
func devRegisterComponent(*HTMLComponent) {}
func devUnregisterComponent(*HTMLComponent) {}