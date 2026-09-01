//go:build js && wasm

package docs

import (
	"testing"

	"github.com/rfwlab/rfw/v2/state"
)

func TestPluginProviderAndOptional(t *testing.T) {
	p := New("/sidebar.json")
	provided := p.Provide()

	if _, ok := provided["sidebar"].(*state.Signal[[]SidebarItem]); !ok {
		t.Fatalf("expected sidebar signal provider, got %T", provided["sidebar"])
	}
	if _, ok := provided["article"].(*state.Signal[*ArticleData]); !ok {
		t.Fatalf("expected article signal provider, got %T", provided["article"])
	}
	if _, ok := provided["loadDoc"].(func(string)); !ok {
		t.Fatalf("expected loadDoc function provider, got %T", provided["loadDoc"])
	}
	if len(p.Optional()) != 1 {
		t.Fatalf("expected SEO optional plugin by default")
	}

	withoutSEO := New("/sidebar.json", true)
	if optional := withoutSEO.Optional(); optional != nil {
		t.Fatalf("expected nil optional plugins when SEO disabled, got %v", optional)
	}
}

func TestHeadingsToAny(t *testing.T) {
	out := headingsToAny([]Heading{{Text: "Intro", Depth: 2, ID: "intro"}})
	if len(out) != 1 {
		t.Fatalf("expected one heading, got %d", len(out))
	}
	m, ok := out[0].(map[string]any)
	if !ok {
		t.Fatalf("expected map heading, got %T", out[0])
	}
	if m["text"] != "Intro" || m["depth"] != 2 || m["id"] != "intro" {
		t.Fatalf("unexpected heading map: %v", m)
	}
}
