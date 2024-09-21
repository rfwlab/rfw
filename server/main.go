package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
)

const defaultPort = "8080"

var (
	green      = color.New(color.FgGreen).SprintFunc()
	red        = color.New(color.FgRed).SprintFunc()
	white      = color.New(color.FgWhite).SprintFunc()
	bold       = color.New(color.FgWhite, color.Bold).SprintFunc()
	faint      = color.New(color.FgWhite, color.Faint).SprintFunc()
	faintGreen = color.New(color.FgGreen, color.Faint).SprintFunc()
	boldGreen  = color.New(color.FgGreen, color.Bold).SprintFunc()
	boldYellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	boldRed    = color.New(color.FgRed, color.Bold).SprintFunc()
	indent     = "  "
	port       = defaultPort
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	fs := http.FileServer(http.Dir("."))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", faint(time.Now().Format("15:04:05")), boldYellow("serving"), faint(r.URL.Path))
		handleFileRequest(w, r, fs)
	})

	localIP, err := getLocalIP()
	if err != nil {
		log.Fatalf("Failed to get local IP address: %v", err)
	}

	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "--port=") {
		port = strings.TrimPrefix(os.Args[1], "--port=")
	}

	clearScreen()
	printStartupInfo(localIP)

	go startServer()
	go listenForShutdown(stop)

	<-stop
	fmt.Println(boldGreen("[rfw]"), "Server stopped.")
}

func handleFileRequest(w http.ResponseWriter, r *http.Request, fs http.Handler) {
	if _, err := os.Stat("." + r.URL.Path); os.IsNotExist(err) {
		http.ServeFile(w, r, "./index.html")
	} else {
		fs.ServeHTTP(w, r)
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
	fmt.Println()
}

func printStartupInfo(localIP string) {
	fmt.Println(indent, boldGreen("rfw"), faint("v0.0.0"))
	fmt.Println()
	fmt.Println(indent, green("➜ "), bold("Local:"), green(fmt.Sprintf("http://localhost:%s/", port)))

	if len(os.Args) > 1 && os.Args[1] == "--host" {
		fmt.Println(indent, green("➜ "), faint(bold("Network:")), white(fmt.Sprintf("http://%s:%s/", localIP, port)))
	} else {
		fmt.Println(indent, green("➜ "), faint(bold("Network:")), white("--host"), faint("to expose"))
	}

	fmt.Println(indent, faintGreen("➜ "), faint("Press"), bold("h + enter"), faint("to show help"))
	fmt.Println()
}

func startServer() {
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf(boldGreen("[rfw]"), "Server failed: %v", err)
	}
}

func listenForShutdown(stop chan os.Signal) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "h" {
			clearScreen()
			fmt.Println()
			fmt.Println(indent, green("➜ "), bold("Help"))
			fmt.Println(indent, indent, red("➜ "), bold("Shortcuts"))
			fmt.Println(indent, indent, indent, faint("Press"), bold("c + enter"), faint("to stop the server"))
			fmt.Println(indent, indent, indent, faint("Press"), bold("o + enter"), faint("to open the browser"))
			fmt.Println(indent, indent, indent, faint("Press"), bold("u + enter"), faint("to show the startup info and clear logs"))
			fmt.Println(indent, indent, indent, faint("Press"), bold("h + enter"), faint("to show this help"))
			fmt.Println(indent, indent, red("➜ "), bold("Flags"))
			fmt.Println(indent, indent, indent, faint("Use"), bold("--host"), faint("to expose the server to the network"))
			fmt.Println(indent, indent, indent, faint("Use"), bold("--port=XXXX"), faint("to specify a port"))
			fmt.Println()
		}

		if strings.ToLower(input) == "u" {
			fmt.Println("Showing the startup info and clear logs...")
			clearScreen()
			localIP, err := getLocalIP()
			if err != nil {
				log.Fatalf("Failed to get local IP address: %v", err)
			}
			printStartupInfo(localIP)
		}

		/*
			// This feature is disabled because it's not working properly yet.

			if strings.ToLower(input) == "r" {
				fmt.Println("Reloading the server...")
				clearScreen()
				go startServer()
			}
		*/

		if strings.ToLower(input) == "c" || strings.ToLower(input) == "q" {
			fmt.Println("Closing the server...")
			close(stop)
			break
		}

		if strings.ToLower(input) == "o" {
			fmt.Println(boldGreen("[rfw]"), "Opening the browser...")
			fmt.Println()
			url := fmt.Sprintf("http://localhost:%s/", port)
			openBrowser(url)
		}
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch {
	case isWindows():
		cmd = exec.Command("cmd", "/c", "start", url)
	case isMac():
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}

func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")
}

func isMac() bool {
	return strings.Contains(strings.ToLower(os.Getenv("HOME")), "/users/")
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}

	return "", fmt.Errorf("no local IP address found")
}
