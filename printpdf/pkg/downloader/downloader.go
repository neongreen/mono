package downloader

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolDownloader manages downloading and caching of external tools
type ToolDownloader struct {
	cacheDir string
}

// New creates a new ToolDownloader with the specified cache directory
func New() (*ToolDownloader, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cache", "printpdf")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ToolDownloader{cacheDir: cacheDir}, nil
}

// GetTool returns the path to a tool, downloading it if necessary
func (td *ToolDownloader) GetTool(name, version, downloadURL string) (string, error) {
	toolDir := filepath.Join(td.cacheDir, name, version)
	
	// Check if already downloaded
	if _, err := os.Stat(toolDir); err == nil {
		// Tool already exists, find the executable
		return td.findExecutable(toolDir, name)
	}

	// Download and extract
	fmt.Printf("Downloading %s %s...\n", name, version)
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tool directory: %w", err)
	}

	// Download to temp file
	tempFile := filepath.Join(toolDir, "download.tmp")
	if err := td.download(downloadURL, tempFile); err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(tempFile)

	// Extract based on file extension
	if strings.HasSuffix(downloadURL, ".tar.gz") || strings.HasSuffix(downloadURL, ".tgz") {
		if err := td.extractTarGz(tempFile, toolDir); err != nil {
			return "", fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	} else if strings.HasSuffix(downloadURL, ".tar.xz") {
		// tar.xz not supported yet, inform user
		return "", fmt.Errorf("tar.xz extraction not yet supported. Please install %s manually", name)
	} else if strings.HasSuffix(downloadURL, ".zip") {
		if err := td.extractZip(tempFile, toolDir); err != nil {
			return "", fmt.Errorf("failed to extract zip: %w", err)
		}
	} else {
		// Assume it's a single binary
		binPath := filepath.Join(toolDir, name)
		if err := os.Rename(tempFile, binPath); err != nil {
			return "", fmt.Errorf("failed to move binary: %w", err)
		}
		if err := os.Chmod(binPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make executable: %w", err)
		}
		return binPath, nil
	}

	return td.findExecutable(toolDir, name)
}

func (td *ToolDownloader) download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (td *ToolDownloader) extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

func (td *ToolDownloader) extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (td *ToolDownloader) findExecutable(dir, name string) (string, error) {
	// Common executable patterns
	patterns := []string{
		filepath.Join(dir, name),
		filepath.Join(dir, "bin", name),
		filepath.Join(dir, "*", name),
		filepath.Join(dir, "*", "bin", name),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && !info.IsDir() {
				// Make sure it's executable
				if err := os.Chmod(match, 0755); err == nil {
					return match, nil
				}
			}
		}
	}

	return "", fmt.Errorf("executable not found in %s", dir)
}

// GetPlatform returns the current platform in the format "os-arch"
func GetPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}
