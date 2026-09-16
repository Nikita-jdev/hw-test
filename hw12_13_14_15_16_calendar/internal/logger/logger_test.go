package logger

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	_ = r.Close()
	return string(out)
}

func TestLoggerFiltersByLevel(t *testing.T) {
	t.Run("error level prints only errors", func(t *testing.T) {
		l := New("error")

		if got := captureOutput(func() { l.Info("info msg") }); strings.Contains(got, "info msg") {
			t.Fatalf("info must not be logged at error level, got: %q", got)
		}
		if got := captureOutput(func() { l.Debug("debug msg") }); strings.Contains(got, "debug msg") {
			t.Fatalf("debug must not be logged at error level, got: %q", got)
		}
		if got := captureOutput(func() { l.Error("error msg") }); !strings.Contains(got, "error msg") {
			t.Fatalf("error must be logged at error level, got: %q", got)
		}
	})

	t.Run("info level prints info and above", func(t *testing.T) {
		l := New("info")

		if got := captureOutput(func() { l.Debug("debug msg") }); strings.Contains(got, "debug msg") {
			t.Fatalf("debug must not be logged at info level, got: %q", got)
		}
		if got := captureOutput(func() { l.Info("info msg") }); !strings.Contains(got, "info msg") {
			t.Fatalf("info must be logged at info level, got: %q", got)
		}
		if got := captureOutput(func() { l.Warn("warn msg") }); !strings.Contains(got, "warn msg") {
			t.Fatalf("warn must be logged at info level, got: %q", got)
		}
		if got := captureOutput(func() { l.Error("error msg") }); !strings.Contains(got, "error msg") {
			t.Fatalf("error must be logged at info level, got: %q", got)
		}
	})
}
