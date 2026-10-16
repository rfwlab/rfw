package core

import "runtime/debug"

// Release builds override version with ldflags.
var version = "(devel)"

// Version returns the framework version for this build.
func Version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return frameworkVersion(info)
	}
	return version
}

func frameworkVersion(info *debug.BuildInfo) string {
	const modulePath = "github.com/rfwlab/rfw/v2"
	if info.Main.Path == modulePath {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
		return version
	}
	for _, dep := range info.Deps {
		if dep.Path != modulePath {
			continue
		}
		if dep.Replace != nil {
			if dep.Replace.Path != modulePath || dep.Replace.Version == "" {
				return "(devel)"
			}
			dep = dep.Replace
		}
		if dep.Version != "" && dep.Version != "(devel)" {
			return dep.Version
		}
	}
	return version
}
