package tailwind

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/rfwlab/rfw/cmd/rfw/plugins"
)

type plugin struct{}

func init() {
	plugins.Register(&plugin{})
}

func (p *plugin) Name() string { return "tailwind" }

func (p *plugin) Build() error {
	log.Printf("tailwind: starting build")
	bin, err := exec.LookPath("tailwindcss")
	if err != nil {
		log.Printf("tailwind: tailwindcss not found, please install it manually")
		return err
	}

	// FIXME: in future an rfw project should have a root manifest file with plugins configurations and so on,
	// for the moment we will look for an input.css file in the project root
	args := []string{
		"-i", "input.css",
		"-o", "tailwind.css",
		"--minify",
	}

	log.Printf("tailwind: running %s %s", bin, strings.Join(args, " "))
	cmd := exec.Command(bin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailwind build failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	log.Printf("tailwind: build complete")
	return nil
}

func (p *plugin) ShouldRebuild(path string) bool {
	if strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, "tailwind.css") {
		log.Printf("tailwind: rebuild triggered by %s", path)
		return true
	}
	if strings.HasSuffix(path, ".rtml") || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".go") {
		log.Printf("tailwind: rebuild triggered by %s", path)
		return true
	}
	return false
}
