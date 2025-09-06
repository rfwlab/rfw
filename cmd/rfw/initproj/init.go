package initproj

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func InitProject(projectName string, skipTidy bool) error {
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	moduleName := projectName
	projectName = path.Base(moduleName)

	projectPath := projectName

	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return fmt.Errorf("project directory already exists")
	}

	if err := os.Mkdir(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	err := fs.WalkDir(TemplatesFS, "template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "template" {
			return nil
		}

		relPath := strings.TrimPrefix(path, "template/")
		targetPath := filepath.Join(projectPath, relPath)
		if strings.HasSuffix(targetPath, ".tmpl") {
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")
		}

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := TemplatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		contentStr := string(content)
		contentStr = strings.ReplaceAll(contentStr, "{{moduleName}}", moduleName)
		contentStr = strings.ReplaceAll(contentStr, "{{projectName}}", projectName)

		return os.WriteFile(targetPath, []byte(contentStr), 0644)
	})
	if err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	if err := copyWasmExec(projectPath); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %w", err)
	}

	if err := initGoModule(projectPath, moduleName, skipTidy); err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}

	fmt.Printf("Project '%s' initialized successfully.\n", projectName)
	return nil
}

func copyWasmExec(projectDir string) error {
	cmd := exec.Command("go", "env", "GOROOT")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get GOROOT: %w", err)
	}
	goRoot := strings.TrimSpace(string(output))

	srcPath := filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	destPath := filepath.Join(projectDir, "wasm_exec.js")

	input, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read wasm_exec.js: %w", err)
	}

	if err := os.WriteFile(destPath, input, 0644); err != nil {
		return fmt.Errorf("failed to write wasm_exec.js: %w", err)
	}

	return nil
}

func initGoModule(projectPath, moduleName string, skipTidy bool) error {
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = projectPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod init failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	if !skipTidy {
		cmd = exec.Command("go", "mod", "tidy")
		cmd.Dir = projectPath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod tidy failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}

	return nil
}
