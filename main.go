// Command boxed prints the effective Claude Code sandbox status as a colored
// statusline label. It reads the managed-preferences plist in-process (binary
// or XML), so it needs no plutil subprocess.
//
// Precedence for sandbox.enabled and sandbox.allowUnsandboxedCommands (both
// TOP-LEVEL keys per https://json.schemastore.org/claude-code-settings.json),
// highest first:
//  1. managed tier: MDM plist + file-based managed settings (fail-safe on
//     conflict — least-protected status wins)
//  2. <project>/.claude/settings.local.json
//  3. <project>/.claude/settings.json
//  4. ~/.claude/settings.json
//
// Some managed plists misnest sandbox under permissions; reading the real
// top-level path deliberately ignores that. *bool tells an explicit false from
// an absent key, so an explicitly disabled setting is never mistaken for unset.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// errReported signals that a diagnostic was already written to stderr; main
// exits non-zero without printing again. Flag/usage errors are printed by
// urfave itself, so those propagate as ordinary errors and are also silent here.
var errReported = errors.New("error already reported")

// projectDir returns CLAUDE_PROJECT_DIR when set, otherwise the working
// directory. env/getwd are injected so the fallback is testable.
func projectDir(env func(string) string, getwd func() (string, error)) string {
	if d := env("CLAUDE_PROJECT_DIR"); d != "" {
		return d
	}
	d, _ := getwd()
	return d
}

// currentState resolves the effective sandbox state from the environment.
func currentState() state {
	proj := projectDir(os.Getenv, os.Getwd)
	home, _ := os.UserHomeDir()
	return resolveStatus(rootFS(), proj, home)
}

// newCommand builds the root urfave/cli v3 command.
func newCommand() *cli.Command {
	return &cli.Command{
		Name:    "boxed",
		Usage:   "print the effective Claude Code sandbox status as a styled label",
		Version: buildVersion(),
		Action: func(_ context.Context, cmd *cli.Command) error {
			if cmd.Args().Present() {
				fmt.Fprintf(os.Stderr, "boxed: unexpected argument %q\n", cmd.Args().First())
				return errReported
			}
			fmt.Print(render(currentState()))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "state",
				Usage: "print the bare resolved state token (on|partial|off), unstyled",
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.Args().Present() {
						fmt.Fprintf(os.Stderr, "boxed: unexpected argument %q\n", cmd.Args().First())
						return errReported
					}
					fmt.Println(currentState())
					return nil
				},
			},
		},
	}
}

func main() {
	if err := newCommand().Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
