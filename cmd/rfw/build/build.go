package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/assets"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/bundler"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/copy"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/docs"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/env"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/pages"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/seo"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/tailwind"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/test"
	"github.com/rfwlab/rfw/cmd/rfw/utils"
)

type buildOptions struct {
	DevBuild     bool
	SkipOptimize bool
}

func goBuildArgs(opts buildOptions) []string {
	args := []string{"build"}
	var tags []string
	if opts.DevBuild {
		tags = append(tags, "rfwdev")
	}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	if !opts.SkipOptimize {
		args = append(args, "-trimpath", "-ldflags=-s -w")
	}
	return args
}

func Build() error {
	var manifest struct {
		Build struct {
			Type string `json:"type"`
		} `json:"build"`
		Plugins map[string]json.RawMessage `json:"plugins"`
	}
	if data, err := os.ReadFile("rfw.json"); err == nil {
		_ = json.Unmarshal(data, &manifest)
	}
	if err := plugins.Configure(manifest.Plugins); err != nil {
		return fmt.Errorf("failed to configure plugins: %w", err)
	}
	if err := plugins.PreBuild(); err != nil {
		return fmt.Errorf("pre build failed: %w", err)
	}

	clientDir := filepath.Join("build", "client")
	hostDir := filepath.Join("build", "host")
	staticDir := filepath.Join("build", "static")
	if err := os.MkdirAll(clientDir, 0o755); err != nil {
		return fmt.Errorf("failed to create client build directory: %w", err)
	}
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		return fmt.Errorf("failed to create static build directory: %w", err)
	}

	wasmExec, err := findWasmExec()
	if err != nil {
		return err
	}
	if err := copyFile(wasmExec, filepath.Join(clientDir, "wasm_exec.js")); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %w", err)
	}

	args := goBuildArgs(buildOptions{
		DevBuild:     os.Getenv("RFW_DEV_BUILD") == "1",
		SkipOptimize: os.Getenv("RFW_DEV_BUILD") == "1" || utils.IsDebug() || os.Getenv("RFW_SKIP_STRIP") == "1",
	})
	wasmPath := filepath.Join(clientDir, "app.wasm")
	args = append(args, "-o", wasmPath, ".")
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOARCH=wasm", "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build project: %s: %w", output, err)
	}

	isDev := utils.IsDebug() || os.Getenv("RFW_DEV_BUILD") == "1"
	if !isDev {
		if err := compressWasmBrotli(wasmPath); err != nil {
			return fmt.Errorf("failed to brotli-compress wasm: %w", err)
		}
	}

	// Always build host binary if host directory exists (SSC mode).
	if _, err := os.Stat("host"); err == nil {
		if err := os.MkdirAll(hostDir, 0o755); err != nil {
			return fmt.Errorf("failed to create host build directory: %w", err)
		}
		hostArgs := []string{"build", "-o", filepath.Join(hostDir, "host"), "./host"}
		if isDev {
			hostArgs = []string{"build", "-o", filepath.Join(hostDir, "host"), "./host"}
		}
		hostCmd := exec.Command("go", hostArgs...)
		if hostOutput, err := hostCmd.CombinedOutput(); err != nil {
			if !isDev {
				return fmt.Errorf("failed to build host components: %s: %w", hostOutput, err)
			}
			fmt.Fprintf(os.Stderr, "warning: host build failed (dev mode, continuing): %s\n", hostOutput)
		}
	}
	if err := plugins.Build(); err != nil {
		return fmt.Errorf("failed to run plugins: %w", err)
	}

	// Copy plugin-generated assets (e.g. tailwind.css) to client build dir.
	for _, name := range []string{"tailwind.css", "input.css"} {
		if data, err := os.ReadFile(name); err == nil {
			if err := os.WriteFile(filepath.Join(clientDir, name), data, 0o644); err != nil {
				return fmt.Errorf("failed to copy %s to client dir: %w", name, err)
			}
		}
	}
	if _, err := os.Stat("index.html"); err == nil {
		if err := copyFile("index.html", filepath.Join(clientDir, "index.html")); err != nil {
			return fmt.Errorf("failed to copy index.html: %w", err)
		}
	}

	if _, err := os.Stat("wasm_loader.js"); err == nil {
		if err := copyFile("wasm_loader.js", filepath.Join(clientDir, "wasm_loader.js")); err != nil {
			return fmt.Errorf("failed to copy wasm_loader.js: %w", err)
		}
	}

	if _, err := os.Stat("static"); err == nil {
		if err := filepath.Walk("static", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel("static", path)
			if err != nil {
				return err
			}
			dst := filepath.Join(staticDir, rel)
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			return copyFile(path, dst)
		}); err != nil {
			return fmt.Errorf("failed to copy static assets: %w", err)
		}
	}

	if err := plugins.PostBuild(); err != nil {
		return fmt.Errorf("post build failed: %w", err)
	}

	return nil
}

// findWasmExec locates wasm_exec.js from the active Go toolchain.
// It tries the canonical Go 1.21+ path ($GOROOT/lib/wasm/), then the
// legacy path ($GOROOT/misc/wasm/), and finally a project-local copy.
func findWasmExec() (string, error) {
	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOROOT: %w", err)
	}
	root := strings.TrimSpace(string(goroot))
	candidates := []string{
		filepath.Join(root, "lib", "wasm", "wasm_exec.js"),
		filepath.Join(root, "misc", "wasm", "wasm_exec.js"),
		"wasm_exec.js",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf(
		"wasm_exec.js not found in GOROOT (%s) or project root; reinstall Go or run 'rfw init'",
		root,
	)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func compressWasmBrotli(src string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dst := src + ".br"
	tmp, err := os.CreateTemp(filepath.Dir(dst), filepath.Base(dst)+".*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if tmp != nil {
			tmp.Close()
		}
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	writer := brotli.NewWriterLevel(tmp, brotli.BestCompression)
	if _, err := io.Copy(writer, in); err != nil {
		writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	tmp = nil
	if err := os.Rename(tmpName, dst); err != nil {
		return err
	}
	return nil
}
