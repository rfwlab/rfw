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
	"strconv"
	"strings"

	"github.com/rfwlab/rfw/v2/cmd/rfw/initproj"
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

// defaultPlugins is the plugin set activated when rfw.json omits the "plugins"
// key. It enables file-based routing (the pages plugin) so a scaffolded project
// routes without extra configuration. Other plugins stay opt-in.
func defaultPlugins() map[string]json.RawMessage {
	return map[string]json.RawMessage{"pages": json.RawMessage("{}")}
}

// pluginsConfig resolves the plugin configuration to apply. A missing "plugins"
// key (nil) falls back to defaultPlugins; an explicit block, including an empty
// one, is honored as written, so "plugins": {} opts out of every plugin.
func pluginsConfig(explicit map[string]json.RawMessage) map[string]json.RawMessage {
	if explicit == nil {
		return defaultPlugins()
	}
	return explicit
}

// Build compiles the configured application and runs its build plugins.
func Build() error {
	shape := buildShape{delivery: deliveryNetwork, transport: sscTransportBrowser, hostWire: hostTransportWebSocket}
	if data, readErr := os.ReadFile("rfw.json"); readErr == nil {
		decoded, err := decodeBuildShape(data)
		if err != nil {
			return err
		}
		shape = decoded
	}
	if configured := strings.TrimSpace(os.Getenv("RFW_TRANSPORT")); configured != "" {
		transport, err := parseHostTransport(configured)
		if err != nil {
			return err
		}
		shape.hostWire = transport
	}
	if err := plugins.Configure(pluginsConfig(shape.plugins)); err != nil {
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

	// A non-static build links the SSC client so component registration keeps
	// working. core no longer imports hostclient, so nothing else would pull it
	// in, and a "type": "static" project drops the websocket and net/http
	// stacks from its bundle.
	if !shape.static {
		sscFile, err := writeSSCImport()
		if err != nil {
			return fmt.Errorf("failed to write SSC import: %w", err)
		}
		defer func() { _ = os.Remove(sscFile) }()
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
	encodings, err := prepareWasmArtifacts(wasmPath, shape, isDev)
	if err != nil {
		return err
	}

	// Build the host binary for SSC when a host directory exists. A static
	// (client-only) build skips it, so the output is a pure static bundle that
	// can be served from a CDN with no live host.
	if !shape.static {
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

	if err := writeWasmLoader(clientDir); err != nil {
		return err
	}

	wasm, err := readFile(wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read wasm for client config: %w", err)
	}
	wasmHash := sha256.Sum256(wasm)
	version := fmt.Sprintf("%x", wasmHash[:8])
	// A live host negotiates the encoding from Accept-Encoding. A static build
	// is served by something that will not, so the client picks an artifact by
	// name instead, and an embedded build is read from the container's local
	// asset handler, which negotiates nothing either.
	if err := writeClientConfig(clientDir, shape.host, version, encodings, shape.negotiates(), !isDev, shape.delivery, shape.transport, shape.hostWire); err != nil {
		return fmt.Errorf("failed to write client config: %w", err)
	}
	if err := stampIndexHTML(filepath.Join(clientDir, "index.html"), version); err != nil {
		return fmt.Errorf("failed to version the bootstrap scripts: %w", err)
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

func writeWasmLoader(clientDir string) error {
	destination := filepath.Join(clientDir, "wasm_loader.js")
	if _, err := os.Stat("wasm_loader.js"); err == nil {
		if err := copyFile("wasm_loader.js", destination); err != nil {
			return fmt.Errorf("failed to copy wasm_loader.js: %w", err)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect wasm_loader.js: %w", err)
	}

	loader, err := initproj.TemplatesFS.ReadFile("template/wasm_loader.js")
	if err != nil {
		return fmt.Errorf("failed to read the framework wasm loader: %w", err)
	}
	if err := writePublicFile(destination, loader); err != nil {
		return fmt.Errorf("failed to write the framework wasm loader: %w", err)
	}
	return nil
}

// sscImportFile is generated before the wasm build and removed after it, the
// same lifecycle the pages plugin uses for routes_gen.go.
const sscImportFile = "rfw_ssc_gen.go"

// writeSSCImport emits a blank import of the hostclient package into the
// project root so linking it installs SSC component registration into core.
// It returns the generated path so the caller can remove it.
func writeSSCImport() (string, error) {
	const src = `//go:build js && wasm

// Code generated by rfw build. DO NOT EDIT.

package main

// Linking hostclient installs SSC component registration into core. A project
// built with "type": "static" skips this file, which keeps the websocket and
// net/http stacks out of the bundle.
import _ "github.com/rfwlab/rfw/v2/hostclient"
`
	if err := os.WriteFile(sscImportFile, []byte(src), 0o600); err != nil {
		return "", err
	}
	return sscImportFile, nil
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
// writeClientConfig describes the build to the loader. Every name it emits is
// defined: a loader that reads an undefined global cannot tell "off" from
// "absent", which is how an application ends up silently downloading the raw
// bundle it has a compressed copy of. That is also why the delivery mode is
// stamped explicitly instead of being inferred from an empty encoding list,
// which a failed compression step would produce too.
func writeClientConfig(clientDir, host, wasmVersion string, encodings []Encoding, negotiated, production bool, deliveryMode delivery, transport sscTransport, hostWire hostTransport) error {
	var b strings.Builder
	b.WriteString("// Generated by rfw build. Do not edit.\n")
	if h := strings.TrimSpace(host); h != "" {
		fmt.Fprintf(&b, "window.RFW_HOST_URL = %q;\n", h)
	}
	fmt.Fprintf(&b, "window.RFW_WASM_VERSION = %q;\n", wasmVersion)
	fmt.Fprintf(&b, "window.RFW_WASM_DELIVERY = %q;\n", string(deliveryMode))
	names := make([]string, 0, len(encodings))
	for _, encoding := range encodings {
		names = append(names, strconv.Quote(string(encoding)))
	}
	fmt.Fprintf(&b, "window.RFW_WASM_ENCODINGS = [%s];\n", strings.Join(names, ", "))
	fmt.Fprintf(&b, "window.RFW_WASM_NEGOTIATED = %t;\n", negotiated)
	fmt.Fprintf(&b, "window.RFW_SSC_TRANSPORT = %q;\n", string(transport))
	fmt.Fprintf(&b, "window.RFW_TRANSPORT = %q;\n", string(hostWire))
	mode := "development"
	if production {
		mode = "production"
	}
	fmt.Fprintf(&b, "window.RFW_BUILD_MODE = %q;\n", mode)
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
