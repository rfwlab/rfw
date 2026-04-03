package logging

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-logger/pkg/logger"
)

var Log logger.Logger

func init() {
	sink := logger.NewConsoleSink(os.Stdout)
	Log = logger.New(logger.WithSink(sink))
}

func F(key string, value any) logger.Field {
	return logger.Field{Key: key, Value: value}
}

type StoreLogAdapter struct{}

func (StoreLogAdapter) Printf(format string, args ...any) {
	Log.Debug(fmt.Sprintf(format, args...))
}
