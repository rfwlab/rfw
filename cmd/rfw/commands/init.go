package commands

import (
	"fmt"

	"github.com/mirkobrombin/go-cli-builder/v1/command"
	"github.com/rfwlab/rfw/cmd/rfw/initproj"
)

// NewInitCommand returns the init command.
func NewInitCommand() *command.Command {
	cmd := &command.Command{
		Name:        "init",
		Usage:       "init <project-name>",
		Description: "Initialize a new RFW project",
		Run:         runInit,
	}
	cmd.AddBoolFlag("skip-tidy", "", "Skip running go mod tidy", false, false)
	return cmd
}

func runInit(cmd *command.Command, _ *command.RootFlags, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please specify a project name")
	}
	projectName := args[0]
	skipTidy := cmd.GetFlagBool("skip-tidy")
	if err := initproj.InitProject(projectName, skipTidy); err != nil {
		return err
	}
	cmd.Logger.Success("Project initialized")
	return nil
}
