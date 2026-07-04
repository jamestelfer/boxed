package main

import (
	"encoding/json"
	"io/fs"

	"howett.net/plist"
)

const managedPlist = "/Library/Managed Preferences/com.anthropic.claudecode.plist"

// settings mirrors only the sandbox keys we care about. The matching json and
// plist tags let the same struct decode from settings files and managed prefs.
type settings struct {
	Sandbox *struct {
		Enabled                  *bool `json:"enabled" plist:"enabled"`
		AllowUnsandboxedCommands *bool `json:"allowUnsandboxedCommands" plist:"allowUnsandboxedCommands"`
	} `json:"sandbox" plist:"sandbox"`
}

// readManaged decodes the managed-preferences plist (binary or XML) in-process,
// returning nil if it is absent, unreadable, or malformed.
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

// readJSON decodes a settings file, returning nil for missing or malformed input.
func readJSON(fsys fs.FS, path string) *settings {
	b, err := fs.ReadFile(fsys, fsPath(path))
	if err != nil {
		return nil
	}
	var s settings
	if err := json.Unmarshal(b, &s); err != nil {
		return nil
	}
	return &s
}

// resolve returns the first non-nil value produced by get across sources, which
// are ordered highest precedence first.
func resolve(sources []*settings, get func(*settings) *bool) *bool {
	for _, s := range sources {
		if s == nil || s.Sandbox == nil {
			continue
		}
		if v := get(s); v != nil {
			return v
		}
	}
	return nil
}

// resolveState maps the resolved sandbox keys to one of the three states.
//
// Defaults (schemastore + code.claude.com): sandbox.enabled defaults false,
// sandbox.allowUnsandboxedCommands defaults true.
func resolveState(sources []*settings) state {
	enabled := resolve(sources, func(s *settings) *bool { return s.Sandbox.Enabled })
	allow := resolve(sources, func(s *settings) *bool { return s.Sandbox.AllowUnsandboxedCommands })

	switch {
	case enabled == nil || !*enabled:
		return stateOff
	case allow != nil && !*allow:
		return stateOn
	default:
		// enabled, but unsandboxed commands permitted (schema default is true)
		return statePartial
	}
}
