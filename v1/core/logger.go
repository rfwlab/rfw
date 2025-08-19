package core

import (
	"log"

	"github.com/rfwlab/rfw/v1/state"
)

// Logger defines logging interface used by the framework.
type Logger interface {
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
}

// logger is the current logger implementation.
var logger Logger = defaultLogger{}

func init() { state.SetLogger(logger) }

// SetLogger allows applications to replace the default logger.
func SetLogger(l Logger) {
	if l != nil {
		logger = l
		state.SetLogger(l)
	}
}

// Log returns the active logger implementation.
func Log() Logger { return logger }

// defaultLogger is the fallback logger using the standard log package.
type defaultLogger struct{}

func (defaultLogger) Debug(format string, v ...any) { log.Printf("DEBUG: "+format, v...) }
func (defaultLogger) Info(format string, v ...any)  { log.Printf("INFO: "+format, v...) }
func (defaultLogger) Warn(format string, v ...any)  { log.Printf("WARN: "+format, v...) }
func (defaultLogger) Error(format string, v ...any) { log.Printf("ERROR: "+format, v...) }
