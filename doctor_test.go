package main

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// When the managed tier decides the outcome, determineState reports it as a
// single unit rather than attributing individual keys — the fail-safe
// combines whole states (see resolveStatus), not keys.
func TestDetermineStateManagedOrigin(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath():                  {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)},
		"proj/.claude/settings.local.json": {Data: []byte(`{"sandbox":{"enabled":false}}`)},
	}
	s, managed, keys := determineState(fsys, "/proj", "/home")
	assert.Equal(t, stateOn, s)
	assert.True(t, managed)
	assert.Nil(t, keys)
}

// `boxed doctor` reports the resolved state plus, per key, the specific
// settings file that supplied it.
func TestDoctorReportsNonManagedOrigin(t *testing.T) {
	dir := fixtureEnv(t, `{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)

	out := captureStdout(t, func() {
		require.NoError(t, newCommand().Run(context.Background(), []string{"boxed", "doctor"}))
	})
	assert.Equal(t, "state: on\n"+
		"sandbox.enabled: true ("+dir+"/.claude/settings.json)\n"+
		"sandbox.allowUnsandboxedCommands: false ("+dir+"/.claude/settings.json)\n", out)
}

// A key nobody set falls back to the schema default and is reported as unset.
func TestDoctorReportsUnsetKey(t *testing.T) {
	dir := fixtureEnv(t, `{"sandbox":{"enabled":true}}`)

	out := captureStdout(t, func() {
		require.NoError(t, newCommand().Run(context.Background(), []string{"boxed", "doctor"}))
	})
	assert.Equal(t, "state: partial\n"+
		"sandbox.enabled: true ("+dir+"/.claude/settings.json)\n"+
		"sandbox.allowUnsandboxedCommands: unset (default)\n", out)
}

// An unexpected positional argument is rejected, matching the other subcommands.
func TestDoctorUnexpectedArgumentRejected(t *testing.T) {
	err := newCommand().Run(context.Background(), []string{"boxed", "doctor", "frobnicate"})
	require.Error(t, err)
}
