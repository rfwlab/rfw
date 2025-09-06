package main

import (
	"testing"

	root "github.com/mirkobrombin/go-cli-builder/v1/root"
	"github.com/rfwlab/rfw/cmd/rfw/commands"
)

// TestRootCommandSetup ensures main registers expected subcommands.
func TestRootCommandSetup(t *testing.T) {
	cmd := root.NewRootCommand("rfw", "rfw [command]", "RFW command line interface", "0.0.0")
	cmd.AddCommand(commands.NewInitCommand())
	cmd.AddCommand(commands.NewDevCommand())
	cmd.AddCommand(commands.NewBuildCommand())
	for _, name := range []string{"init", "dev", "build"} {
		if _, ok := cmd.Commands[name]; !ok {
			t.Fatalf("expected subcommand %s registered", name)
		}
	}
}
