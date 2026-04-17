package device

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Disk struct {
	Name      string
	Path      string
	SizeBytes int64
	Model     string
	Serial    string
	ByIDPaths []string
	Removable bool
	Mounted   bool
}

var nonAlphaNum = regexp.MustCompile(`[^A-Za-z0-9]+`)

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
	resolvedPath := path
	if rp, err := filepath.EvalSymlinks(path); err == nil {
		resolvedPath = rp
	}
	disks, err := ListRemovableDisks()
	if err != nil {
		return Disk{}, err
	}
	for _, d := range disks {
		if d.Path == path || d.Path == resolvedPath || containsString(d.ByIDPaths, path) || containsString(d.ByIDPaths, resolvedPath) {
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
	if expectedModel != "" && normalizeIdentity(d.Model) != normalizeIdentity(expectedModel) {
		return fmt.Errorf("model mismatch: expected %q got %q", expectedModel, d.Model)
	}
	if expectedSerial != "" && normalizeIdentity(d.Serial) != normalizeIdentity(expectedSerial) {
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
	byIDs := byIDPathsFor(name)
	if serial == "" {
		serial = serialFromByID(byIDs)
	}

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
		ByIDPaths: byIDs,
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

func normalizeIdentity(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ToUpper(v)
	v = nonAlphaNum.ReplaceAllString(v, "")
	return v
}

func byIDPathsFor(deviceName string) []string {
	dir := "/dev/disk/by-id"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	devicePath := filepath.Join("/dev", deviceName)
	var out []string
	for _, e := range entries {
		if strings.Contains(e.Name(), "-part") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		rp, err := filepath.EvalSymlinks(full)
		if err != nil {
			continue
		}
		if rp == devicePath {
			out = append(out, full)
		}
	}
	sort.Strings(out)
	return out
}

func serialFromByID(ids []string) string {
	for _, idPath := range ids {
		base := filepath.Base(idPath)
		parts := strings.Split(base, "_")
		if len(parts) == 0 {
			continue
		}
		candidate := parts[len(parts)-1]
		candidate = strings.TrimSpace(candidate)
		if normalizeIdentity(candidate) != "" {
			return candidate
		}
	}
	return ""
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
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
