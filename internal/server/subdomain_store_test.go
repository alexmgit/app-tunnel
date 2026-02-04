package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSubdomainStoreReuse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdomains.txt")

	store, err := NewSubdomainStore(path, 6)
	if err != nil {
		t.Fatalf("store create: %v", err)
	}

	first, err := store.Register("")
	if err != nil {
		t.Fatalf("register first: %v", err)
	}

	store.Release(first)

	second, err := store.Register("")
	if err != nil {
		t.Fatalf("register second: %v", err)
	}

	if second != first {
		t.Fatalf("expected reuse of %q got %q", first, second)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected store file to be written")
	}
}

func TestSubdomainStoreRequested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdomains.txt")

	store, err := NewSubdomainStore(path, 6)
	if err != nil {
		t.Fatalf("store create: %v", err)
	}

	name := "custom"
	got, err := store.Register(name)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if got != name {
		t.Fatalf("expected %q got %q", name, got)
	}

	if _, err := store.Register(name); err == nil {
		t.Fatalf("expected conflict on active subdomain")
	}
}
