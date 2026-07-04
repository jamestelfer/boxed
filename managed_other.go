//go:build !darwin

package main

import "io/fs"

// readManaged is a no-op off macOS: MDM managed preferences are a macOS
// mechanism, so non-Darwin builds contribute no settings from this source and
// resolution proceeds using only the file-based and user sources.
func readManaged(fs.FS) *settings { return nil }
