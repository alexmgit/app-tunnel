package server

import (
	"bufio"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

type SubdomainStore struct {
	path           string
	subdomainLen   int
	known          []string
	active         map[string]struct{}
	mu             sync.Mutex
}

func NewSubdomainStore(path string, subdomainLen int) (*SubdomainStore, error) {
	store := &SubdomainStore{
		path:         path,
		subdomainLen: subdomainLen,
		active:       make(map[string]struct{}),
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *SubdomainStore) load() error {
	file, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		s.known = append(s.known, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *SubdomainStore) Register(requested string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if requested != "" {
		if s.isActive(requested) {
			return "", fmt.Errorf("subdomain already active")
		}
		s.active[requested] = struct{}{}
		if !s.isKnown(requested) {
			if err := s.appendKnown(requested); err != nil {
				delete(s.active, requested)
				return "", err
			}
		}
		return requested, nil
	}

	for _, candidate := range s.known {
		if !s.isActive(candidate) {
			s.active[candidate] = struct{}{}
			return candidate, nil
		}
	}

	generated, err := s.generate()
	if err != nil {
		return "", err
	}
	s.active[generated] = struct{}{}
	if err := s.appendKnown(generated); err != nil {
		delete(s.active, generated)
		return "", err
	}
	return generated, nil
}

func (s *SubdomainStore) Release(subdomain string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.active, subdomain)
}

func (s *SubdomainStore) isActive(subdomain string) bool {
	_, ok := s.active[subdomain]
	return ok
}

func (s *SubdomainStore) isKnown(subdomain string) bool {
	for _, existing := range s.known {
		if existing == subdomain {
			return true
		}
	}
	return false
}

func (s *SubdomainStore) appendKnown(subdomain string) error {
	file, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(subdomain + "\n"); err != nil {
		return err
	}
	s.known = append(s.known, subdomain)
	return nil
}

func (s *SubdomainStore) generate() (string, error) {
	bytes := make([]byte, s.subdomainLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	encoded = strings.ToLower(encoded)
	if len(encoded) < s.subdomainLen {
		return "", fmt.Errorf("generated subdomain too short")
	}
	return encoded[:s.subdomainLen], nil
}
