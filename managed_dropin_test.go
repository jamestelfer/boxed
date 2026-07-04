package main

import (
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

// dropinPath returns the fs.FS key for a drop-in file of the given name.
func dropinPath(name string) string {
	return fsPath(filepath.Join(managedDir(), "managed-settings.d", name))
}

// Later drop-in files override earlier ones for a scalar sandbox key.
func TestDropinOverridesAlphabetical(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath():       {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)}, // on
		dropinPath("10-a.json"): {Data: []byte(`{"sandbox":{"allowUnsandboxedCommands":false}}`)},
		dropinPath("20-b.json"): {Data: []byte(`{"sandbox":{"allowUnsandboxedCommands":true}}`)}, // wins → partial
	}
	assert.Equal(t, statePartial, resolveStatus(fsys, "/proj", "/home"))
}

// The base file provides values that a drop-in only partially overrides.
func TestDropinLayersOnBase(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath():     {Data: []byte(`{"sandbox":{"enabled":true}}`)},                   // partial (base)
		dropinPath("50.json"): {Data: []byte(`{"sandbox":{"allowUnsandboxedCommands":false}}`)}, // → on
	}
	assert.Equal(t, stateOn, resolveStatus(fsys, "/proj", "/home"))
}

// Hidden drop-in files (starting with ".") are ignored.
func TestDropinDotfileIgnored(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath():          {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)}, // on
		dropinPath(".hidden.json"): {Data: []byte(`{"sandbox":{"enabled":false}}`)},                                 // ignored
	}
	assert.Equal(t, stateOn, resolveStatus(fsys, "/proj", "/home"))
}

// Non-.json drop-in files are ignored.
func TestDropinNonJSONIgnored(t *testing.T) {
	fsys := fstest.MapFS{
		managedFilePath():       {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)},
		dropinPath("notes.txt"): {Data: []byte(`this is not json`)},
	}
	assert.Equal(t, stateOn, resolveStatus(fsys, "/proj", "/home"))
}

// leastProtected picks the weaker state.
func TestLeastProtected(t *testing.T) {
	assert.Equal(t, stateOff, leastProtected(stateOn, stateOff))
	assert.Equal(t, stateOff, leastProtected(statePartial, stateOff))
	assert.Equal(t, statePartial, leastProtected(stateOn, statePartial))
	assert.Equal(t, stateOn, leastProtected(stateOn, stateOn))
}
