package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// rootFS returns an fs.FS rooted at the filesystem root. Well-known settings
// paths are absolute; fsPath converts them to the slash-separated, no-leading-
// slash form fs.FS requires.
func rootFS() fs.FS {
	return os.DirFS("/")
}

// fsPath converts an absolute OS path into an fs.FS-relative path: forward
// slashes, no leading slash. An empty result maps to ".", which fs.ReadFile
// rejects with a path error (harmless: treated as an absent source).
func fsPath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return "."
	}
	return p
}
