package commands

import (
	"github.com/mirkobrombin/go-cli-builder/v1/command"
	"github.com/rfwlab/rfw/cmd/rfw/build"
)

// NewBuildCommand returns the build command.
func NewBuildCommand() *command.Command {
	cmd := &command.Command{
		Name:        "build",
		Usage:       "build",
		Description: "Build the current project",
		Run:         runBuild,
	}
	return cmd
}

func runBuild(cmd *command.Command, _ *command.RootFlags, _ []string) error {
	if err := build.Build(nil); err != nil {
		return err
	}
	cmd.Logger.Success("Build completed")
	return nil
}
