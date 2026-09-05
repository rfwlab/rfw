//go:build !js

// Package initproj creates projects from the embedded rfw scaffold.
package initproj

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const scaffoldGoVersion = "1.26.0"

// InitProject creates a new rfw project from the embedded template.
func InitProject(projectName string, skipTidy bool) (err error) {
	if err := validateModulePath(projectName); err != nil {
		return err
	}

	moduleName := projectName
	projectName = path.Base(moduleName)

	projectPath := projectName

	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return fmt.Errorf("project directory already exists")
	}

	if err := os.Mkdir(projectPath, 0o750); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}
	projectRoot, err := os.OpenRoot(projectPath)
	if err != nil {
		return fmt.Errorf("open project directory: %w", err)
	}
	defer func() {
		if closeErr := projectRoot.Close(); err == nil {
			err = closeErr
		}
	}()

	err = fs.WalkDir(TemplatesFS, "template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "template" {
			return nil
		}

		relPath := strings.TrimPrefix(path, "template/")
		targetPath := strings.TrimSuffix(relPath, ".tmpl")

		if d.IsDir() {
			if err := projectRoot.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			return nil
		}

		content, err := TemplatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		contentStr := string(content)
		contentStr = strings.ReplaceAll(contentStr, "{{moduleName}}", moduleName)
		contentStr = strings.ReplaceAll(contentStr, "{{projectName}}", projectName)

		return writeRootFile(projectRoot, targetPath, []byte(contentStr))
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

func copyWasmExec(projectDir string) (err error) {
	cmd := exec.Command("go", "env", "GOROOT")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get GOROOT: %w", err)
	}
	goRoot := strings.TrimSpace(string(output))

	candidates := []struct {
		root string
		file string
	}{
		{root: filepath.Join(goRoot, "lib"), file: filepath.Join("wasm", "wasm_exec.js")},
		{root: filepath.Join(goRoot, "misc"), file: filepath.Join("wasm", "wasm_exec.js")},
	}
	var input []byte
	for _, candidate := range candidates {
		resolvedRoot, resolveErr := filepath.EvalSymlinks(candidate.root)
		if resolveErr != nil {
			continue
		}
		goRootDirectory, openErr := os.OpenRoot(resolvedRoot)
		if openErr != nil {
			continue
		}
		input, err = goRootDirectory.ReadFile(candidate.file)
		closeErr := goRootDirectory.Close()
		if err == nil && closeErr == nil {
			break
		}
		if err == nil {
			return fmt.Errorf("failed to close Go root: %w", closeErr)
		}
	}
	if input == nil {
		return fmt.Errorf("wasm_exec.js not found in GOROOT (%s); ensure Go is properly installed", goRoot)
	}

	destinationRoot, err := os.OpenRoot(projectDir)
	if err != nil {
		return fmt.Errorf("failed to open project directory: %w", err)
	}
	defer func() {
		if closeErr := destinationRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := writeRootFile(destinationRoot, "wasm_exec.js", input); err != nil {
		return fmt.Errorf("failed to write wasm_exec.js: %w", err)
	}

	return nil
}

func initGoModule(projectPath, moduleName string, skipTidy bool) (err error) {
	if err := validateModulePath(moduleName); err != nil {
		return err
	}
	goMod := fmt.Sprintf("module %s\n\ngo %s\n", moduleName, scaffoldGoVersion)
	projectRoot, err := os.OpenRoot(projectPath)
	if err != nil {
		return fmt.Errorf("open project directory: %w", err)
	}
	defer func() {
		if closeErr := projectRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := writeRootFile(projectRoot, "go.mod", []byte(goMod)); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}

	if !skipTidy {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = projectPath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod tidy failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}

	return nil
}

func validateModulePath(modulePath string) error {
	if modulePath == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if strings.TrimSpace(modulePath) != modulePath || strings.HasPrefix(modulePath, "-") {
		return fmt.Errorf("invalid module path %q", modulePath)
	}
	for _, char := range modulePath {
		if char >= 'a' && char <= 'z' ||
			char >= 'A' && char <= 'Z' ||
			char >= '0' && char <= '9' ||
			strings.ContainsRune("-._/", char) {
			continue
		}
		return fmt.Errorf("invalid character %q in module path", char)
	}
	for _, segment := range strings.Split(modulePath, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("invalid module path segment %q", segment)
		}
		if !isASCIILetterOrDigit(rune(segment[0])) || !isASCIILetterOrDigit(rune(segment[len(segment)-1])) {
			return fmt.Errorf("module path segment %q must start and end with a letter or digit", segment)
		}
		if isReservedPathName(segment) {
			return fmt.Errorf("module path segment %q is reserved", segment)
		}
	}
	return nil
}

func isASCIILetterOrDigit(char rune) bool {
	return char >= 'a' && char <= 'z' ||
		char >= 'A' && char <= 'Z' ||
		char >= '0' && char <= '9'
}

func isReservedPathName(segment string) bool {
	base := strings.ToUpper(strings.SplitN(segment, ".", 2)[0])
	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	return len(base) == 4 &&
		(base[:3] == "COM" || base[:3] == "LPT") &&
		base[3] >= '1' && base[3] <= '9'
}

func writeRootFile(root *os.Root, path string, data []byte) error {
	file, err := root.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}
	return file.Close()
}
