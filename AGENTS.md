# boxed

CLI tool that prints the effective Claude Code sandbox status as a colored label, for use in a statusline. It reads the managed-preferences plist and `.claude/settings*.json` files, resolving `sandbox.enabled` and `sandbox.allowUnsandboxedCommands` by precedence.

## Go version

This project uses **Go 1.26**, which was released after the AI knowledge cutoff. Do not rely on training data for Go stdlib or dependency APIs — always fetch current documentation via Context7 before using unfamiliar APIs.

## Build and test

```
just verify    # fmt + build + lint + test (run before committing)
just build     # produces dist/boxed
just test      # go test ./...
just fmt       # gofmt -w .
just lint      # golangci-lint run ./...
```

Tool versions (Go, golangci-lint, just) are pinned via mise (`mise.toml`); run `mise install` to match CI.

## Project layout

```
main.go        entry point; reads settings sources, resolves sandbox state, prints label
test/          fixture input (input.json)
```

`main.go` is deliberately a single file: struct `settings` decodes both plist and JSON, `readManaged`/`readJSON` load sources, and `resolve` walks them in precedence order.

## Commits and PR titles

Use Conventional Commits for all commit messages and PR titles. The `pr-title.yml` workflow enforces the format on PRs (the squash-merge commit is taken from the PR title). There is no release automation, so commit types carry no version-bump semantics.

| Type | When to use |
|---|---|
| `feat: <description>` | new user-visible feature |
| `fix: <description>` | bug fix |
| `chore:`, `docs:`, `refactor:`, `test:`, `ci:` | maintenance, no behaviour change |

## Key conventions

- **Real schema, not the on-disk shape.** `sandbox` is a top-level settings key. Some managed plists misnest it under `permissions`; read the real top-level path so a misnested block is ignored.
- **`false` is not "unset".** Use `*bool` so an explicitly disabled setting is distinguished from an absent one; a `false` in a higher-precedence source must be honored, not skipped.
- **In-process plist decode.** Decode the managed plist with `howett.net/plist` (binary or XML) — no `plutil` subprocess.
- **Precedence order:** managed preferences → `$CLAUDE_PROJECT_DIR/.claude/settings.local.json` → `.claude/settings.json` → `~/.claude/settings.json`. `CLAUDE_PROJECT_DIR` falls back to the working directory.

## Major dependencies

Use Context7 for up-to-date documentation — do not guess at APIs.

| Library | Notes |
|---|---|
| `howett.net/plist` | In-process plist decode (binary or XML) for managed preferences |
