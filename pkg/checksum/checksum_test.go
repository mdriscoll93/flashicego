package checksum

import "testing"

func TestParseSHA256Sidecar(t *testing.T) {
	in := "a3f5f1c88b8b4f6c0f35c5aa1178184f2f5c94761ad4e4f538454fc2ee5d1b08  sample.iso\n"
	got, err := ParseSHA256Sidecar(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "a3f5f1c88b8b4f6c0f35c5aa1178184f2f5c94761ad4e4f538454fc2ee5d1b08" {
		t.Fatalf("unexpected hash: %s", got)
	}
}

func TestParseSHA256SidecarMissing(t *testing.T) {
	_, err := ParseSHA256Sidecar("not-a-checksum")
	if err == nil {
		t.Fatal("expected error")
	}
}
