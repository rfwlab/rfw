package logging

import (
	"fmt"

	"github.com/mirkobrombin/go-logger/pkg/logger"
)

var Log logger.Logger

func init() {
	Log = logger.New()
}

func F(key string, value any) logger.Field {
	return logger.Field{Key: key, Value: value}
}

type StoreLogAdapter struct{}

func (StoreLogAdapter) Printf(format string, args ...any) {
	Log.Debug(fmt.Sprintf(format, args...))
}
