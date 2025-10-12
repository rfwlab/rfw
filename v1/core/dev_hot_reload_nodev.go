//go:build js && wasm && !rfwdev

package core

func startDevTemplateWatcher() {}

func devApplyTemplateUpdate(string, string) {}

func devOverrideTemplate(c *HTMLComponent, template string) string { return template }

func devRegisterComponent(*HTMLComponent) {}

func devUnregisterComponent(*HTMLComponent) {}
