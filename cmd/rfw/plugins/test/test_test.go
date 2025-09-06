package test

import "testing"

// TestShouldRebuild verifies that the test plugin triggers rebuilds for Go test
// files only.
func TestShouldRebuild(t *testing.T) {
	p := &plugin{}
	if !p.ShouldRebuild("foo_test.go") {
		t.Fatalf("expected rebuild for _test.go files")
	}
	if p.ShouldRebuild("main.go") {
		t.Fatalf("non-test files should not trigger rebuild")
	}
}
