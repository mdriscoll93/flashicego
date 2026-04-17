package checksum

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func SHA256File(path string, progress io.Writer) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	reader := io.Reader(f)
	if progress != nil {
		reader = io.TeeReader(f, progress)
	}
	if _, err := io.Copy(h, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ParseSHA256Sidecar(content string) (string, error) {
	s := bufio.NewScanner(strings.NewReader(content))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		h := strings.ToLower(fields[0])
		if len(h) != 64 {
			continue
		}
		return h, nil
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no sha256 checksum found in sidecar")
}

func VerifySHA256(path string, expected string, progress io.Writer) error {
	actual, err := SHA256File(path, progress)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, strings.TrimSpace(expected)) {
		return fmt.Errorf("checksum mismatch: expected %s got %s", expected, actual)
	}
	return nil
}
