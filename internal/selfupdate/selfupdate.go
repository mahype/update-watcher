package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mahype/update-watcher/internal/version"
)

const (
	repo       = "mahype/update-watcher"
	binaryName = "update-watcher"
	apiURL     = "https://api.github.com/repos/" + repo + "/releases/latest"
)

// Release holds information about a GitHub release.
type Release struct {
	TagName     string
	Version     string // TagName without "v" prefix
	DownloadURL string
}

// httpClient is used for all requests (longer timeout for downloads).
var httpClient = &http.Client{Timeout: 120 * time.Second}

// LatestRelease queries the GitHub API for the latest release and returns
// the matching download URL for the current platform.
func LatestRelease() (*Release, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "update-watcher/"+version.Version)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var data struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	archiveName := ArchiveName()
	var downloadURL string
	for _, asset := range data.Assets {
		if asset.Name == archiveName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("no matching asset %q found in release %s", archiveName, data.TagName)
	}

	ver := strings.TrimPrefix(data.TagName, "v")
	return &Release{
		TagName:     data.TagName,
		Version:     ver,
		DownloadURL: downloadURL,
	}, nil
}

// NeedsUpdate compares the current version with the latest release.
// Returns true if an update is available.
func NeedsUpdate(currentVersion string, latest *Release) bool {
	current := strings.TrimPrefix(currentVersion, "v")
	return current != latest.Version && currentVersion != "dev"
}

// DownloadAndReplace downloads the release archive, extracts the binary,
// and replaces the current executable.
func DownloadAndReplace(release *Release) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Check write permissions by creating a temp file in the same directory.
	execDir := filepath.Dir(execPath)
	tmpCheck, err := os.CreateTemp(execDir, ".update-watcher-check-*")
	if err != nil {
		return fmt.Errorf("no write permission to %s — try running with sudo", execDir)
	}
	tmpCheck.Close()
	os.Remove(tmpCheck.Name())

	// Download archive to temp directory.
	tmpDir, err := os.MkdirTemp("", "update-watcher-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "archive.tar.gz")
	if err := downloadFile(release.DownloadURL, archivePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Extract the binary from the archive.
	newBinaryPath := filepath.Join(tmpDir, binaryName)
	if err := extractBinary(archivePath, newBinaryPath); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Atomic replace: write temp file in target directory, then rename.
	tmpBinary, err := os.CreateTemp(execDir, ".update-watcher-new-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file for replacement: %w", err)
	}
	tmpBinaryPath := tmpBinary.Name()

	src, err := os.Open(newBinaryPath)
	if err != nil {
		os.Remove(tmpBinaryPath)
		return fmt.Errorf("failed to open new binary: %w", err)
	}

	if _, err := io.Copy(tmpBinary, src); err != nil {
		src.Close()
		tmpBinary.Close()
		os.Remove(tmpBinaryPath)
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	src.Close()
	tmpBinary.Close()

	if err := os.Chmod(tmpBinaryPath, 0755); err != nil {
		os.Remove(tmpBinaryPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tmpBinaryPath, execPath); err != nil {
		os.Remove(tmpBinaryPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// ArchiveName returns the expected archive filename for the current platform.
func ArchiveName() string {
	arch := runtime.GOARCH
	if arch == "arm" {
		arch = "armv7"
	}
	return fmt.Sprintf("%s_%s_%s.tar.gz", binaryName, runtime.GOOS, arch)
}

func downloadFile(url, dest string) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to open gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		if filepath.Base(header.Name) == binaryName && header.Typeflag == tar.TypeReg {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			return nil
		}
	}

	return fmt.Errorf("binary %q not found in archive", binaryName)
}
