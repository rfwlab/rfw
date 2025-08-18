package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/rfwlab/rfw-cli/internal/build"
	"github.com/rfwlab/rfw-cli/internal/initproj"
	"github.com/rfwlab/rfw-cli/internal/server"
	"github.com/rfwlab/rfw-cli/internal/utils"
)

func Execute() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	switch os.Args[1] {
	case "init":
		initProject(os.Args[2:])
	case "dev":
		startServer(os.Args[2:])
	case "build":
		buildProject(os.Args[2:])
	case "-h", "--help":
		showHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		showHelp()
	}
}

func startServer(args []string) {
	devFlags := flag.NewFlagSet("dev", flag.ExitOnError)
	port := devFlags.String("port", "8080", "Port from which the server will serve")
	host := devFlags.Bool("host", false, "Expose the server to the network")

	err := devFlags.Parse(args)
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	fmt.Println("Starting server on port", *port)

	srv := server.NewServer(*port, *host)
	if err := srv.Start(); err != nil {
		utils.Fatal("Server failed to start: ", err)
	}
}

func buildProject(args []string) {
	buildFlags := flag.NewFlagSet("build", flag.ExitOnError)

	err := buildFlags.Parse(args)
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	if err := build.Build(); err != nil {
		utils.Fatal("Failed to build project: ", err)
	}
}

func initProject(args []string) {
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)

	err := initFlags.Parse(args)
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	remainingArgs := initFlags.Args()
	if len(remainingArgs) < 1 {
		fmt.Println("Please specify a project name: rfw-cli init <project-name>")
		return
	}

	projectName := remainingArgs[0]
	if err := initproj.InitProject(projectName); err != nil {
		utils.Fatal("Failed to initialize project: ", err)
	}
}

func showHelp() {
	helpMessage := `
Usage:
  rfw-cli <command> [options]

Commands:
  init <project-name>           Initialize a new project
  dev [--port <port>] [--host]  Start the development server
  -h, --help                    Show this help message

Examples:
  rfw-cli init my-project
  rfw-cli dev --port 9090 --host
`
	fmt.Println(helpMessage)
}
