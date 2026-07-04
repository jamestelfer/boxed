package main

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
)

// settings mirrors only the sandbox keys we care about. The matching json and
// plist tags let the same struct decode from settings files and managed prefs.
type settings struct {
	Sandbox *struct {
		Enabled                  *bool `json:"enabled" plist:"enabled"`
		AllowUnsandboxedCommands *bool `json:"allowUnsandboxedCommands" plist:"allowUnsandboxedCommands"`
	} `json:"sandbox" plist:"sandbox"`
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

// resolveStatus computes the effective sandbox state, highest authority first:
//
//  1. Managed tier: the MDM plist and the merged file-based managed settings.
//     Claude Code does not define how these two mechanisms combine, so when both
//     express a sandbox configuration and disagree, boxed fails safe to the
//     least-protected status (off > partial > on) rather than over-reporting.
//  2. Otherwise, per-key precedence across <project>/.claude/settings.local.json
//     → <project>/.claude/settings.json → ~/.claude/settings.json.
func resolveStatus(fsys fs.FS, proj, home string) state {
	if s, ok := managedState(fsys); ok {
		return s
	}
	nonManaged := []*settings{
		readJSON(fsys, filepath.Join(proj, ".claude", "settings.local.json")),
		readJSON(fsys, filepath.Join(proj, ".claude", "settings.json")),
		readJSON(fsys, filepath.Join(home, ".claude", "settings.json")),
	}
	return resolveState(nonManaged)
}

// managedState returns the managed-tier state and whether any managed source
// expressed a sandbox configuration. Each managed source's state is computed
// independently; conflicts fail safe to the least-protected status.
func managedState(fsys fs.FS) (state, bool) {
	var states []state
	for _, s := range []*settings{readManaged(fsys), fileBasedManaged(fsys)} {
		if s != nil && s.Sandbox != nil {
			states = append(states, resolveState([]*settings{s}))
		}
	}
	if len(states) == 0 {
		return 0, false
	}
	result := states[0]
	for _, st := range states[1:] {
		result = leastProtected(result, st)
	}
	return result, true
}

// leastProtected returns the weaker of two states (off < partial < on).
func leastProtected(a, b state) state {
	if a.protection() <= b.protection() {
		return a
	}
	return b
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
