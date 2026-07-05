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

// nonManagedPaths returns the three non-managed settings file paths, highest
// precedence first.
func nonManagedPaths(proj, home string) [3]string {
	return [3]string{
		filepath.Join(proj, ".claude", "settings.local.json"),
		filepath.Join(proj, ".claude", "settings.json"),
		filepath.Join(home, ".claude", "settings.json"),
	}
}

// resolveStatus computes the effective sandbox state; see determineState for
// the precedence it follows.
func resolveStatus(fsys fs.FS, proj, home string) state {
	s, _, _ := determineState(fsys, proj, home)
	return s
}

// sandboxEnabled and sandboxAllow are the accessors resolve() walks sources
// with, shared by resolveState and determineState so both resolve the same
// two keys the same way.
func sandboxEnabled(s *settings) *bool { return s.Sandbox.Enabled }
func sandboxAllow(s *settings) *bool   { return s.Sandbox.AllowUnsandboxedCommands }

// keyOrigin is the resolved value of one sandbox key and which settings file
// supplied it, for the doctor command. Origin is "" and Value is nil when no
// non-managed source set the key (the schema default applies).
type keyOrigin struct {
	key    string
	value  *bool
	origin string
}

// determineState computes the effective sandbox state, highest authority
// first, plus where it came from:
//
//  1. Managed tier: the MDM plist and the merged file-based managed settings.
//     Claude Code does not define how these two mechanisms combine, so when both
//     express a sandbox configuration and disagree, boxed fails safe to the
//     least-protected status (off > partial > on) rather than over-reporting.
//     This tier decides atomically — managed is true and keys is nil, since
//     there's no single winning per-key value once two states are combined.
//  2. Otherwise, per-key precedence across <project>/.claude/settings.local.json
//     → <project>/.claude/settings.json → ~/.claude/settings.json. Each key is
//     resolved independently, so keys names the specific settings file behind
//     each one.
func determineState(fsys fs.FS, proj, home string) (s state, managed bool, keys []keyOrigin) {
	if st, ok := managedState(fsys); ok {
		return st, true, nil
	}

	paths := nonManagedPaths(proj, home)
	sources := []*settings{readJSON(fsys, paths[0]), readJSON(fsys, paths[1]), readJSON(fsys, paths[2])}

	label := func(i int) string {
		if i < 0 {
			return ""
		}
		return paths[i]
	}

	enabled, ei := resolve(sources, sandboxEnabled)
	allow, ai := resolve(sources, sandboxAllow)

	return stateFromKeys(enabled, allow), false, []keyOrigin{
		{"sandbox.enabled", enabled, label(ei)},
		{"sandbox.allowUnsandboxedCommands", allow, label(ai)},
	}
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

// resolve returns the first non-nil value produced by get across sources,
// which are ordered highest precedence first, along with the index of the
// source it came from (-1 if none matched).
func resolve(sources []*settings, get func(*settings) *bool) (*bool, int) {
	for i, s := range sources {
		if s == nil || s.Sandbox == nil {
			continue
		}
		if v := get(s); v != nil {
			return v, i
		}
	}
	return nil, -1
}

// resolveState maps the resolved sandbox keys to one of the three states.
func resolveState(sources []*settings) state {
	enabled, _ := resolve(sources, sandboxEnabled)
	allow, _ := resolve(sources, sandboxAllow)
	return stateFromKeys(enabled, allow)
}

// stateFromKeys maps the resolved sandbox keys to one of the three states.
//
// Defaults (schemastore + code.claude.com): sandbox.enabled defaults false,
// sandbox.allowUnsandboxedCommands defaults true.
func stateFromKeys(enabled, allow *bool) state {
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
