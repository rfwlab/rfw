//go:build !js || !wasm

package core

type HTMLComponent struct{}

func startDevTemplateWatcher() {}

func devApplyTemplateUpdate(string, string) {}

func devOverrideTemplate(c *HTMLComponent, template string) string { return template }

func devRegisterComponent(*HTMLComponent) {}

func devUnregisterComponent(*HTMLComponent) {}
