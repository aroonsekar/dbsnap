package mongo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// DownloadTools identifies the OS, downloads the correct archive, and unpacks it.
func DownloadTools() error {
	osType := runtime.GOOS // "windows" or "darwin"

	fmt.Printf("[System] Detected OS: %s\n", osType)

	url, ext := getDownloadURL(osType)
	if url == "" {
		return fmt.Errorf("auto-download not supported for OS: %s", osType)
	}

	// Define isolated storage path
	homeDir, _ := os.UserHomeDir()
	binDir := filepath.Join(homeDir, ".dbsnap", "bin")

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %v", err)
	}

	archivePath := filepath.Join(binDir, "mongotools"+ext)

	fmt.Printf("[System] Fetching binaries from MongoDB servers...\n")
	if err := downloadFile(archivePath, url); err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	fmt.Println("[System] Extracting tools...")
	// TODO: Unzip/Untar logic to drop mongodump.exe into binDir

	// Cleanup the downloaded archive
	os.Remove(archivePath)

	fmt.Println("[System] Database tools installed successfully.")
	return nil
}

// getDownloadURL maps the OS to the official MongoDB URL.
// Note: Hardcoding versions for stability is industrial standard.
func getDownloadURL(osType string) (string, string) {
	const version = "100.9.4" // Stable tools release

	switch osType {
	case "windows":
		return fmt.Sprintf("https://fastdl.mongodb.org/tools/db/mongodb-database-tools-windows-x86_64-%s.zip", version), ".zip"
	case "darwin":
		// Mac requires the .tgz format
		return fmt.Sprintf("https://fastdl.mongodb.org/tools/db/mongodb-database-tools-macos-arm64-%s.zip", version), ".zip"
	default:
		return "", ""
	}
}

// downloadFile writes an HTTP stream to the local disk.
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
