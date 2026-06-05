package main

import (
	"fmt"
	"io"
	"sync"
)

type VerboseLogger struct {
	enabled bool
	out     io.Writer
	mu      sync.Mutex
}

func newVerboseLogger(enabled bool, out io.Writer) *VerboseLogger {
	return &VerboseLogger{
		enabled: enabled,
		out:     out,
	}
}

func (l *VerboseLogger) Enabled() bool {
	return l != nil && l.enabled && l.out != nil
}

func (l *VerboseLogger) Logf(category string, format string, args ...any) {
	if !l.Enabled() {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintf(l.out, "[%s] %s\n", category, fmt.Sprintf(format, args...))
}
