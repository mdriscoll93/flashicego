package device

import "testing"

func TestNormalizeIdentity(t *testing.T) {
	in := "  USB-Serial_12:ab  "
	got := normalizeIdentity(in)
	if got != "USBSERIAL12AB" {
		t.Fatalf("unexpected normalized identity: %s", got)
	}
}

func TestEnsureIdentityNormalizedMatch(t *testing.T) {
	d := Disk{Model: "SanDisk Ultra USB", Serial: "1234-abcd"}
	if err := EnsureIdentity(d, "sandisk ultra usb", "1234ABCD"); err != nil {
		t.Fatalf("expected identity match: %v", err)
	}
}

func TestEnsureIdentityMismatch(t *testing.T) {
	d := Disk{Model: "SanDisk Ultra USB", Serial: "1234-abcd"}
	if err := EnsureIdentity(d, "Kingston", "1234ABCD"); err == nil {
		t.Fatal("expected mismatch error")
	}
}
