package commands

import (
	"encoding/json"
	"os"
	"strconv"

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
	cmd.AddFlag("port", "p", "Port to serve on", "", true)
	cmd.AddBoolFlag("host", "", "Expose the server to the network", false, false)
	cmd.AddBoolFlag("debug", "", "Enable debug logs, profiling and overlay", false, false)
	return cmd
}

func runDev(cmd *command.Command, _ *command.RootFlags, _ []string) error {
	port := os.Getenv("RFW_PORT")
	if port == "" {
		port = cmd.GetFlagString("port")
		if port == "" {
			port = readPortFromManifest()
			if port == "" {
				port = "8080"
			}
		}
	}
	host := cmd.GetFlagBool("host")
	debug := cmd.GetFlagBool("debug")
	if debug {
		os.Setenv("RFW_DEVTOOLS", "1")
	}
	srv := server.NewServer(port, host, debug)
	return srv.Start()
}

func readPortFromManifest() string {
	var manifest struct {
		Port int `json:"port"`
	}
	data, err := os.ReadFile("rfw.json")
	if err != nil {
		return ""
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return ""
	}
	if manifest.Port == 0 {
		return ""
	}
	return strconv.Itoa(manifest.Port)
}
