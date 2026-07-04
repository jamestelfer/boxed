package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout runs f with os.Stdout redirected to a buffer and returns what
// was written.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()
	f()
	_ = w.Close()
	os.Stdout = orig
	return <-done
}

// `boxed state` prints exactly the resolved token plus a trailing newline, with
// no ANSI escapes.
func TestStateCommandOutput(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(dir+"/.claude", 0o755))
	require.NoError(t, os.WriteFile(dir+"/.claude/settings.json",
		[]byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`), 0o644))
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Setenv("HOME", "/nonexistent")

	out := captureStdout(t, func() {
		require.NoError(t, newCommand().Run(context.Background(), []string{"boxed", "state"}))
	})
	assert.Equal(t, "on\n", out)
	assert.NotContains(t, out, "\033")
	assert.False(t, strings.ContainsRune(out, '\x1b'))
}
