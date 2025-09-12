package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/assets"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/bundler"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/copy"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/devtools"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/docs"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/env"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/pages"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/tailwind"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/test"
)

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
	if os.Getenv("RFW_DEVTOOLS") == "1" {
		if manifest.Plugins == nil {
			manifest.Plugins = map[string]json.RawMessage{}
		}
		manifest.Plugins["devtools"] = nil
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

	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return fmt.Errorf("failed to get GOROOT: %w", err)
	}
	wasmExec := filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js")
	if err := copyFile(wasmExec, filepath.Join(clientDir, "wasm_exec.js")); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %w", err)
	}

	args := []string{"build"}
	if os.Getenv("RFW_DEVTOOLS") == "1" {
		args = append(args, "-tags=devtools")
	}
	args = append(args, "-o", filepath.Join(clientDir, "app.wasm"), ".")
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GOARCH=wasm", "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build project: %s: %w", output, err)
	}

	if err := os.MkdirAll(hostDir, 0o755); err != nil {
		return fmt.Errorf("failed to create host build directory: %w", err)
	}
	hostCmd := exec.Command("go", "build", "-o", filepath.Join(hostDir, "host"), "./host")
	if hostOutput, err := hostCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build host components: %s: %w", hostOutput, err)
	}
	if err := plugins.Build(); err != nil {
		return fmt.Errorf("failed to run plugins: %w", err)
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
