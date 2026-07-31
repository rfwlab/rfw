//go:build js && wasm

package core

func devOverrideTemplate(_ *HTMLComponent, template string) string { return template }
func devRegisterComponent(*HTMLComponent)                          {}
func devUnregisterComponent(*HTMLComponent)                        {}
