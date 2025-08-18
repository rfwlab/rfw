package build

import (
	"fmt"
	"os"
	"os/exec"
)

func Build() error {
	cmd := exec.Command("go", "build", "-o", "main.wasm", "main.go")
	cmd.Env = append(os.Environ(), "GOARCH=wasm", "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build project: %s: %w", output, err)
	}

	return nil
}
