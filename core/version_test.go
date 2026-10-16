package core

import (
	"runtime/debug"
	"testing"
)

func TestVersionUsesFrameworkDependencyInsteadOfApplicationVersion(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Path: "example.com/application", Version: "v9.0.0"},
		Deps: []*debug.Module{{Path: "github.com/rfwlab/rfw/v2", Version: "v2.3.1-0.20261016153501-8ce9ba2d776b"}},
	}
	if got := frameworkVersion(info); got != info.Deps[0].Version {
		t.Fatalf("framework version: %q", got)
	}
}

func TestVersionUsesInstalledCLIModule(t *testing.T) {
	info := &debug.BuildInfo{Main: debug.Module{Path: "github.com/rfwlab/rfw/v2", Version: "v2.3.0"}}
	if got := frameworkVersion(info); got != "v2.3.0" {
		t.Fatalf("CLI version: %q", got)
	}
}

func TestVersionDoesNotStampLocalReplacementAsRelease(t *testing.T) {
	info := &debug.BuildInfo{Deps: []*debug.Module{{
		Path: "github.com/rfwlab/rfw/v2", Version: "v2.3.0",
		Replace: &debug.Module{Path: "../rfw"},
	}}}
	if got := frameworkVersion(info); got != "(devel)" {
		t.Fatalf("local replacement version: %q", got)
	}
}

func TestVersionUsesVersionedFrameworkReplacement(t *testing.T) {
	info := &debug.BuildInfo{Deps: []*debug.Module{{
		Path: "github.com/rfwlab/rfw/v2", Version: "v2.3.0",
		Replace: &debug.Module{Path: "github.com/rfwlab/rfw/v2", Version: "v2.3.1"},
	}}}
	if got := frameworkVersion(info); got != "v2.3.1" {
		t.Fatalf("versioned replacement: %q", got)
	}
}

func TestVersionUsesReleaseLinkerFallback(t *testing.T) {
	old := version
	t.Cleanup(func() { version = old })
	version = "v2.4.0"
	if got := frameworkVersion(&debug.BuildInfo{Main: debug.Module{Path: "github.com/rfwlab/rfw/v2", Version: "(devel)"}}); got != version {
		t.Fatalf("release fallback: %q", got)
	}
}
