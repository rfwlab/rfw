//go:build !js

// Package logging provides shared CLI logging helpers.
package logging

import (
	"fmt"

	"github.com/mirkobrombin/go-logger/pkg/logger"
)

// Log is the shared CLI logger.
var Log logger.Logger

func init() {
	Log = logger.New()
}

// F builds a structured log field.
func F(key string, value any) logger.Field {
	return logger.Field{Key: key, Value: value}
}

// StoreLogAdapter adapts the CLI logger to state.Logger.
type StoreLogAdapter struct{}

// Printf writes a formatted debug message.
func (StoreLogAdapter) Printf(format string, args ...any) {
	Log.Debug(fmt.Sprintf(format, args...))
}
