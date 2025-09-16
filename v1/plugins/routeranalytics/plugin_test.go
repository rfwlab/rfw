package routeranalytics

import "testing"

func TestTransitionProbabilities(t *testing.T) {
	p := New(Options{})
	p.Reset()
	p.handleNavigation("/home")
	p.handleNavigation("/posts")
	p.handleNavigation("/home")
	p.handleNavigation("/posts")
	p.handleNavigation("/settings")
	p.handleNavigation("/posts")
	p.handleNavigation("/home")

	probs := p.TransitionProbabilities("/posts")
	if len(probs) != 2 {
		t.Fatalf("expected two transitions, got %d", len(probs))
	}
	if probs[0].To != "/home" {
		t.Fatalf("expected /home to be most probable, got %s", probs[0].To)
	}
	if probs[0].Count != 2 || probs[1].Count != 1 {
		t.Fatalf("expected counts to reflect visits, got %v", probs)
	}

	if got := p.TransitionProbabilities("/missing"); got != nil {
		t.Fatalf("expected nil for unknown route, got %v", got)
	}
}

func TestMostLikelyNextOrdering(t *testing.T) {
	p := New(Options{})
	p.Reset()
	p.handleNavigation("/a")
	p.handleNavigation("/b")
	p.handleNavigation("/c")
	p.handleNavigation("/b")
	p.handleNavigation("/c")
	p.handleNavigation("/d")
	p.handleNavigation("/b")
	p.handleNavigation("/d")

	probs := p.MostLikelyNext("/b", 2)
	if len(probs) != 2 {
		t.Fatalf("expected two results, got %d", len(probs))
	}
	if probs[0].To != "/c" {
		t.Fatalf("expected /c to be first, got %s", probs[0].To)
	}
	if probs[0].Probability <= probs[1].Probability {
		t.Fatalf("expected probability ordering")
	}

	if limited := p.MostLikelyNext("/b", 1); len(limited) != 1 || limited[0].To != "/c" {
		t.Fatalf("expected limit to return most probable route")
	}

	if none := p.MostLikelyNext("/z", 3); none != nil {
		t.Fatalf("expected nil for unknown transitions, got %v", none)
	}
}

func TestDefaultNormalize(t *testing.T) {
	p := New(Options{})
	normalized := p.opts.Normalize(" docs/view?id=7#section ")
	if normalized != "/docs/view" {
		t.Fatalf("expected normalized path, got %q", normalized)
	}

	if global := NormalizePath(" docs/view?id=7#section "); global != normalized {
		t.Fatalf("expected NormalizePath to match default normalize, got %q", global)
	}
}
