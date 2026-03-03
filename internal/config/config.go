package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	ErrMissingEnv = errors.New("missing required environment variable")
)

func RequireString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingEnv, key)
	}
	return value, nil
}

func RequireInt(key string) (int, error) {
	value, err := RequireString(key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid int for %s: %w", key, err)
	}
	return parsed, nil
}

func RequireDuration(key string) (time.Duration, error) {
	value, err := RequireString(key)
	if err != nil {
		return 0, err
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	return parsed, nil
}

func OptionalString(key string) string {
	return os.Getenv(key)
}

func OptionalInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid int for %s: %w", key, err)
	}
	return parsed, nil
}

func OptionalDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	return parsed, nil
}
