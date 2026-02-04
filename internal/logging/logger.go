package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

type Logger struct {
	level Level
	log   *log.Logger
}

func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return Debug, nil
	case "info":
		return Info, nil
	case "warn", "warning":
		return Warn, nil
	case "error":
		return Error, nil
	case "":
		return Info, nil
	default:
		return Info, fmt.Errorf("unknown log level: %s", value)
	}
}

func NewLogger(prefix string, level Level) *Logger {
	logger := log.New(os.Stdout, prefix, log.LstdFlags|log.Lmicroseconds)
	return &Logger{level: level, log: logger}
}

func (l *Logger) Debugf(format string, args ...any) {
	if l.level <= Debug {
		l.log.Printf("DEBUG "+format, args...)
	}
}

func (l *Logger) Infof(format string, args ...any) {
	if l.level <= Info {
		l.log.Printf("INFO "+format, args...)
	}
}

func (l *Logger) Warnf(format string, args ...any) {
	if l.level <= Warn {
		l.log.Printf("WARN "+format, args...)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	if l.level <= Error {
		l.log.Printf("ERROR "+format, args...)
	}
}
