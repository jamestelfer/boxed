package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDirFallback(t *testing.T) {
	// CLAUDE_PROJECT_DIR set → used verbatim.
	got := projectDir(
		func(string) string { return "/from/env" },
		func() (string, error) { return "/from/cwd", nil },
	)
	assert.Equal(t, "/from/env", got)

	// Unset → falls back to cwd.
	got = projectDir(
		func(string) string { return "" },
		func() (string, error) { return "/from/cwd", nil },
	)
	assert.Equal(t, "/from/cwd", got)
}

// buildVersion never returns empty, so urfave always wires -v/--version.
func TestBuildVersionNonEmpty(t *testing.T) {
	assert.NotEmpty(t, buildVersion())
	assert.NotEmpty(t, newCommand().Version)
}

// A valid no-subcommand invocation succeeds.
func TestRootCommandRuns(t *testing.T) {
	err := newCommand().Run(context.Background(), []string{"boxed"})
	assert.NoError(t, err)
}

// An unknown flag is rejected (urfave usage error).
func TestUnknownFlagRejected(t *testing.T) {
	err := newCommand().Run(context.Background(), []string{"boxed", "--bogus"})
	require.Error(t, err)
}

// An unexpected positional argument is rejected.
func TestUnexpectedArgumentRejected(t *testing.T) {
	err := newCommand().Run(context.Background(), []string{"boxed", "frobnicate"})
	require.Error(t, err)
}

// Golden: default output for each state. Visually identical to the historical
// hardcoded labels; the only byte difference from the pre-lipgloss binary is
// the reset sequence (\x1b[m vs \x1b[0m), which renders the same.
func TestDefaultRenderGolden(t *testing.T) {
	golden := map[state]string{
		stateOn:      "\x1b[32m📦 sandboxed\x1b[m",
		statePartial: "\x1b[33m😬 sandbox (escape allowed)\x1b[m",
		stateOff:     "\x1b[1;31m☢️ NOT sandboxed\x1b[m",
	}
	for s, want := range golden {
		out, err := render(s, "")
		require.NoError(t, err)
		assert.Equal(t, want, out, "state %s", s)
	}
}
