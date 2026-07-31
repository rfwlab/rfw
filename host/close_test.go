package host

import (
	"io"
	"testing"
)

func closeTestResource(t *testing.T, resource io.Closer) {
	t.Helper()
	if err := resource.Close(); err != nil {
		t.Errorf("close resource: %v", err)
	}
}
