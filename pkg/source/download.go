package source

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultDownloadRetries = 3
	defaultRetryDelay      = 750 * time.Millisecond
)

func DownloadFile(uri, outputPath string, progress io.Writer) (int64, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	info, _ := remoteInfo(client, uri)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, err
	}

	tmpPath := outputPath + ".partial"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	for attempt := 0; attempt <= defaultDownloadRetries; attempt++ {
		startOffset, err := fileSize(f)
		if err != nil {
			return 0, err
		}
		if info.length > 0 && startOffset == info.length {
			if err := finalizeDownload(tmpPath, outputPath); err != nil {
				return 0, err
			}
			return startOffset, nil
		}

		resp, reqStart, err := getWithResume(client, uri, info, startOffset)
		if err != nil {
			if attempt == defaultDownloadRetries {
				return 0, err
			}
			time.Sleep(defaultRetryDelay * time.Duration(attempt+1))
			continue
		}

		copyErr := streamResponseToFile(f, resp, reqStart, progress)
		_ = resp.Body.Close()
		if copyErr != nil {
			if attempt == defaultDownloadRetries {
				return 0, copyErr
			}
			time.Sleep(defaultRetryDelay * time.Duration(attempt+1))
			continue
		}

		if err := f.Sync(); err != nil {
			if attempt == defaultDownloadRetries {
				return 0, err
			}
			time.Sleep(defaultRetryDelay * time.Duration(attempt+1))
			continue
		}

		finalSize, err := fileSize(f)
		if err != nil {
			return 0, err
		}
		if info.length > 0 && finalSize < info.length {
			if attempt == defaultDownloadRetries {
				return 0, fmt.Errorf("incomplete download: got %d of %d bytes", finalSize, info.length)
			}
			time.Sleep(defaultRetryDelay * time.Duration(attempt+1))
			continue
		}

		if err := finalizeDownload(tmpPath, outputPath); err != nil {
			return 0, err
		}
		_ = os.Chtimes(outputPath, time.Now(), time.Now())
		return finalSize, nil
	}

	return 0, errors.New("download attempts exhausted")
}

func DownloadBytes(uri string) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	var lastErr error
	for attempt := 0; attempt <= defaultDownloadRetries; attempt++ {
		resp, err := client.Get(uri)
		if err != nil {
			lastErr = err
		} else {
			b, readErr := func() ([]byte, error) {
				defer resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					return io.ReadAll(resp.Body)
				}
				if resp.StatusCode >= 500 {
					return nil, fmt.Errorf("request failed with status %s", resp.Status)
				}
				return nil, fmt.Errorf("request failed with status %s", resp.Status)
			}()
			if readErr == nil {
				return b, nil
			}
			lastErr = readErr
		}
		if attempt < defaultDownloadRetries {
			time.Sleep(defaultRetryDelay * time.Duration(attempt+1))
		}
	}
	return nil, lastErr
}

type remoteDownloadInfo struct {
	length      int64
	acceptRange bool
}

func remoteInfo(client *http.Client, uri string) (remoteDownloadInfo, error) {
	req, err := http.NewRequest(http.MethodHead, uri, nil)
	if err != nil {
		return remoteDownloadInfo{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return remoteDownloadInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return remoteDownloadInfo{}, fmt.Errorf("head request failed with status %s", resp.Status)
	}

	return remoteDownloadInfo{
		length:      resp.ContentLength,
		acceptRange: strings.Contains(strings.ToLower(resp.Header.Get("Accept-Ranges")), "bytes"),
	}, nil
}

func getWithResume(client *http.Client, uri string, info remoteDownloadInfo, startOffset int64) (*http.Response, int64, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, 0, err
	}
	if startOffset > 0 && info.acceptRange {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startOffset))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if startOffset > 0 && info.acceptRange && resp.StatusCode != http.StatusPartialContent {
			_ = resp.Body.Close()
			return nil, 0, fmt.Errorf("server ignored range request")
		}
		if resp.StatusCode == http.StatusOK {
			return resp, 0, nil
		}
		return resp, startOffset, nil
	}

	if resp.StatusCode >= 500 {
		_ = resp.Body.Close()
		return nil, 0, fmt.Errorf("server error %s", resp.Status)
	}

	_ = resp.Body.Close()
	return nil, 0, fmt.Errorf("download failed with status %s", resp.Status)
}

func streamResponseToFile(f *os.File, resp *http.Response, startOffset int64, progress io.Writer) error {
	if _, err := f.Seek(startOffset, 0); err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK && startOffset == 0 {
		if err := f.Truncate(0); err != nil {
			return err
		}
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}
	}

	reader := io.Reader(resp.Body)
	if progress != nil {
		reader = io.TeeReader(resp.Body, progress)
	}
	_, err := io.Copy(f, reader)
	return err
}

func fileSize(f *os.File) (int64, error) {
	st, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

func finalizeDownload(tmpPath, outputPath string) error {
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return err
	}
	return nil
}
