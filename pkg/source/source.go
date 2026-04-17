package source

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Resolved struct {
	Path          string
	FromURL       bool
	DownloadedURL string
	TempFile      bool
}

func IsURL(v string) bool {
	u, err := url.Parse(v)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func ResolveLocal(path string) (Resolved, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Resolved{}, err
	}
	if _, err := os.Stat(abs); err != nil {
		return Resolved{}, err
	}
	return Resolved{Path: abs}, nil
}

func SidecarChecksumURL(isoURL string) string {
	if strings.HasSuffix(isoURL, ".iso") {
		return isoURL + ".sha256"
	}
	return isoURL + ".sha256"
}
