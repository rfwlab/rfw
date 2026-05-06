package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rfwlab/rfw/v2/core"
)

const githubRepo = "rfwlab/rfw"
const checkFile = ".rfw-update-check"
const checkInterval = 24 * time.Hour

var (
	boldCyan = color.New(color.FgCyan, color.Bold).SprintFunc()
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
		Name               string `json:"name"`
	} `json:"assets"`
}

func fetchLatestVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/"+githubRepo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 3*time.Second)
		},
	}}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func shouldCheckUpdate() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return true
	}
	path := home + "/" + checkFile
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > checkInterval
}

func markChecked() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := home + "/" + checkFile
	os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0644)
}

func isNewer(current, latest string) bool {
	c := strings.TrimPrefix(current, "v")
	l := strings.TrimPrefix(latest, "v")
	if c == "" || l == "" {
		return false
	}
	partsC := strings.SplitN(c, ".", 3)
	partsL := strings.SplitN(l, ".", 3)
	for i := 0; i < 3; i++ {
		if i >= len(partsC) || i >= len(partsL) {
			break
		}
		if partsL[i] > partsC[i] {
			return true
		}
		if partsL[i] < partsC[i] {
			return false
		}
	}
	return false
}

func downloadAndReplace(assetURL string) error {
	tmp, err := os.CreateTemp("", "rfw-update-*")
	if err != nil {
		return err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", assetURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	if err := os.Chmod(tmp.Name(), 0755); err != nil {
		return err
	}

	if err := os.Rename(tmp.Name(), exePath); err != nil {
		return os.WriteFile(exePath, mustReadFile(tmp.Name()), 0755)
	}

	return nil
}

func mustReadFile(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	return data
}

func getAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("rfw-%s-%s%s", goos, goarch, ext)
}

func CheckForUpdate() {
	if !shouldCheckUpdate() {
		return
	}

	latest, err := fetchLatestVersion()
	if err != nil {
		return
	}
	markChecked()

	if !isNewer(core.Version, latest) {
		return
	}

	assetName := getAssetName()
	fmt.Println()
	Info(fmt.Sprintf("Update available: %s → %s", faint(core.Version), boldCyan(latest)))

	fmt.Print(indent, red("➜ "), bold("Update now? [y/N] "))
	var answer string
	fmt.Scanln(&answer)
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println(indent, faint("Skipped."))
		return
	}

	assetURL := ""
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/"+githubRepo+"/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Info("Failed to fetch release info")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var release githubRelease
	json.Unmarshal(body, &release)

	for _, a := range release.Assets {
		if a.Name == assetName {
			assetURL = a.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		Info(fmt.Sprintf("No binary found for %s/%s", runtime.GOOS, runtime.GOARCH))
		return
	}

	Info("Downloading...")
	if err := downloadAndReplace(assetURL); err != nil {
		Info(fmt.Sprintf("Update failed: %v", err))
		return
	}

	Info(fmt.Sprintf("Updated to %s!", latest))
}