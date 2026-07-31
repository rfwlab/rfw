//go:build !js

// Package test runs Go tests as a build plugin.
package test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/rfwlab/rfw/v2/cmd/rfw/logging"
	"github.com/rfwlab/rfw/v2/cmd/rfw/plugins"
)

type plugin struct{}

func init() { plugins.Register(&plugin{}) }

func (p *plugin) Name() string { return "test" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) error {
	cfg := struct {
		Packages []string `json:"packages"`
	}{Packages: []string{"./..."}}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("decode test plugin config: %w", err)
		}
	}
	for _, pattern := range configuredPackages(cfg.Packages) {
		cmd, err := packageTestCommand(pattern)
		if err != nil {
			return err
		}
		output, runErr := cmd.CombinedOutput()
		out := strings.TrimSpace(string(output))
		if runErr != nil {
			logging.Log.Error("go test failed", logging.F("plugin", "test"), logging.F("package", pattern), logging.F("output", out), logging.F("error", runErr.Error()))
			return runErr
		}
		logging.Log.Info("go test ok", logging.F("plugin", "test"), logging.F("package", pattern), logging.F("output", out))
	}
	return nil
}

func configuredPackages(packages []string) []string {
	if len(packages) == 0 {
		return []string{"."}
	}
	return packages
}

func validatePackagePattern(pkg string) error {
	if pkg == "." || pkg == "./..." {
		return nil
	}
	if pkg == "" || strings.HasPrefix(pkg, "-") || strings.HasPrefix(pkg, "/") {
		return fmt.Errorf("test package %q is not a package pattern", pkg)
	}
	for _, char := range pkg {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return fmt.Errorf("test package %q contains whitespace or control characters", pkg)
		}
		if char >= 'a' && char <= 'z' ||
			char >= 'A' && char <= 'Z' ||
			char >= '0' && char <= '9' ||
			strings.ContainsRune("-_./+", char) {
			continue
		}
		return fmt.Errorf("test package %q contains invalid character %q", pkg, char)
	}
	segments := strings.Split(strings.TrimPrefix(pkg, "./"), "/")
	for index, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("test package %q escapes the project", pkg)
		}
		if segment == "..." && index != len(segments)-1 {
			return fmt.Errorf("test package %q has a non-terminal recursive wildcard", pkg)
		}
	}
	return nil
}

func packageTestCommand(pattern string) (*exec.Cmd, error) {
	if err := validatePackagePattern(pattern); err != nil {
		return nil, err
	}
	localPattern := pattern
	if pattern != "." && !strings.HasPrefix(pattern, "./") {
		moduleOutput, err := exec.Command("go", "list", "-m").Output()
		if err != nil {
			return nil, fmt.Errorf("resolve project module: %w", err)
		}
		modulePath := strings.TrimSpace(string(moduleOutput))
		switch {
		case pattern == modulePath:
			localPattern = "."
		case strings.HasPrefix(pattern, modulePath+"/"):
			localPattern = "./" + strings.TrimPrefix(pattern, modulePath+"/")
		default:
			return nil, fmt.Errorf("test package %q is outside module %q", pattern, modulePath)
		}
	}
	switch localPattern {
	case ".":
		return exec.Command("go", "test", "."), nil
	case "./...":
		return exec.Command("go", "test", "./..."), nil
	}

	relative := strings.TrimPrefix(localPattern, "./")
	recursive := strings.HasSuffix(relative, "/...")
	relative = strings.TrimSuffix(relative, "/...")
	directory, err := resolveProjectDirectory(relative)
	if err != nil {
		return nil, fmt.Errorf("test package %q: %w", pattern, err)
	}

	var cmd *exec.Cmd
	if recursive {
		cmd = exec.Command("go", "test", "./...")
	} else {
		cmd = exec.Command("go", "test", ".")
	}
	cmd.Dir = directory
	return cmd, nil
}

func resolveProjectDirectory(relative string) (string, error) {
	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return "", err
	}
	resolvedDirectory, err := filepath.EvalSymlinks(filepath.Join(resolvedRoot, filepath.Clean(relative)))
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolvedDirectory)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", relative)
	}
	pathWithinRoot, err := filepath.Rel(resolvedRoot, resolvedDirectory)
	if err != nil || pathWithinRoot == ".." || strings.HasPrefix(pathWithinRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q resolves outside the project", relative)
	}
	return resolvedDirectory, nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}
