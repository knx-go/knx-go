package util

import (
	"fmt"
	"strings"
	"testing"
)

type recordingLogger struct {
	messages []string
}

func (r *recordingLogger) Printf(format string, args ...interface{}) {
	r.messages = append(r.messages, fmt.Sprintf(format, args...))
}

type sampleComponent struct{}

func TestLogWritesMessage(t *testing.T) {
	logger := &recordingLogger{}
	Logger = logger
	longestLogger = 10
	t.Cleanup(func() {
		Logger = nil
		longestLogger = 10
	})

	component := &sampleComponent{}
	Log(component, "status %s", "ok")
	if len(logger.messages) != 1 {
		t.Fatalf("expected 1 log message, got %d", len(logger.messages))
	}
	msg := logger.messages[0]
	if !strings.Contains(msg, "status ok") {
		t.Fatalf("log message %q missing formatted content", msg)
	}
	if !strings.Contains(msg, "*util.sampleComponent") {
		t.Fatalf("log message %q missing type information", msg)
	}
}

func TestLogWithoutLogger(t *testing.T) {
	Logger = nil
	longestLogger = 10
	Log(&sampleComponent{}, "noop")
}
