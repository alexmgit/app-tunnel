package client

import "testing"

func TestBuildPublicURL(t *testing.T) {
	if got := buildPublicURL("abc123", "example.com"); got != "https://abc123.example.com" {
		t.Fatalf("unexpected url: %s", got)
	}
}

func TestBuildPublicURLEmpty(t *testing.T) {
	if got := buildPublicURL("", "example.com"); got != "" {
		t.Fatalf("expected empty url, got: %s", got)
	}
	if got := buildPublicURL("abc123", ""); got != "" {
		t.Fatalf("expected empty url, got: %s", got)
	}
}
