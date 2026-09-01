package http

import (
	"sync"
	"time"
)

var (
	httpHookMu sync.RWMutex
	httpHook   func(start bool, url string, status int, duration time.Duration)
)

// RegisterHTTPHook registers fn to receive HTTP request events.
func RegisterHTTPHook(fn func(start bool, url string, status int, duration time.Duration)) {
	httpHookMu.Lock()
	httpHook = fn
	httpHookMu.Unlock()
}

func emitHTTPHook(start bool, url string, status int, duration time.Duration) {
	httpHookMu.RLock()
	hook := httpHook
	httpHookMu.RUnlock()
	if hook != nil {
		hook(start, url, status, duration)
	}
}
