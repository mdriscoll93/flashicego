package verify

import "testing"

func TestResolveProfile(t *testing.T) {
	p, err := ResolveProfile("quick")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Full {
		t.Fatal("quick should not be full")
	}

	p, err = ResolveProfile("full")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.Full {
		t.Fatal("full should be full compare")
	}

	if _, err := ResolveProfile("nope"); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}
