//go:build !js

package utils

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/internal/safehttp"
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
		Digest             string `json:"digest"`
		Name               string `json:"name"`
	} `json:"assets"`
}

func fetchLatestRelease() (githubRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := safehttp.NewRequest(
		ctx,
		http.MethodGet,
		"https://api.github.com/repos/"+githubRepo+"/releases/latest",
	)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := safehttp.NewClient().Do(req)
	if err != nil {
		return githubRelease{}, err
	}

	if resp.StatusCode != http.StatusOK {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return githubRelease{}, closeErr
		}
		return githubRelease{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		return githubRelease{}, err
	}
	if closeErr != nil {
		return githubRelease{}, closeErr
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return githubRelease{}, err
	}
	if release.TagName == "" {
		return githubRelease{}, fmt.Errorf("latest release has no tag")
	}
	return release, nil
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

func markChecked() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := home + "/" + checkFile
	return os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0o600)
}

type semanticVersion struct {
	core       [3]string
	prerelease []string
}

func isNewer(current, latest string) bool {
	currentVersion, currentOK := parseSemanticVersion(current)
	latestVersion, latestOK := parseSemanticVersion(latest)
	if !currentOK || !latestOK {
		return false
	}

	for i := 0; i < 3; i++ {
		comparison := compareNumericIdentifier(latestVersion.core[i], currentVersion.core[i])
		if comparison > 0 {
			return true
		}
		if comparison < 0 {
			return false
		}
	}
	return comparePrerelease(latestVersion.prerelease, currentVersion.prerelease) > 0
}

func parseSemanticVersion(raw string) (semanticVersion, bool) {
	value := strings.TrimPrefix(raw, "v")
	if value == "" {
		return semanticVersion{}, false
	}

	main, build, hasBuild := strings.Cut(value, "+")
	if hasBuild && !validIdentifierList(build, false) {
		return semanticVersion{}, false
	}

	core, prerelease, hasPrerelease := strings.Cut(main, "-")
	coreParts := strings.Split(core, ".")
	if len(coreParts) != 3 {
		return semanticVersion{}, false
	}

	var parsed semanticVersion
	for i, part := range coreParts {
		if !validNumericIdentifier(part) {
			return semanticVersion{}, false
		}
		parsed.core[i] = part
	}

	if hasPrerelease {
		if !validIdentifierList(prerelease, true) {
			return semanticVersion{}, false
		}
		parsed.prerelease = strings.Split(prerelease, ".")
	}

	return parsed, true
}

func validIdentifierList(value string, rejectNumericLeadingZero bool) bool {
	parts := strings.Split(value, ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		numeric := true
		for i := 0; i < len(part); i++ {
			character := part[i]
			if character < '0' || character > '9' {
				numeric = false
			}
			if (character < '0' || character > '9') &&
				(character < 'A' || character > 'Z') &&
				(character < 'a' || character > 'z') &&
				character != '-' {
				return false
			}
		}
		if rejectNumericLeadingZero && numeric && len(part) > 1 && part[0] == '0' {
			return false
		}
	}
	return true
}

