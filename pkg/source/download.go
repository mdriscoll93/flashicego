package source

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func DownloadFile(uri, outputPath string, progress io.Writer) (int64, error) {
	client := &http.Client{Timeout: 0}
	resp, err := client.Get(uri)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("download failed with status %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, err
	}

	tmpPath := outputPath + ".partial"
	f, err := os.Create(tmpPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	reader := io.Reader(resp.Body)
	if progress != nil {
		reader = io.TeeReader(resp.Body, progress)
	}

	n, err := io.Copy(f, reader)
	if err != nil {
		return 0, err
	}
	if err := f.Sync(); err != nil {
		return 0, err
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return 0, err
	}
	_ = os.Chtimes(outputPath, time.Now(), time.Now())
	return n, nil
}

func DownloadBytes(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
