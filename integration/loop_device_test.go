//go:build linux

package integration

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"flashicego/pkg/verify"
	writerpkg "flashicego/pkg/write"
)

func TestLoopDeviceWriteAndVerifyQuick(t *testing.T) {
	requireRootAndTools(t)

	tmp := t.TempDir()
	src := tmp + "/source.iso"
	backing := tmp + "/target.img"

	payload := make([]byte, 8*1024*1024)
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("rand payload: %v", err)
	}
	if err := os.WriteFile(src, payload, 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := truncate(backing, 32*1024*1024); err != nil {
		t.Fatalf("truncate backing: %v", err)
	}

	loopDev := setupLoop(t, backing)

	written, err := writerpkg.ImageToDevice(src, loopDev, nil)
	if err != nil {
		t.Fatalf("write image: %v", err)
	}
	if written != int64(len(payload)) {
		t.Fatalf("written bytes mismatch: got %d want %d", written, len(payload))
	}

	profile, err := verify.ResolveProfile("quick")
	if err != nil {
		t.Fatalf("resolve profile: %v", err)
	}
	if err := verify.CompareImageAndDevice(src, loopDev, profile, nil); err != nil {
		t.Fatalf("quick verify failed: %v", err)
	}
}

func TestLoopDeviceWriteAndVerifyFull(t *testing.T) {
	requireRootAndTools(t)

	tmp := t.TempDir()
	src := tmp + "/source.iso"
	backing := tmp + "/target.img"

	payload := bytes.Repeat([]byte("flashicego"), 1024*256)
	if err := os.WriteFile(src, payload, 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := truncate(backing, int64(len(payload))*2); err != nil {
		t.Fatalf("truncate backing: %v", err)
	}

	loopDev := setupLoop(t, backing)

	if _, err := writerpkg.ImageToDevice(src, loopDev, nil); err != nil {
		t.Fatalf("write image: %v", err)
	}
	profile, err := verify.ResolveProfile("full")
	if err != nil {
		t.Fatalf("resolve profile: %v", err)
	}
	if err := verify.CompareImageAndDevice(src, loopDev, profile, nil); err != nil {
		t.Fatalf("full verify failed: %v", err)
	}
}

func requireRootAndTools(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("requires root to configure loop devices")
	}
	for _, bin := range []string{"losetup", "truncate"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("requires %s in PATH", bin)
		}
	}
}

func truncate(path string, size int64) error {
	cmd := exec.Command("truncate", "-s", fmt.Sprintf("%d", size), path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("truncate: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func setupLoop(t *testing.T, backingFile string) string {
	t.Helper()
	cmd := exec.Command("losetup", "--find", "--show", backingFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("losetup --find --show: %v (%s)", err, strings.TrimSpace(string(out)))
	}
	loopDev := strings.TrimSpace(string(out))
	if loopDev == "" {
		t.Fatalf("losetup returned empty loop device")
	}
	t.Cleanup(func() {
		detach := exec.Command("losetup", "-d", loopDev)
		if out, err := detach.CombinedOutput(); err != nil {
			t.Logf("failed to detach %s: %v (%s)", loopDev, err, strings.TrimSpace(string(out)))
		}
	})
	return loopDev
}
