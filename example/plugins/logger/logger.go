//go:build js && wasm

// Package logger provides a plugin that redirects framework logs to the
// browser console.
package logger

import (
	"encoding/json"
	"fmt"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
)

// Plugin installs a custom logger that forwards framework logs to the
// JavaScript console. It replaces the default logger used by rfw.
type Plugin struct{}

// New creates a new logger plugin instance.
func New() core.Plugin { return &Plugin{} }

// Install sets the core logger implementation to use the browser console.
func (p *Plugin) Install(a *core.App) {
	core.SetLogger(consoleLogger{console: js.Console()})
	core.Log().Info("rfw console logger active")
}

func (p *Plugin) Build(json.RawMessage) error { return nil }

type consoleLogger struct{ console jst.Value }

func (cl consoleLogger) Debug(format string, v ...any) {
	cl.console.Call("debug", fmt.Sprintf(format, v...))
}

func (cl consoleLogger) Info(format string, v ...any) {
	cl.console.Call("info", fmt.Sprintf(format, v...))
}

func (cl consoleLogger) Warn(format string, v ...any) {
	cl.console.Call("warn", fmt.Sprintf(format, v...))
}

func (cl consoleLogger) Error(format string, v ...any) {
	cl.console.Call("error", fmt.Sprintf(format, v...))
}
