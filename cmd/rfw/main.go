package main

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/v1/root"
	"github.com/rfwlab/rfw/cmd/rfw/commands"
)

func main() {
	rootCmd := root.NewRootCommand("rfw", "rfw [command]", "RFW command line interface", "0.0.0")

	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewDevCommand())
	rootCmd.AddCommand(commands.NewBuildCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
