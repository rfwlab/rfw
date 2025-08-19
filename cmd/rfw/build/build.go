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
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/tailwind"
	_ "github.com/rfwlab/rfw/cmd/rfw/plugins/test"
)

func Build() error {
	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return fmt.Errorf("failed to get GOROOT: %w", err)
	}
	wasmExec := filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js")
	if err := copyFile(wasmExec, "wasm_exec.js"); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", "main.wasm", "main.go")
	cmd.Env = append(os.Environ(), "GOARCH=wasm", "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build project: %s: %w", output, err)
	}

	var manifest struct {
		Plugins map[string]json.RawMessage `json:"plugins"`
	}
	if data, err := os.ReadFile("rfw.json"); err == nil {
		_ = json.Unmarshal(data, &manifest)
	}
	if err := plugins.Configure(manifest.Plugins); err != nil {
		return fmt.Errorf("failed to run plugins: %w", err)
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
