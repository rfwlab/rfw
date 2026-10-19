//go:build !js

package commands

import (
	"fmt"

	"github.com/mirkobrombin/go-cli-builder/v1/command"
	"github.com/rfwlab/rfw/v2/cmd/rfw/initproj"
	"github.com/rfwlab/rfw/v2/core"
)

// NewInitCommand returns the init command.
func NewInitCommand() *command.Command {
	cmd := &command.Command{
		Name:        "init",
		Usage:       "init [--rfw-version <version>] [--skip-tidy] <project-name>",
		Description: "Initialize a new rfw project",
		Run:         runInit,
	}
	cmd.AddFlag("rfw-version", "", "Framework v2 tag or pseudo-version (default: CLI version)", "", false)
	cmd.AddBoolFlag("skip-tidy", "", "Skip running go mod tidy", false, false)
	return cmd
}

func runInit(cmd *command.Command, _ *command.RootFlags, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please specify a project name")
	}
	if len(args) != 1 {
		return fmt.Errorf("specify one project name; place options before the project name")
	}
	projectName := args[0]
	skipTidy := cmd.GetFlagBool("skip-tidy")
	rfwVersion := cmd.GetFlagString("rfw-version")
	if rfwVersion == "" {
		rfwVersion = core.Version()
	}
	if err := initproj.InitProjectWithVersion(projectName, skipTidy, rfwVersion); err != nil {
		return err
	}
	cmd.Logger.Success("Project initialized")
	return nil
}
