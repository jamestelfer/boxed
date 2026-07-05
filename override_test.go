package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixtureEnv points resolution at a temp project tree with the given
// settings.json body, and returns the project directory.
func fixtureEnv(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(dir+"/.claude", 0o755))
	require.NoError(t, os.WriteFile(dir+"/.claude/settings.json", []byte(body), 0o644))
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Setenv("HOME", "/nonexistent")
	return dir
}

// The flag matching the resolved state supplies the rendered output.
func TestOverrideMatchingState(t *testing.T) {
	fixtureEnv(t, `{"sandbox":{"enabled":true}}`) // partial
	out := captureStdout(t, func() {
		require.NoError(t, newCommand().Run(context.Background(),
			[]string{"boxed", "--partial", "[x](red)"}))
	})
	assert.Equal(t, "\x1b[31mx\x1b[m", out)
}

// A flag for a non-resolved state is ignored; the default renders.
func TestOverrideNonMatchingIgnored(t *testing.T) {
	fixtureEnv(t, `{"sandbox":{"enabled":true}}`) // partial
	out := captureStdout(t, func() {
		require.NoError(t, newCommand().Run(context.Background(),
			[]string{"boxed", "--on", "[nope](green)"}))
	})
	assert.Equal(t, "\x1b[33m😬 sandbox (escape allowed)\x1b[m", out)
}

// A malformed override exits non-zero (reuses the Phase 7 error path).
func TestOverrideMalformedErrors(t *testing.T) {
	fixtureEnv(t, `{"sandbox":{"enabled":false}}`) // off
	err := newCommand().Run(context.Background(),
		[]string{"boxed", "--off", "[broken"})
	require.Error(t, err)
}
