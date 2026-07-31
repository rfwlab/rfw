//go:build !js

package tailwind

import (
	"os"
	"path/filepath"
	"testing"
)

// TestShouldRebuild ensures the plugin's rebuild triggers are detected
// correctly based on file paths and extensions.
func TestShouldRebuild(t *testing.T) {
	p := &plugin{output: "tailwind.css"}

	if !p.ShouldRebuild("style.css") {
		t.Fatalf("expected css change to trigger rebuild")
	}
	if p.ShouldRebuild("tailwind.css") {
		t.Fatalf("output file should not trigger rebuild")
	}
	if !p.ShouldRebuild("index.html") || !p.ShouldRebuild("tmpl.rtml") {
		t.Fatalf("html/rtml should trigger rebuild")
	}
	if !p.ShouldRebuild("main.go") {
		t.Fatalf("go files should trigger rebuild")
	}
	if p.ShouldRebuild("image.png") {
		t.Fatalf("unrelated files should not trigger rebuild")
	}
}

func TestValidateProjectFile(t *testing.T) {
	root := t.TempDir()
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
	if err := os.WriteFile("input.css", []byte("body {}"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := validateProjectFile("input.css", true); err != nil {
		t.Fatalf("expected project file to be valid: %v", err)
	}
	if _, err := validateProjectFile("dist/css/app.css", false); err != nil {
		t.Fatalf("expected nested output to be valid: %v", err)
	}
	for _, filePath := range []string{"../input.css", "/tmp/input.css", "-o"} {
		if _, err := validateProjectFile(filePath, false); err == nil {
			t.Errorf("expected %q to be rejected", filePath)
		}
	}

	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Fatal(err)
	}
	if _, err := validateProjectFile("outside/output.css", false); err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}

func TestValidateExtraArgs(t *testing.T) {
	root := t.TempDir()
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
	if err := os.WriteFile("tailwind.config.js", nil, 0o600); err != nil {
		t.Fatal(err)
	}

	useConfig, err := validateExtraArgs([]string{"--config", "tailwind.config.js"})
	if err != nil {
		t.Fatalf("expected config argument to be valid: %v", err)
	}
	if !useConfig {
		t.Fatal("expected config to be enabled")
	}
	for _, args := range [][]string{{"--watch"}, {"--config"}, {"--config", "../config.js"}, {"--config", "config/tailwind.config.js"}} {
		if _, err := validateExtraArgs(args); err == nil {
			t.Errorf("expected %v to be rejected", args)
		}
	}
}

func TestTailwindCommandUsesFixedArguments(t *testing.T) {
	tests := []struct {
		name      string
		minify    bool
		useConfig bool
		want      []string
	}{
		{name: "default", want: []string{"tailwindcss", "-i", "-"}},
		{name: "minified", minify: true, want: []string{"tailwindcss", "-i", "-", "--minify"}},
		{name: "configured", useConfig: true, want: []string{"tailwindcss", "-i", "-", "--config", "tailwind.config.js"}},
		{
			name:      "configured and minified",
			minify:    true,
			useConfig: true,
			want:      []string{"tailwindcss", "-i", "-", "--minify", "--config", "tailwind.config.js"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := tailwindCommand(test.minify, test.useConfig)
			if len(command.Args) != len(test.want) {
				t.Fatalf("args %v, want %v", command.Args, test.want)
			}
			for index := range test.want {
				if command.Args[index] != test.want[index] {
					t.Fatalf("args %v, want %v", command.Args, test.want)
				}
			}
		})
	}
}
