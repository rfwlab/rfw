package commands

import (
	"github.com/mirkobrombin/go-cli-builder/v1/command"
	"github.com/rfwlab/rfw/cmd/rfw/server"
)

// NewDevCommand returns the dev command.
func NewDevCommand() *command.Command {
	cmd := &command.Command{
		Name:        "dev",
		Usage:       "dev [--port <port>] [--host]",
		Description: "Start the development server",
		Run:         runDev,
	}
	cmd.AddFlag("port", "p", "Port to serve on", "8080", true)
	cmd.AddBoolFlag("host", "", "Expose the server to the network", false, false)
	return cmd
}

func runDev(cmd *command.Command, _ *command.RootFlags, _ []string) error {
	port := cmd.GetFlagString("port")
	host := cmd.GetFlagBool("host")
	srv := server.NewServer(port, host)
	return srv.Start()
}
