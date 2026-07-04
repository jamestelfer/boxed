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
	"fmt"
	"os"
)

// projectDir returns CLAUDE_PROJECT_DIR when set, otherwise the working
// directory. env/getwd are injected so the fallback is testable.
func projectDir(env func(string) string, getwd func() (string, error)) string {
	if d := env("CLAUDE_PROJECT_DIR"); d != "" {
		return d
	}
	d, _ := getwd()
	return d
}

func main() {
	proj := projectDir(os.Getenv, os.Getwd)
	home, _ := os.UserHomeDir()

	fmt.Print(render(resolveStatus(rootFS(), proj, home)))
}
