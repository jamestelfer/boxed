//go:build darwin

package main

import (
	"io/fs"

	"howett.net/plist"
)

// managedPlist is the macOS managed-preferences path written by MDM.
const managedPlist = "/Library/Managed Preferences/com.anthropic.claudecode.plist"

// readManaged decodes the managed-preferences plist (binary or XML) in-process,
// returning nil if it is absent, unreadable, or malformed. macOS only.
func readManaged(fsys fs.FS) *settings {
	b, err := fs.ReadFile(fsys, fsPath(managedPlist))
	if err != nil {
		return nil
	}
	var s settings
	if _, err := plist.Unmarshal(b, &s); err != nil {
		return nil
	}
	return &s
}
