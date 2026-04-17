package write

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

func ImageToDevice(imagePath, devicePath string, progress io.Writer) (int64, error) {
	src, err := os.Open(imagePath)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	reader := io.Reader(src)
	if progress != nil {
		reader = io.TeeReader(src, progress)
	}
	written, err := io.CopyBuffer(dst, reader, make([]byte, 4*1024*1024))
	if err != nil {
		return written, err
	}
	if err := dst.Sync(); err != nil {
		return written, fmt.Errorf("sync destination: %w", err)
	}
	syscall.Sync()
	return written, nil
}
