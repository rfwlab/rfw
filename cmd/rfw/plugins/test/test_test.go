//go:build !js

package test

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestValidatePackagePattern(t *testing.T) {
	for _, pattern := range []string{".", "./...", "./pkg/...", "./foo-bar", "example.com/team/module/pkg"} {
		if err := validatePackagePattern(pattern); err != nil {
			t.Errorf("expected %q to be valid: %v", pattern, err)
		}
	}
	for _, pattern := range []string{"", "-exec=sh", "../...", "/tmp/pkg", "./pkg name", "./pkg=./other", "./pkg/.../nested"} {
		if err := validatePackagePattern(pattern); err == nil {
			t.Errorf("expected %q to be rejected", pattern)
		}
	}
}

func TestConfiguredPackagesDefaultsToCurrentPackage(t *testing.T) {
	for _, packages := range [][]string{nil, {}} {
		configured := configuredPackages(packages)
		if len(configured) != 1 || configured[0] != "." {
			t.Fatalf("configured packages %v, want current package", configured)
		}
	}

	configured := configuredPackages([]string{"./..."})
	if len(configured) != 1 || configured[0] != "./..." {
		t.Fatalf("configured packages %v, want recursive project pattern", configured)
	}
}

func TestPackageTestCommand(t *testing.T) {
	root := t.TempDir()
	packageDirectory := filepath.Join(root, "pkg")
	if err := os.Mkdir(packageDirectory, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/team/module\n\ngo 1.25\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	cmd, err := packageTestCommand("./pkg/...")
	if err != nil {
		t.Fatalf("create package test command: %v", err)
	}
	if cmd.Path == "" || len(cmd.Args) != 3 || cmd.Args[1] != "test" || cmd.Args[2] != "./..." {
		t.Fatalf("unexpected command: path=%q args=%v", cmd.Path, cmd.Args)
	}
	resolvedPackageDirectory, err := filepath.EvalSymlinks(packageDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != resolvedPackageDirectory {
		t.Fatalf("command directory %q, want %q", cmd.Dir, resolvedPackageDirectory)
	}

	moduleCommand, err := packageTestCommand("example.com/team/module/pkg")
	if err != nil {
		t.Fatalf("create module package test command: %v", err)
	}
	if moduleCommand.Dir != resolvedPackageDirectory || moduleCommand.Args[2] != "." {
		t.Fatalf("unexpected module command: dir=%q args=%v", moduleCommand.Dir, moduleCommand.Args)
	}
	if _, err := packageTestCommand("example.com/other/module/pkg"); err == nil {
		t.Fatal("expected external module package to be rejected")
	}

	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Fatal(err)
	}
	if _, err := packageTestCommand("./outside"); err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}
