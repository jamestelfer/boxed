package main

import (
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// managedFilePath is the OS-appropriate file-based managed-settings.json path
// as an fs.FS-relative key.
func managedFilePath() string {
	return fsPath(filepath.Join(managedDir(), "managed-settings.json"))
}

// File-based managed settings alone resolve correctly.
func TestFileBasedManagedOnly(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath(): {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)},
	}
	s := fileBasedManaged(fsys)
	require.NotNil(t, s)
	assert.Equal(t, stateOn, resolveState([]*settings{s}))
}

// A malformed managed-settings.json is skipped; resolution continues.
func TestFileBasedManagedMalformed(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath():            {Data: []byte(`{ not json`)},
		"home/.claude/settings.json": {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)},
	}
	// No readable managed source → falls through to user settings.
	assert.Equal(t, stateOn, resolveStatus(fsys, "/proj", "/home"))
}

// The managed tier sits above local/project/user: file-based managed off wins
// over a local settings.json that would otherwise report on.
func TestManagedAboveLocal(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath():                  {Data: []byte(`{"sandbox":{"enabled":false}}`)},
		"proj/.claude/settings.local.json": {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)},
	}
	assert.Equal(t, stateOff, resolveStatus(fsys, "/proj", "/home"))
}
