package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fatih/color"
)

var (
	red        = color.New(color.FgRed).SprintFunc()
	white      = color.New(color.FgWhite).SprintFunc()
	yellow     = color.New(color.FgYellow).SprintFunc()
	bold       = color.New(color.FgWhite, color.Bold).SprintFunc()
	faint      = color.New(color.FgWhite, color.Faint).SprintFunc()
	faintRed   = color.New(color.FgRed, color.Faint).SprintFunc()
	boldYellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	boldRed    = color.New(color.FgRed, color.Bold).SprintFunc()
	indent     = "  "
)

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
	fmt.Println()
}

func PrintStartupInfo(port, localIP string, host bool) {
	fmt.Println(indent, boldRed("rfw"), faint("v0.0.0"))
	fmt.Println()
	fmt.Println(indent, red("➜ "), bold("Local:"), red(fmt.Sprintf("http://localhost:%s/", port)))

	if host {
		fmt.Println(indent, red("➜ "), faint(bold("Network:")), white(fmt.Sprintf("http://%s:%s/", localIP, port)))
	} else {
		fmt.Println(indent, red("➜ "), faint(bold("Network:")), white("--host"), faint("to expose"))
	}

	fmt.Println(indent, faintRed("➜ "), faint("Press"), bold("h + enter"), faint("to show help"))
	fmt.Println()
}

func LogServeRequest(r *http.Request) {
	fmt.Printf("%s %s %s\n", faint(time.Now().Format("15:04:05")), boldYellow("serving"), faint(r.URL.Path))
}

func Info(message string) {
	fmt.Println(boldRed("[rfw]"), message)
}

func Fatal(message string, err error) {
	log.Fatalf(boldRed("[rfw] "), message, err)
}

func PrintHelp() {
	ClearScreen()
	fmt.Println()
	fmt.Println(indent, red("➜ "), bold("Help"))
	fmt.Println(indent, indent, yellow("➜ "), bold("Shortcuts"))
	fmt.Println(indent, indent, indent, faint("Press"), bold("c + enter"), faint("to stop the server"))
	fmt.Println(indent, indent, indent, faint("Press"), bold("o + enter"), faint("to open the browser"))
	fmt.Println(indent, indent, indent, faint("Press"), bold("u + enter"), faint("to show the startup info and clear logs"))
	fmt.Println(indent, indent, indent, faint("Press"), bold("h + enter"), faint("to show this help"))
	fmt.Println(indent, indent, yellow("➜ "), bold("Flags"))
	fmt.Println(indent, indent, indent, faint("Use"), bold("--host"), faint("to expose the server to the network"))
	fmt.Println(indent, indent, indent, faint("Use"), bold("--port=XXXX"), faint("to specify a port"))
	fmt.Println()
}
