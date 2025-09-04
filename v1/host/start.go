package host

import (
	"encoding/json"
	"fmt"
	"os"
)

func readPort() int {
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
