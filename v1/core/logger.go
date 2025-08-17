package core

import (
	"log"

	"github.com/rfwlab/rfw/v1/state"
)

// Logger defines logging interface used by the framework.
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
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

func (defaultLogger) Debug(format string, v ...interface{}) { log.Printf("DEBUG: "+format, v...) }
func (defaultLogger) Info(format string, v ...interface{})  { log.Printf("INFO: "+format, v...) }
func (defaultLogger) Warn(format string, v ...interface{})  { log.Printf("WARN: "+format, v...) }
func (defaultLogger) Error(format string, v ...interface{}) { log.Printf("ERROR: "+format, v...) }
