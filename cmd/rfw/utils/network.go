package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/browser"
)

func GetLocalIP() (string, error) {
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

func OpenBrowser(url string) error {
	if configured := os.Getenv("BROWSER"); configured != "" {
		cmd := strings.Fields(configured)[0]
		if _, err := exec.LookPath(cmd); err != nil {
			return err
		}
	}
	return browser.OpenURL(url)
}
