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

## Installation

Every release ships provenance-attested archives, a self-contained install
script, and a Homebrew formula. All download paths can be verified — see
[Verifying releases](#verifying-releases).

<details>
<summary><strong>mise (recommended)</strong></summary>

[mise](https://mise.jdx.dev/) installs directly from GitHub Releases via the
[GitHub backend](https://mise.jdx.dev/dev-tools/backends/github.html):

```sh
mise use -g github:jamestelfer/boxed
```

</details>

<details>
<summary><strong>Install script</strong></summary>

Each release ships a self-contained installer (generated with
[binstaller](https://github.com/binary-install/binstaller)) that detects your
platform and checks the download against checksums embedded in the script — no
separate checksum file is fetched:

```sh
curl -fsSL https://github.com/jamestelfer/boxed/releases/latest/download/install.sh | sh
```

It installs to `~/.local/bin`; pass `-b` for another directory and a tag to pin
a version:

```sh
curl -fsSL https://github.com/jamestelfer/boxed/releases/latest/download/install.sh \
  | sh -s -- -b /usr/local/bin v0.1.0
```

The script carries a build-provenance attestation, so you can verify it before
running it (see [Verifying releases](#verifying-releases)). This transitively
covers the binary too: a verified script is guaranteed to hold the genuine
checksums it then enforces on the download.

```sh
curl -fsSL -O https://github.com/jamestelfer/boxed/releases/latest/download/install.sh
gh attestation verify install.sh --repo jamestelfer/boxed
sh install.sh
```

</details>

<details>
<summary><strong>Homebrew (macOS)</strong></summary>

```sh
brew install jamestelfer/tap/boxed
```

</details>

<details>
<summary><strong>Manual download</strong></summary>

Grab the archive for your platform from the
[latest release](https://github.com/jamestelfer/boxed/releases/latest)
(`boxed_<os>_<arch>.tar.gz`, or `.zip` on Windows), verify its provenance, then
extract:

```sh
gh attestation verify boxed_linux_amd64.tar.gz --repo jamestelfer/boxed
tar -xzf boxed_linux_amd64.tar.gz
```

Move the extracted `boxed` binary somewhere on your `$PATH`. See
[Verifying releases](#verifying-releases) for details.

</details>

<details>
<summary><strong>Build from source</strong></summary>

Requires Go 1.26 (pinned via [mise](https://mise.jdx.dev)):

```sh
mise install      # installs go 1.26 from mise.toml
just build        # produces dist/boxed
```

Add `dist/boxed` to your `$PATH`, or install wherever suits.

</details>

## Verifying releases

Every published archive, the `install.sh` script, and `checksums.txt` carry a
[SLSA build-provenance](https://slsa.dev/) attestation, signed keylessly through
[Sigstore](https://www.sigstore.dev/) during the release workflow. Verify any
artifact with an authenticated [GitHub CLI](https://cli.github.com/):

```sh
gh attestation verify "$ARTIFACT" --repo jamestelfer/boxed
```

A successful verification proves the artifact was built by this repository's
release workflow and has not been altered since.

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
