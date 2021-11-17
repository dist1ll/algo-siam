package siam

import (
	"io"
)

func NewLogWrapper(w io.Writer) *LogWrapper {
	return &LogWrapper{Writer: w, enabled: true}
}

// LogWrapper wraps any io.Writer so that the Write method can be toggled on or off.
type LogWrapper struct {
	Writer  io.Writer
	enabled bool
}

func (l *LogWrapper) Enable() {
	l.enabled = true
}

func (l *LogWrapper) Disable() {
	l.enabled = false
}

func (l *LogWrapper) Write(p []byte) (int, error) {
	if l.enabled {
		return l.Writer.Write(p)
	}
	return 0, nil
}
