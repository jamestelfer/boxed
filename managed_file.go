package main

import (
	"io/fs"
	"path/filepath"
	"runtime"
)

// managedDir returns the OS-specific directory holding file-based managed
// settings, per the Claude Code settings docs:
//   - macOS:      /Library/Application Support/ClaudeCode
//   - Linux/WSL:  /etc/claude-code
//   - Windows:    C:\Program Files\ClaudeCode
func managedDir() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Library/Application Support/ClaudeCode"
	case "windows":
		return `C:\Program Files\ClaudeCode`
	default:
		return "/etc/claude-code"
	}
}

// readManagedFile decodes the file-based managed-settings.json, the secondary
// managed source ranked below the MDM plist. Absent/malformed input yields nil.
func readManagedFile(fsys fs.FS) *settings {
	return readJSON(fsys, filepath.Join(managedDir(), "managed-settings.json"))
}
