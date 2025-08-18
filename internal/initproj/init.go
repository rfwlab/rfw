package initproj

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InitProject(projectName string) error {
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	projectName = strings.Split(projectName, "/")[len(strings.Split(projectName, "/"))-1]

	projectPath := projectName
	if len(os.Args) > 3 {
		projectPath = os.Args[3]
	}

	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return fmt.Errorf("project directory already exists")
	}

	if err := os.Mkdir(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	err := fs.WalkDir(TemplatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		relPath := strings.TrimPrefix(path, "./")
		targetPath := filepath.Join(projectPath, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := TemplatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		contentStr := strings.ReplaceAll(string(content), "{{packageName}}", projectName)

		return os.WriteFile(targetPath, []byte(contentStr), 0644)
	})
	if err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	if err := copyWasmExec(projectName); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %w", err)
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

	srcPath := filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js")
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
