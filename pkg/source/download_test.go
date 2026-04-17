package source

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDownloadFileResumesPartial(t *testing.T) {
	data := bytes.Repeat([]byte("abc12345"), 128*1024)
	var sawRange atomic.Bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path != "/image.iso" {
			http.NotFound(w, r)
			return
		}

		rangeHeader := r.Header.Get("Range")
		if rangeHeader == "" {
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
		sawRange.Store(true)
		startStr := strings.TrimPrefix(rangeHeader, "bytes=")
		startStr = strings.TrimSuffix(startStr, "-")
		start, err := strconv.Atoi(startStr)
		if err != nil || start < 0 || start >= len(data) {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(data)-1, len(data)))
		w.Header().Set("Content-Length", strconv.Itoa(len(data)-start))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(data[start:])
	}))
	defer srv.Close()

	tmp := t.TempDir()
	out := filepath.Join(tmp, "image.iso")
	partial := out + ".partial"
	seed := data[:len(data)/3]
	if err := os.WriteFile(partial, seed, 0o644); err != nil {
		t.Fatalf("seed partial: %v", err)
	}

	n, err := DownloadFile(srv.URL+"/image.iso", out, io.Discard)
	if err != nil {
		t.Fatalf("download resume: %v", err)
	}
	if n != int64(len(data)) {
		t.Fatalf("size mismatch: got %d want %d", n, len(data))
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("downloaded content mismatch")
	}
	if !sawRange.Load() {
		t.Fatal("expected ranged GET for resumed download")
	}
}

func TestDownloadFileRetriesOnServerErrors(t *testing.T) {
	data := bytes.Repeat([]byte("retry-me"), 16*1024)
	var getCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			call := getCalls.Add(1)
			if call == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("temporary failure"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	out := filepath.Join(tmp, "image.iso")
	n, err := DownloadFile(srv.URL+"/image.iso", out, io.Discard)
	if err != nil {
		t.Fatalf("download retry: %v", err)
	}
	if n != int64(len(data)) {
		t.Fatalf("size mismatch: got %d want %d", n, len(data))
	}
	if getCalls.Load() < 2 {
		t.Fatalf("expected at least 2 GET attempts, got %d", getCalls.Load())
	}
}

func TestDownloadBytesRetries(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	b, err := DownloadBytes(srv.URL)
	if err != nil {
		t.Fatalf("download bytes retry: %v", err)
	}
	if string(b) != "ok" {
		t.Fatalf("unexpected response: %q", string(b))
	}
}
