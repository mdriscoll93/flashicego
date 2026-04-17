package source

import "testing"

func TestIsURL(t *testing.T) {
	if !IsURL("https://example.com/test.iso") {
		t.Fatal("expected URL to be detected")
	}
	if IsURL("/tmp/test.iso") {
		t.Fatal("did not expect local path to be URL")
	}
}

func TestSidecarChecksumURL(t *testing.T) {
	got := SidecarChecksumURL("https://example.com/test.iso")
	if got != "https://example.com/test.iso.sha256" {
		t.Fatalf("unexpected sidecar url: %s", got)
	}
}
