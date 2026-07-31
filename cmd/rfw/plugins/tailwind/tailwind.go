//go:build !js

// Package tailwind registers Tailwind CSS compilation.
package tailwind

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rfwlab/rfw/v2/cmd/rfw/logging"
	"github.com/rfwlab/rfw/v2/cmd/rfw/plugins"
)

type plugin struct {
	output string
}

func init() {
	plugins.Register(&plugin{})
}

func (p *plugin) Name() string { return "tailwind" }

func (p *plugin) Priority() int { return 0 }

func (p *plugin) Build(raw json.RawMessage) error {
	logging.Log.Info("starting build", logging.F("plugin", "tailwind"))
	cfg := struct {
		Input  string   `json:"input"`
		Output string   `json:"output"`
		Minify bool     `json:"minify"`
		Args   []string `json:"args"`
	}{
		Input:  "index.css",
		Output: "tailwind.css",
		Minify: true,
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("decode tailwind plugin config: %w", err)
		}
	}
	input, err := validateProjectFile(cfg.Input, true)
	if err != nil {
		return fmt.Errorf("invalid tailwind input: %w", err)
	}
	output, err := validateProjectFile(cfg.Output, false)
	if err != nil {
		return fmt.Errorf("invalid tailwind output: %w", err)
	}
	if input == output {
		return fmt.Errorf("tailwind input and output must differ")
	}
	useConfig, err := validateExtraArgs(cfg.Args)
	if err != nil {
		return err
	}
	p.output = output

	root, err := os.OpenRoot(".")
	if err != nil {
		return fmt.Errorf("open project root: %w", err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			logging.Log.Error("close project root", logging.F("plugin", "tailwind"), logging.F("error", closeErr.Error()))
		}
	}()
	inputCSS, err := root.ReadFile(input)
	if err != nil {
		return fmt.Errorf("read tailwind input: %w", err)
	}

	cmd := tailwindCommand(cfg.Minify, useConfig)
	cmd.Stdin = bytes.NewReader(inputCSS)
	var generatedCSS bytes.Buffer
	var diagnostics bytes.Buffer
	cmd.Stdout = &generatedCSS
	cmd.Stderr = &diagnostics
	logging.Log.Info("running command", logging.F("plugin", "tailwind"), logging.F("args", strings.Join(cmd.Args, " ")))
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			logging.Log.Warn("tailwindcss not found, please install it manually", logging.F("plugin", "tailwind"))
		}
		return fmt.Errorf("tailwind build failed: %s: %w", strings.TrimSpace(diagnostics.String()), err)
	}
	if err := writeProjectFile(root, output, generatedCSS.Bytes()); err != nil {
		return fmt.Errorf("write tailwind output: %w", err)
	}
	logging.Log.Info("build complete", logging.F("plugin", "tailwind"))
	return nil
}

func tailwindCommand(minify, useConfig bool) *exec.Cmd {
	switch {
	case minify && useConfig:
		return exec.Command("tailwindcss", "-i", "-", "--minify", "--config", "tailwind.config.js")
	case minify:
		return exec.Command("tailwindcss", "-i", "-", "--minify")
	case useConfig:
		return exec.Command("tailwindcss", "-i", "-", "--config", "tailwind.config.js")
	default:
		return exec.Command("tailwindcss", "-i", "-")
	}
}

func writeProjectFile(root *os.Root, filePath string, content []byte) error {
	directory := filepath.Dir(filePath)
	if directory != "." {
		if err := root.MkdirAll(directory, 0o755); err != nil {
			return err
		}
	}
	if err := root.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := root.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(content)
	closeErr := file.Close()
	return errors.Join(writeErr, closeErr)
}

func validateProjectFile(filePath string, mustExist bool) (string, error) {
	if filePath == "" || filepath.IsAbs(filePath) || strings.HasPrefix(filePath, "-") {
		return "", fmt.Errorf("path %q must be a project-relative file", filePath)
	}
	cleaned := filepath.Clean(filePath)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes the project", filePath)
	}

	projectRoot, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return "", err
	}
	candidate := filepath.Join(resolvedRoot, cleaned)
	checkPath := candidate
	if !mustExist {
		for {
			if _, statErr := os.Lstat(checkPath); statErr == nil {
				break
			} else if !os.IsNotExist(statErr) {
				return "", statErr
			}
			checkPath = filepath.Dir(candidate)
			if checkPath == resolvedRoot {
				break
			}
			candidate = checkPath
		}
	}
	resolved, err := filepath.EvalSymlinks(checkPath)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q resolves outside the project", filePath)
	}
	return cleaned, nil
}

func validateExtraArgs(args []string) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}
	if len(args) != 2 || args[0] != "--config" && args[0] != "-c" {
		return false, fmt.Errorf("tailwind args only support the project-root tailwind.config.js")
	}
	configPath, err := validateProjectFile(args[1], true)
	if err != nil {
		return false, fmt.Errorf("invalid tailwind config: %w", err)
	}
	if configPath != "tailwind.config.js" {
		return false, fmt.Errorf("tailwind config must be tailwind.config.js in the project root")
	}
	return true, nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	if strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, p.output) {
		logging.Log.Info("rebuild triggered", logging.F("plugin", "tailwind"), logging.F("path", path))
		return true
	}
	if strings.HasSuffix(path, ".rtml") || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".go") {
		logging.Log.Info("rebuild triggered", logging.F("plugin", "tailwind"), logging.F("path", path))
		return true
	}
	return false
}
