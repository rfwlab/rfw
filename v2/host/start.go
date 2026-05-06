package host

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readPort() int {
	if override := strings.TrimSpace(os.Getenv("RFW_HOST_PORT")); override != "" {
		if p, err := strconv.Atoi(override); err == nil && p > 0 {
			return p
		}
	}
	var manifest struct {
		Port int `json:"port"`
	}
	data, err := os.ReadFile("rfw.json")
	if err != nil {
		return 8080
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return 8080
	}
	if manifest.Port == 0 {
		return 8080
	}
	return manifest.Port
}

// StartAuto launches HTTP and HTTPS servers serving files from the default
// client build directory. It resolves the root path from rfw.json or falls
// back to "build/client". This is the recommended way to start the host server.
func StartAuto() error {
	root := resolveRoot()
	return Start(root)
}

func resolveRoot() string {
	// Check rfw.json for build configuration.
	var manifest struct {
		Build struct {
			Dir string `json:"dir"`
		} `json:"build"`
	}
	if data, err := os.ReadFile("rfw.json"); err == nil {
		_ = json.Unmarshal(data, &manifest)
		if manifest.Build.Dir != "" {
			return manifest.Build.Dir
		}
	}
	return "build/client"
}

// Start launches HTTP and HTTPS servers serving files from root.
// The HTTPS port is the HTTP port + 1.
func Start(root string) error {
	port := readPort()
	httpsPort := port + 1

	go func() {
		addr := fmt.Sprintf(":%d", port)
		if err := ListenAndServe(addr, root); err != nil {
			logger.Error("HTTP server error", "err", err)
		}
	}()

	httpsAddr := fmt.Sprintf(":%d", httpsPort)
	return ListenAndServeTLS(httpsAddr, root)
}
