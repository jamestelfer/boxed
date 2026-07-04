package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestRenderParity(t *testing.T) {
	assert.Equal(t, "\033[32m📦 sandboxed\033[0m", render(stateOn))
	assert.Equal(t, "\033[33m😬 sandbox (escape allowed)\033[0m", render(statePartial))
	assert.Equal(t, "\033[1;31m☢️ NOT sandboxed\033[0m", render(stateOff))
}
