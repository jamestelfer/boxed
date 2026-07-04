package main

import (
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sandboxJSON builds a settings.json body for the given optional keys.
func sandboxJSON(enabled, allow *bool) string {
	body := `{"sandbox":{`
	sep := ""
	if enabled != nil {
		body += `"enabled":` + boolStr(*enabled)
		sep = ","
	}
	if allow != nil {
		body += sep + `"allowUnsandboxedCommands":` + boolStr(*allow)
	}
	return body + `}}`
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// enabled × allow → state mapping, driven through resolveState.
func TestResolveStateMapping(t *testing.T) {
	cases := []struct {
		name    string
		enabled *bool
		allow   *bool
		want    state
	}{
		{"unset/unset", nil, nil, stateOff},
		{"unset/allow-true", nil, new(true), stateOff},
		{"unset/allow-false", nil, new(false), stateOff},
		{"enabled-false/unset", new(false), nil, stateOff},
		{"enabled-false/allow-false", new(false), new(false), stateOff},
		{"enabled-true/unset", new(true), nil, statePartial},
		{"enabled-true/allow-true", new(true), new(true), statePartial},
		{"enabled-true/allow-false", new(true), new(false), stateOn},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &settings{}
			s.Sandbox = &struct {
				Enabled                  *bool `json:"enabled" plist:"enabled"`
				AllowUnsandboxedCommands *bool `json:"allowUnsandboxedCommands" plist:"allowUnsandboxedCommands"`
			}{Enabled: c.enabled, AllowUnsandboxedCommands: c.allow}
			assert.Equal(t, c.want, resolveState([]*settings{s}))
		})
	}
}

// Four-tier precedence: first present key wins, per key independently.
func TestPrecedence(t *testing.T) {
	const (
		local   = "proj/.claude/settings.local.json"
		project = "proj/.claude/settings.json"
		user    = "home/.claude/settings.json"
	)
	fsys := fstest.MapFS{
		local:   {Data: []byte(sandboxJSON(new(true), new(false)))}, // on
		project: {Data: []byte(sandboxJSON(new(false), nil))},       // off
		user:    {Data: []byte(sandboxJSON(new(true), new(true)))},  // partial
	}
	sources := []*settings{
		readManaged(fsys),
		readJSON(fsys, "/"+local),
		readJSON(fsys, "/"+project),
		readJSON(fsys, "/"+user),
	}
	// local wins: enabled=true, allow=false → on
	assert.Equal(t, stateOn, resolveState(sources))
}

// Per-key precedence: enabled from a lower tier, allow from a higher tier.
func TestPrecedencePerKey(t *testing.T) {
	fsys := fstest.MapFS{
		"proj/.claude/settings.local.json": {Data: []byte(`{"sandbox":{"allowUnsandboxedCommands":false}}`)},
		"proj/.claude/settings.json":       {Data: []byte(`{"sandbox":{"enabled":true}}`)},
	}
	sources := []*settings{
		nil,
		readJSON(fsys, "/proj/.claude/settings.local.json"),
		readJSON(fsys, "/proj/.claude/settings.json"),
		nil,
	}
	// enabled=true (project) + allow=false (local) → on
	assert.Equal(t, stateOn, resolveState(sources))
}

// Misnested sandbox under permissions must be ignored (read as top-level only).
func TestMisnestedSandboxIgnored(t *testing.T) {
	fsys := fstest.MapFS{
		"proj/.claude/settings.json": {Data: []byte(`{"permissions":{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}}`)},
	}
	s := readJSON(fsys, "/proj/.claude/settings.json")
	require.NotNil(t, s)
	assert.Nil(t, s.Sandbox)
	assert.Equal(t, stateOff, resolveState([]*settings{s}))
}

// Explicit false in a higher-precedence source is honoured over a lower true.
func TestExplicitFalseHonoured(t *testing.T) {
	fsys := fstest.MapFS{
		"proj/.claude/settings.local.json": {Data: []byte(`{"sandbox":{"enabled":false}}`)},
		"home/.claude/settings.json":       {Data: []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`)},
	}
	sources := []*settings{
		nil,
		readJSON(fsys, "/proj/.claude/settings.local.json"),
		nil,
		readJSON(fsys, "/home/.claude/settings.json"),
	}
	assert.Equal(t, stateOff, resolveState(sources))
}

// Absent and malformed sources are skipped, not fatal.
func TestAbsentAndMalformedSkipped(t *testing.T) {
	fsys := fstest.MapFS{
		"proj/.claude/settings.json": {Data: []byte(`{ this is not json`)},
	}
	assert.Nil(t, readJSON(fsys, "/proj/.claude/settings.json"))
	assert.Nil(t, readJSON(fsys, "/does/not/exist.json"))
	assert.Nil(t, readManaged(fsys))
}

// fsPath strips the leading slash and normalises separators.
func TestFSPath(t *testing.T) {
	assert.Equal(t, "Library/Managed Preferences/x.plist", fsPath("/Library/Managed Preferences/x.plist"))
	assert.Equal(t, "tmp/pf/.claude/settings.json", fsPath(filepath.Join("/tmp/pf", ".claude", "settings.json")))
	assert.Equal(t, ".", fsPath("/"))
}
