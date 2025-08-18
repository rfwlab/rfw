package tailwind

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
)

type plugin struct{}

func init() {
	plugins.Register(&plugin{})
}

func (p *plugin) Name() string { return "tailwind" }

func (p *plugin) Build() error {
	bin, err := exec.LookPath("tailwindcss")
	if err != nil {
		if err := downloadTailwindCLI(); err != nil {
			return err
		}
		bin = "./tailwindcss"
	}

	args := []string{
		"-i", "input.css",
		"-o", "tailwind.css",
		"--minify",
	}
	if _, err := os.Stat("tailwind.config.js"); err == nil {
		args = append(args, "-c", "tailwind.config.js")
	} else {
		args = append(args,
			"--content", "./**/*.rtml",
			"--content", "./**/*.go",
			"--content", "./*.html",
		)
	}

	cmd := exec.Command(bin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailwind build failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	if strings.HasSuffix(path, "tailwind.config.js") {
		return true
	}
	if strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, "tailwind.css") {
		return true
	}
	return false
}

func downloadTailwindCLI() error {
	var url string
	osys := runtime.GOOS
	arch := runtime.GOARCH
	switch osys {
	case "linux":
		if arch == "amd64" {
			url = "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64"
		} else if arch == "arm64" {
			url = "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-arm64"
		}
	case "darwin":
		if arch == "amd64" {
			url = "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64"
		} else if arch == "arm64" {
			url = "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64"
		}
	}
	if url == "" {
		return fmt.Errorf("unsupported platform %s/%s for tailwindcss", osys, arch)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download tailwindcss: %w", err)
	}
	defer resp.Body.Close()

	f, err := os.Create("tailwindcss")
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	if err := f.Chmod(0o755); err != nil {
		return err
	}
	return nil
}
