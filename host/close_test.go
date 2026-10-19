package host

import (
	"errors"
	"io"
	"net"
	"syscall"
	"testing"
)

func closeTestResource(t *testing.T, resource io.Closer) {
	t.Helper()
	if err := resource.Close(); err != nil && !errors.Is(err, net.ErrClosed) &&
		!errors.Is(err, syscall.EPIPE) && !errors.Is(err, syscall.ECONNRESET) {
		t.Errorf("close resource: %v", err)
	}
}
