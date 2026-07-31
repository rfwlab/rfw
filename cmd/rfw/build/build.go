//go:build !js

// Package build compiles RFW applications and runs build plugins.
package build

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/rfwlab/rfw/v2/cmd/rfw/plugins"
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/assets"   // Register the assets build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/bundler"  // Register the bundler build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/copy"     // Register the copy build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/docs"     // Register the docs build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/env"      // Register the environment build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/pages"    // Register the pages build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/seo"      // Register the SEO build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/tailwind" // Register the Tailwind build plugin.
	_ "github.com/rfwlab/rfw/v2/cmd/rfw/plugins/test"     // Register the test build plugin.
	"github.com/rfwlab/rfw/v2/cmd/rfw/utils"
)

// Build compiles the configured application and runs its build plugins.
func Build() error {
	var manifest struct {
		Build struct {
			Type string `json:"type"`
			Host string `json:"host"`
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
	if err := makePublicDir(clientDir); err != nil {
		return fmt.Errorf("failed to create client build directory: %w", err)
	}
	if err := makePublicDir(staticDir); err != nil {
		return fmt.Errorf("failed to create static build directory: %w", err)
	}

	wasmExec, err := readWasmExec()
	if err != nil {
		return err
	}
	if err := writePublicFile(filepath.Join(clientDir, "wasm_exec.js"), wasmExec); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %w", err)
	}

	devBuild := os.Getenv("RFW_DEV_BUILD") == "1"
	skipOptimize := devBuild || utils.IsDebug() || os.Getenv("RFW_SKIP_STRIP") == "1"
	wasmPath := filepath.Join(clientDir, "app.wasm")
	var cmd *exec.Cmd
	switch {
	case devBuild:
		cmd = exec.Command("go", "build", "-tags=rfwdev", "-o", "build/client/app.wasm", ".")
	case skipOptimize:
		cmd = exec.Command("go", "build", "-o", "build/client/app.wasm", ".")
	default:
		cmd = exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", "build/client/app.wasm", ".")
	}
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

	// Build the host binary for SSC when a host directory exists. A static
	// (client-only) build skips it, so the output is a pure static bundle that
	// can be served from a CDN with no live host.
	if manifest.Build.Type != "static" {
		if _, err := os.Stat("host"); err == nil {
			if err := makePublicDir(hostDir); err != nil {
				return fmt.Errorf("failed to create host build directory: %w", err)
			}
			hostCmd := exec.Command("go", "build", "-o", "build/host/host", "./host")
			if hostOutput, err := hostCmd.CombinedOutput(); err != nil {
				if !isDev {
					return fmt.Errorf("failed to build host components: %s: %w", hostOutput, err)
				}
				fmt.Fprintf(os.Stderr, "warning: host build failed (dev mode, continuing): %s\n", hostOutput)
			}
		}
	}
	if err := plugins.Build(); err != nil {
		return fmt.Errorf("failed to run plugins: %w", err)
	}

	// Copy plugin-generated assets (e.g. tailwind.css) to client build dir.
	for _, name := range []string{"tailwind.css", "input.css"} {
		if data, err := readFile(name); err == nil {
			if err := writePublicFile(filepath.Join(clientDir, name), data); err != nil {
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

	wasm, err := readFile(wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read wasm for client config: %w", err)
	}
	wasmHash := sha256.Sum256(wasm)
	if err := writeClientConfig(clientDir, manifest.Build.Host, fmt.Sprintf("%x", wasmHash[:8])); err != nil {
		return fmt.Errorf("failed to write client config: %w", err)
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
			if err := makePublicDir(filepath.Dir(dst)); err != nil {
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

// readWasmExec reads wasm_exec.js from the active Go toolchain.
// It tries the canonical Go 1.21+ path ($GOROOT/lib/wasm/), then the
// legacy path ($GOROOT/misc/wasm/), and finally a project-local copy.
func readWasmExec() ([]byte, error) {
	goRootOutput, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return nil, fmt.Errorf("find Go root: %w", err)
	}
	goRootPath := strings.TrimSpace(string(goRootOutput))
	candidates := []struct {
		root string
		file string
	}{
		{root: filepath.Join(goRootPath, "lib"), file: filepath.Join("wasm", "wasm_exec.js")},
		{root: filepath.Join(goRootPath, "misc"), file: filepath.Join("wasm", "wasm_exec.js")},
	}
	if goRootPath != "" {
		for _, candidate := range candidates {
			resolvedRoot, resolveErr := filepath.EvalSymlinks(candidate.root)
			if resolveErr != nil {
				continue
			}
			goRoot, openErr := os.OpenRoot(resolvedRoot)
			if openErr != nil {
				continue
			}
			data, readErr := goRoot.ReadFile(candidate.file)
			closeErr := goRoot.Close()
			if readErr == nil && closeErr == nil {
				return data, nil
			}
			if readErr == nil {
				return nil, closeErr
			}
		}
	}
	if data, err := readFile("wasm_exec.js"); err == nil {
		return data, nil
	}
	return nil, fmt.Errorf(
		"wasm_exec.js not found in GOROOT (%s) or project root; reinstall Go or run 'rfw init'",
		goRootPath,
	)
}

// writeClientConfig emits build/client/rfw_config.js, which the client loads
// before the wasm to learn its host endpoint for the client-to-host WebSocket.
// It is always written so the index.html include never 404s; the global is set
// only when a host is configured in rfw.json (build.host). The value may be a
// full URL (ws, wss, http, https) or a bare host[:port] with an optional path.
func writeClientConfig(clientDir, host, wasmVersion string) error {
	var b strings.Builder
	b.WriteString("// Generated by rfw build. Do not edit.\n")
	if h := strings.TrimSpace(host); h != "" {
		fmt.Fprintf(&b, "window.RFW_HOST_URL = %q;\n", h)
	}
	fmt.Fprintf(&b, "window.RFW_WASM_VERSION = %q;\n", wasmVersion)
	return writePublicFile(filepath.Join(clientDir, "rfw_config.js"), []byte(b.String()))
}

func readFile(path string) (data []byte, err error) {
	root, file, err := openFile(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if closeErr := root.Close(); err == nil {
			err = closeErr
		}
	}()
	return io.ReadAll(file)
}

func openFile(path string) (*os.Root, *os.File, error) {
	cleaned, err := projectPath(path)
	if err != nil {
		return nil, nil, err
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		return nil, nil, err
	}
	file, err := root.Open(cleaned)
	if err != nil {
		if closeErr := root.Close(); closeErr != nil {
			return nil, nil, closeErr
		}
		return nil, nil, err
	}
	return root, file, nil
}

func projectPath(path string) (string, error) {
	if path == "" || filepath.IsAbs(path) {
		return "", fmt.Errorf("path %q must be relative to the project", path)
	}
	cleaned := filepath.Clean(path)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes the project", path)
	}
	return cleaned, nil
}

func makePublicDir(path string) (err error) {
	cleaned, err := projectPath(path)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := root.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := root.MkdirAll(cleaned, 0o755); err != nil {
		return err
	}
	return nil
}

func writePublicFile(path string, data []byte) (err error) {
	cleaned, err := projectPath(path)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := root.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := root.Remove(cleaned); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := root.OpenFile(cleaned, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}
	return file.Close()
}

func copyFile(src, dst string) (err error) {
	srcRoot, in, err := openFile(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); err == nil {
			err = closeErr
		}
		if closeErr := srcRoot.Close(); err == nil {
			err = closeErr
		}
	}()

	cleanedDestination, err := projectPath(dst)
	if err != nil {
		return err
	}
	dstRoot, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dstRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := dstRoot.Remove(cleanedDestination); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	out, err := dstRoot.OpenFile(cleanedDestination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	return errors.Join(copyErr, out.Close())
}

func compressWasmBrotli(src string) (err error) {
	srcRoot, in, err := openFile(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); err == nil {
			err = closeErr
		}
		if closeErr := srcRoot.Close(); err == nil {
			err = closeErr
		}
	}()

	dst, err := projectPath(src + ".br")
	if err != nil {
		return err
	}
	outputRoot, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := outputRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := outputRoot.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	out, err := outputRoot.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()

	writer := brotli.NewWriterLevel(out, brotli.BestCompression)
	if _, err := io.Copy(writer, in); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}
