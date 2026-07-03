# boxed

Prints the effective Claude Code sandbox status as a colored label, for use in a
statusline.

| Output | Color | Meaning |
| --- | --- | --- |
| `sandboxed` | green | `sandbox.enabled` is on and unsandboxed commands are disallowed |
| `sandbox (weak)` | yellow | enabled, but unsandboxed commands are allowed (the schema default) |
| `NOT sandboxed` | bold red | disabled or unset |

## How it resolves the setting

`boxed` reads two top-level keys from the [Claude Code settings
schema](https://json.schemastore.org/claude-code-settings.json) —
`sandbox.enabled` and `sandbox.allowUnsandboxedCommands` — resolving each
independently from these sources, highest precedence first:

1. managed preferences — `/Library/Managed Preferences/com.anthropic.claudecode.plist`
2. `$CLAUDE_PROJECT_DIR/.claude/settings.local.json`
3. `$CLAUDE_PROJECT_DIR/.claude/settings.json`
4. `~/.claude/settings.json`

`CLAUDE_PROJECT_DIR` falls back to the current working directory.

Two deliberate details:

- **Real schema, not the on-disk shape.** `sandbox` is a top-level key. Some
  managed plists misnest it under `permissions`; `boxed` reads the real
  top-level path, so a misnested block is ignored rather than trusted.
- **`false` is not "unset".** An explicitly disabled setting is distinguished
  from an absent one, so a `false` in a higher-precedence source is honored
  instead of being skipped.

The managed plist is decoded in-process ([`howett.net/plist`](https://github.com/DHowett/go-plist),
binary or XML) — no `plutil` subprocess, so startup is sub-millisecond.

## Build

Requires Go 1.26 (pinned via [mise](https://mise.jdx.dev)):

```sh
mise install      # installs go 1.26 from mise.toml
go build -o boxed .
```

## Use as a statusline

Point your Claude Code `statusLine` command at the binary:

```json
{
  "statusLine": {
    "type": "command",
    "command": "/path/to/boxed"
  }
}
```

## Caveat

`boxed` reads the managed-preferences file directly, which bypasses macOS
preference layering (`cfprefsd`). The canonically-correct read is
`CFPreferencesCopyAppValue`, but that requires cgo and forfeits a static,
cross-compilable binary. Reading the file matches how the equivalent shell
tooling behaves.
