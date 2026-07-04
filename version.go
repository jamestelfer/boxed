package main

import "runtime/debug"

// Build metadata. version is injected at release time via
// -ldflags "-X main.version=… -X main.commit=… -X main.date=…" (GoReleaser).
// For plain `go build`/`go install` these stay empty and buildVersion() derives
// a value from the embedded build info instead.
var (
	version = ""
	commit  = ""
	date    = ""
)

// buildVersion always returns a non-empty version string. urfave/cli only wires
// the built-in -v/--version flag when Command.Version is non-empty, so the
// fallback must never yield "".
func buildVersion() string {
	if version != "" {
		v := version
		if commit != "" {
			v += " (" + commit
			if date != "" {
				v += ", " + date
			}
			v += ")"
		}
		return v
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	var rev string
	var modified bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if rev == "" {
		return "(devel)"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if modified {
		return rev + "-dirty"
	}
	return rev
}
