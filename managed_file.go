package main

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
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

// fileBasedManaged reads the file-based managed settings: managed-settings.json
// as the base, then every *.json in the managed-settings.d/ drop-in directory
// merged on top in ascending alphabetical order (systemd convention — later
// files override earlier ones for scalar sandbox keys). Hidden files (starting
// with ".") are ignored. Absent/malformed sources are skipped. Returns nil when
// no readable source expressed anything.
func fileBasedManaged(fsys fs.FS) *settings {
	dir := managedDir()
	merged := readJSON(fsys, filepath.Join(dir, "managed-settings.json"))

	entries, err := fs.ReadDir(fsys, fsPath(filepath.Join(dir, "managed-settings.d")))
	if err != nil {
		return merged
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || strings.HasPrefix(n, ".") || !strings.HasSuffix(n, ".json") {
			continue
		}
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		merged = mergeSettings(merged, readJSON(fsys, filepath.Join(dir, "managed-settings.d", n)))
	}
	return merged
}

// mergeSettings returns base with over's non-nil sandbox scalars layered on top.
func mergeSettings(base, over *settings) *settings {
	if over == nil || over.Sandbox == nil {
		return base
	}
	if base == nil || base.Sandbox == nil {
		return over
	}
	sb := *base.Sandbox
	out := &settings{Sandbox: &sb}
	if over.Sandbox.Enabled != nil {
		out.Sandbox.Enabled = over.Sandbox.Enabled
	}
	if over.Sandbox.AllowUnsandboxedCommands != nil {
		out.Sandbox.AllowUnsandboxedCommands = over.Sandbox.AllowUnsandboxedCommands
	}
	return out
}
