package dom

import (
	"strings"
	"testing"
)

func TestStyleInline(t *testing.T) {
	got := StyleInline(map[string]string{"color": "red", "display": "block"})
	if !strings.Contains(got, "color:red") || !strings.Contains(got, "display:block") {
		t.Fatalf("StyleInline() = %q", got)
	}
}
