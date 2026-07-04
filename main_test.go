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

func TestRenderParity(t *testing.T) {
	assert.Equal(t, "\033[32m📦 sandboxed\033[0m", render(stateOn))
	assert.Equal(t, "\033[33m😬 sandbox (escape allowed)\033[0m", render(statePartial))
	assert.Equal(t, "\033[1;31m☢️ NOT sandboxed\033[0m", render(stateOff))
}
