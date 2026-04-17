package cleanup

import (
	"crypto/rand"
	"fmt"
	"os"
)

func ShredAndDelete(path string, passes int) error {
	if passes < 1 {
		passes = 1
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	st, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}
	size := st.Size()
	buf := make([]byte, 1024*1024)

	for pass := 0; pass < passes; pass++ {
		if _, err := f.Seek(0, 0); err != nil {
			_ = f.Close()
			return err
		}
		remaining := size
		for remaining > 0 {
			chunk := int64(len(buf))
			if remaining < chunk {
				chunk = remaining
			}
			if pass == 0 {
				for i := int64(0); i < chunk; i++ {
					buf[i] = 0x00
				}
			} else if pass == 1 {
				for i := int64(0); i < chunk; i++ {
					buf[i] = 0xFF
				}
			} else {
				if _, err := rand.Read(buf[:chunk]); err != nil {
					_ = f.Close()
					return fmt.Errorf("random overwrite: %w", err)
				}
			}
			if _, err := f.Write(buf[:chunk]); err != nil {
				_ = f.Close()
				return err
			}
			remaining -= chunk
		}
		if err := f.Sync(); err != nil {
			_ = f.Close()
			return err
		}
	}

	if err := f.Close(); err != nil {
		return err
	}
	return os.Remove(path)
}
