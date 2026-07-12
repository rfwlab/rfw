package core

import "runtime/debug"

// version is the fallback when no build metadata is available (e.g. a plain
// `go build` inside the repo). Release builds override it with ldflags, and
// `go install module@version` builds report the module version from build
// info, so users installed via the proxy see the real tag.
var version = "v2.0.0-beta.8"

// Version returns the framework version for this build.
func Version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		if v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}
