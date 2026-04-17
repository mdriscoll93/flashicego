package verify

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
)

func CompareImageAndDevice(imagePath, devicePath string, profile Profile, progress io.Writer) error {
	src, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Open(devicePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	st, err := src.Stat()
	if err != nil {
		return err
	}
	imageSize := st.Size()
	if imageSize <= 0 {
		return fmt.Errorf("image has invalid size")
	}

	if profile.Full {
		return compareRange(src, dst, 0, imageSize, progress)
	}

	edge := minInt64(profile.EdgeBytes, imageSize)
	if err := compareRange(src, dst, 0, edge, progress); err != nil {
		return fmt.Errorf("first edge mismatch: %w", err)
	}
	if err := compareRange(src, dst, imageSize-edge, imageSize, progress); err != nil {
		return fmt.Errorf("last edge mismatch: %w", err)
	}

	r := rand.New(rand.NewSource(imageSize + int64(profile.RandomSamples)))
	for i := 0; i < profile.RandomSamples; i++ {
		if imageSize <= edge {
			break
		}
		maxStart := imageSize - edge
		start := r.Int63n(maxStart)
		end := start + edge
		if end > imageSize {
			end = imageSize
		}
		if err := compareRange(src, dst, start, end, progress); err != nil {
			return fmt.Errorf("random sample %d mismatch at offset %d: %w", i+1, start, err)
		}
	}
	return nil
}

func compareRange(src, dst *os.File, start, end int64, progress io.Writer) error {
	if end <= start {
		return nil
	}
	const chunkSize = 4 * 1024 * 1024
	offset := start
	bufA := make([]byte, chunkSize)
	bufB := make([]byte, chunkSize)

	for offset < end {
		toRead := int64(chunkSize)
		if end-offset < toRead {
			toRead = end - offset
		}
		nA, err := src.ReadAt(bufA[:toRead], offset)
		if err != nil && err != io.EOF {
			return err
		}
		nB, err := dst.ReadAt(bufB[:toRead], offset)
		if err != nil && err != io.EOF {
			return err
		}
		if nA != nB {
			return fmt.Errorf("short read mismatch at %d", offset)
		}
		if !bytes.Equal(bufA[:nA], bufB[:nB]) {
			return fmt.Errorf("content mismatch at %d", offset)
		}
		offset += int64(nA)
		if progress != nil {
			_, _ = progress.Write(make([]byte, nA))
		}
	}
	return nil
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
