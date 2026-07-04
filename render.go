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

// ANSI display attributes.
const (
	cGreen   = "\033[32m"
	cYellow  = "\033[33m"
	cRedBold = "\033[1;31m"
	cReset   = "\033[0m"
)

// render returns the colored statusline label for a state.
func render(s state) string {
	var color, label string
	switch s {
	case stateOn:
		color, label = cGreen, "📦 sandboxed"
	case statePartial:
		color, label = cYellow, "😬 sandbox (escape allowed)"
	default:
		color, label = cRedBold, "☢️ NOT sandboxed"
	}
	return color + label + cReset
}