func validNumericIdentifier(value string) bool {
	if value == "" || len(value) > 1 && value[0] == '0' {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func compareNumericIdentifier(left, right string) int {
	if len(left) != len(right) {
		if len(left) > len(right) {
			return 1
		}
		return -1
	}
	return strings.Compare(left, right)
}

func comparePrerelease(left, right []string) int {
	if len(left) == 0 || len(right) == 0 {
		switch {
		case len(left) == 0 && len(right) == 0:
			return 0
		case len(left) == 0:
			return 1
		default:
			return -1
		}
	}

	sharedLength := min(len(left), len(right))
	for i := 0; i < sharedLength; i++ {
		leftNumeric := validNumericIdentifier(left[i])
		rightNumeric := validNumericIdentifier(right[i])
		switch {
		case leftNumeric && rightNumeric:
			if comparison := compareNumericIdentifier(left[i], right[i]); comparison != 0 {
				return comparison
			}
		case leftNumeric:
			return -1
		case rightNumeric:
			return 1
		default:
			if comparison := strings.Compare(left[i], right[i]); comparison != 0 {
				return comparison
			}
		}
	}

	switch {
	case len(left) > len(right):
		return 1
	case len(left) < len(right):
		return -1
	default:
		return 0
	}
}

func downloadAndReplace(assetURL, digest string) (err error) {
	expectedDigest, err := parseSHA256Digest(digest)
	if err != nil {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(exePath), ".rfw-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if tmp != nil {
			if closeErr := tmp.Close(); err == nil {
				err = closeErr
			}
		}
		if removeErr := os.Remove(tmpName); err == nil && removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = removeErr
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := safehttp.NewRequest(ctx, http.MethodGet, assetURL)
	if err != nil {
		return err
	}

	resp, err := safehttp.NewClient().Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return closeErr
		}
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	copyErr := copyWithSHA256(tmp, resp.Body, expectedDigest)
	closeErr := resp.Body.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := tmp.Chmod(0o700); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	tmp = nil

	if err := replaceExecutable(tmpName, exePath); err != nil {
		return err
	}

	return nil
}

func parseSHA256Digest(digest string) ([sha256.Size]byte, error) {
	var expected [sha256.Size]byte
	encoded, found := strings.CutPrefix(digest, "sha256:")
	if !found {
		return expected, fmt.Errorf("release asset has no SHA-256 digest")
	}
	decoded, err := hex.DecodeString(encoded)
	if err != nil || len(decoded) != sha256.Size {
		return expected, fmt.Errorf("release asset has an invalid SHA-256 digest")
	}
	copy(expected[:], decoded)
	return expected, nil
}

func copyWithSHA256(destination io.Writer, source io.Reader, expected [sha256.Size]byte) error {
	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(destination, hasher), source); err != nil {
		return err
	}
	if subtle.ConstantTimeCompare(hasher.Sum(nil), expected[:]) != 1 {
		return fmt.Errorf("downloaded asset checksum does not match the release")
	}
	return nil
}

func replaceExecutable(source, target string) error {
	if runtime.GOOS != "windows" {
		return os.Rename(source, target)
	}

	backup := target + ".old"
	if err := os.Remove(backup); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(target, backup); err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		if rollbackErr := os.Rename(backup, target); rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}
	return nil
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

// shouldSkipUpdateCheck reports whether the update check must not run at all:
// explicit opt-out via RFW_NO_UPDATE_CHECK, or a non-interactive session (CI,
// scripts, pipes) where a network call and an update prompt would get in the
// way.
func shouldSkipUpdateCheck(noUpdateEnv string, stdinTTY, stdoutTTY bool) bool {
	if noUpdateEnv != "" {
		return true
	}
	return !stdinTTY || !stdoutTTY
}

// isTerminal reports whether f is attached to a character device.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// CheckForUpdate checks GitHub for a newer CLI release in interactive sessions.
func CheckForUpdate() {
	if shouldSkipUpdateCheck(os.Getenv("RFW_NO_UPDATE_CHECK"), isTerminal(os.Stdin), isTerminal(os.Stdout)) {
		return
	}
	if !shouldCheckUpdate() {
		return
	}

	release, err := fetchLatestRelease()
	if err != nil {
		return
	}
	if err := markChecked(); err != nil {
		Debug(fmt.Sprintf("failed to record update check: %v", err))
	}
	latest := release.TagName

	if !isNewer(core.Version(), release.TagName) {
		return
	}

	assetName := getAssetName()
	fmt.Println()
	Info(fmt.Sprintf("Update available: %s → %s", faint(core.Version()), boldCyan(latest)))

	fmt.Print(indent, red("➜ "), bold("Update now? [y/N] "))
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println(indent, faint("Skipped."))
		return
	}

	assetURL := ""
	assetDigest := ""
	for _, a := range release.Assets {
		if a.Name == assetName {
			assetURL = a.BrowserDownloadURL
			assetDigest = a.Digest
			break
		}
	}
	if assetURL == "" {
		Info(fmt.Sprintf("No binary found for %s/%s", runtime.GOOS, runtime.GOARCH))
		return
	}

	Info("Downloading...")
	if err := downloadAndReplace(assetURL, assetDigest); err != nil {
		Info(fmt.Sprintf("Update failed: %v", err))
		return
	}

	Info(fmt.Sprintf("Updated to %s!", latest))
}
