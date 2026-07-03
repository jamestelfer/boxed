// Command boxed prints the effective Claude Code sandbox status as a colored
// statusline label. It reads the managed-preferences plist in-process (binary
// or XML), so it needs no plutil subprocess.
//
// Precedence for sandbox.enabled and sandbox.allowUnsandboxedCommands (both
// TOP-LEVEL keys per https://json.schemastore.org/claude-code-settings.json),
// highest first:
//  1. managed preferences (managedPlist)
//  2. <project>/.claude/settings.local.json
//  3. <project>/.claude/settings.json
//  4. ~/.claude/settings.json
//
// Some managed plists misnest sandbox under permissions; reading the real
// top-level path deliberately ignores that. *bool tells an explicit false from
// an absent key, so an explicitly disabled setting is never mistaken for unset.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"howett.net/plist"
)

const managedPlist = "/Library/Managed Preferences/com.anthropic.claudecode.plist"

// ANSI display attributes.
const (
	cGreen   = "\033[32m"
	cYellow  = "\033[33m"
	cRedBold = "\033[1;31m"
	cReset   = "\033[0m"
)

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
func readManaged() *settings {
	b, err := os.ReadFile(managedPlist)
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
func readJSON(path string) *settings {
	// Paths are well-known Claude Code settings locations, not user-controlled input.
	b, err := os.ReadFile(path) //nolint:gosec // G304: config paths are fixed, trusted locations
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

func main() {
	projectDir := os.Getenv("CLAUDE_PROJECT_DIR")
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	home, _ := os.UserHomeDir()

	// Sources in descending precedence.
	sources := []*settings{
		readManaged(),
		readJSON(filepath.Join(projectDir, ".claude", "settings.local.json")),
		readJSON(filepath.Join(projectDir, ".claude", "settings.json")),
		readJSON(filepath.Join(home, ".claude", "settings.json")),
	}

	enabled := resolve(sources, func(s *settings) *bool { return s.Sandbox.Enabled })
	allow := resolve(sources, func(s *settings) *bool { return s.Sandbox.AllowUnsandboxedCommands })

	color, label := cRedBold, "☢️ NOT sandboxed"
	switch {
	case enabled == nil || !*enabled:
		// off or unset
	case allow != nil && !*allow:
		color, label = cGreen, "📦 sandboxed"
	default:
		// enabled, but unsandboxed commands permitted (schema default is true)
		color, label = cYellow, "😬 sandbox (escape allowed)"
	}

	fmt.Printf("%s%s%s", color, label, cReset)
}
