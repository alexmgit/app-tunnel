package server

import "testing"

func TestExtractSubdomain(t *testing.T) {
	domain := "example.com"

	cases := []struct {
		host     string
		expected string
		ok       bool
	}{
		{"abc.example.com", "abc", true},
		{"abc.example.com:8080", "abc", true},
		{"example.com", "", false},
		{"foo.bar.example.com", "", false},
		{"notexample.com", "", false},
		{"abc.other.com", "", false},
	}

	for _, tc := range cases {
		got, ok := extractSubdomain(tc.host, domain)
		if ok != tc.ok {
			t.Fatalf("host %q expected ok=%v got=%v", tc.host, tc.ok, ok)
		}
		if got != tc.expected {
			t.Fatalf("host %q expected subdomain=%q got=%q", tc.host, tc.expected, got)
		}
	}
}

func TestIsAllowedCaddyHost(t *testing.T) {
	domain := "example.com"

	cases := []struct {
		host string
		ok   bool
	}{
		{"example.com", true},
		{"control.example.com", true},
		{"abc.example.com", true},
		{"foo.bar.example.com", false},
		{"example.org", false},
		{"", false},
	}

	for _, tc := range cases {
		got := isAllowedCaddyHost(tc.host, domain)
		if got != tc.ok {
			t.Fatalf("host %q expected ok=%v got=%v", tc.host, tc.ok, got)
		}
	}
}
