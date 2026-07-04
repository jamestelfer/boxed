package main

// state is the effective sandbox status. Exactly three values exist; this
// vocabulary is shared by code, the `state` token, flag names, and docs.
type state int

const (
	stateOff     state = iota // no sandbox in effect
	statePartial              // sandbox enabled, but unsandboxed commands allowed
	stateOn                   // sandbox enabled, unsandboxed commands denied
)

// protection ranks states by how much sandboxing they guarantee: off is the
// least protected, on the most. Used by the managed-tier fail-safe.
func (s state) protection() int {
	switch s {
	case stateOn:
		return 2
	case statePartial:
		return 1
	default: // stateOff
		return 0
	}
}

func (s state) String() string {
	switch s {
	case stateOn:
		return "on"
	case statePartial:
		return "partial"
	default:
		return "off"
	}
}

// defaultFormats pins the starship-style format string for each state. They
// reproduce the historical hardcoded labels/colours: green "sandboxed", yellow
// "sandbox (escape allowed)", bold-red "NOT sandboxed".
var defaultFormats = map[state]string{
	stateOn:      "[📦 sandboxed](green)",
	statePartial: "[😬 sandbox (escape allowed)](yellow)",
	stateOff:     "[☢️ NOT sandboxed](bold red)",
}

// render returns the styled statusline label for a state. format overrides the
// pinned default when non-empty. A malformed format string is a hard error.
func render(s state, format string) (string, error) {
	if format == "" {
		format = defaultFormats[s]
	}
	return renderFormat(format)
}
