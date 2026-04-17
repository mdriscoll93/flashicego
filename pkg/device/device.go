package device

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Disk struct {
	Name      string
	Path      string
	SizeBytes int64
	Model     string
	Serial    string
	Removable bool
	Mounted   bool
}

func ListRemovableDisks() ([]Disk, error) {
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("read /sys/block: %w", err)
	}

	var disks []Disk
	for _, e := range entries {
		name := e.Name()
		if isPseudoBlock(name) {
			continue
		}
		removable, err := readTrimmed(filepath.Join("/sys/block", name, "removable"))
		if err != nil || removable != "1" {
			continue
		}
		d, err := inspectDisk(name)
		if err != nil {
			continue
		}
		disks = append(disks, d)
	}
	return disks, nil
}

func FindDiskByPath(path string) (Disk, error) {
	disks, err := ListRemovableDisks()
	if err != nil {
		return Disk{}, err
	}
	for _, d := range disks {
		if d.Path == path {
			return d, nil
		}
	}
	return Disk{}, fmt.Errorf("disk %s not found among removable devices", path)
}

func EnsureUnmounted(d Disk) error {
	if !d.Mounted {
		return nil
	}
	return fmt.Errorf("target %s has mounted partitions; unmount before writing", d.Path)
}

func EnsureIdentity(d Disk, expectedModel, expectedSerial string) error {
	if expectedModel != "" && !strings.EqualFold(strings.TrimSpace(d.Model), strings.TrimSpace(expectedModel)) {
		return fmt.Errorf("model mismatch: expected %q got %q", expectedModel, d.Model)
	}
	if expectedSerial != "" && !strings.EqualFold(strings.TrimSpace(d.Serial), strings.TrimSpace(expectedSerial)) {
		return fmt.Errorf("serial mismatch: expected %q got %q", expectedSerial, d.Serial)
	}
	return nil
}

func HumanBytes(v int64) string {
	if v < 1024 {
		return fmt.Sprintf("%dB", v)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	fv := float64(v)
	idx := 0
	for fv >= 1024 && idx < len(units)-1 {
		fv /= 1024
		idx++
	}
	return fmt.Sprintf("%.2f%s", fv, units[idx])
}

func inspectDisk(name string) (Disk, error) {
	devPath := filepath.Join("/dev", name)
	if _, err := os.Stat(devPath); err != nil {
		return Disk{}, err
	}
	sectorsRaw, err := readTrimmed(filepath.Join("/sys/class/block", name, "size"))
	if err != nil {
		return Disk{}, err
	}
	sectors, err := strconv.ParseInt(sectorsRaw, 10, 64)
	if err != nil {
		return Disk{}, err
	}
	model, _ := readTrimmed(filepath.Join("/sys/block", name, "device/model"))
	serial, _ := readTrimmed(filepath.Join("/sys/block", name, "device/serial"))

	mounted, err := isMounted(name)
	if err != nil {
		return Disk{}, err
	}

	return Disk{
		Name:      name,
		Path:      devPath,
		SizeBytes: sectors * 512,
		Model:     model,
		Serial:    serial,
		Removable: true,
		Mounted:   mounted,
	}, nil
}

func isMounted(name string) (bool, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return false, err
	}
	defer f.Close()
	prefix := "/dev/" + name

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, prefix) {
			return true, nil
		}
	}
	if err := s.Err(); err != nil {
		return false, err
	}
	return false, nil
}

func readTrimmed(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func isPseudoBlock(name string) bool {
	prefixes := []string{"loop", "ram", "dm-", "md"}
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func RequireLinux() error {
	if _, err := os.Stat("/sys/block"); err != nil {
		return errors.New("linux block device interfaces are unavailable")
	}
	return nil
}
