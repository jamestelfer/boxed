//go:build darwin

package main

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const managedPlistXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>sandbox</key>
  <dict>
    <key>enabled</key><true/>
    <key>allowUnsandboxedCommands</key><false/>
  </dict>
</dict>
</plist>`

// On macOS, readManaged decodes the managed-preferences plist (top-level
// sandbox keys) and feeds resolution.
func TestReadManagedDarwin(t *testing.T) {
	fsys := fstest.MapFS{
		fsPath(managedPlist): {Data: []byte(managedPlistXML)},
	}
	s := readManaged(fsys)
	require.NotNil(t, s)
	require.NotNil(t, s.Sandbox)
	require.NotNil(t, s.Sandbox.Enabled)
	assert.True(t, *s.Sandbox.Enabled)
	require.NotNil(t, s.Sandbox.AllowUnsandboxedCommands)
	assert.False(t, *s.Sandbox.AllowUnsandboxedCommands)
	assert.Equal(t, stateOn, resolveState([]*settings{s}))
}

// A missing/malformed plist yields nil (skipped, not fatal).
func TestReadManagedDarwinAbsent(t *testing.T) {
	assert.Nil(t, readManaged(fstest.MapFS{}))
	assert.Nil(t, readManaged(fstest.MapFS{fsPath(managedPlist): {Data: []byte("not a plist")}}))
}
