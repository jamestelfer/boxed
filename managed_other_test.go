//go:build !darwin

package main

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

// Off macOS, readManaged is a no-op: it contributes no settings regardless of
// filesystem contents, so resolution uses only the non-MDM sources.
func TestReadManagedNoOp(t *testing.T) {
	fsys := fstest.MapFS{
		"Library/Managed Preferences/com.anthropic.claudecode.plist": {
			Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`),
		},
	}
	assert.Nil(t, readManaged(fsys))
}
