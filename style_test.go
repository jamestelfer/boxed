package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal text (outside any group) passes through untouched.
func TestRenderFormatLiteral(t *testing.T) {
	out, err := renderFormat("plain text 123")
	require.NoError(t, err)
	assert.Equal(t, "plain text 123", out)
}

// Each supported style token renders and produces ANSI.
func TestStyleTokens(t *testing.T) {
	tokens := []string{
		"bold", "italic", "underline", "dimmed", "inverted", "strikethrough",
		"red", "bright-red", "green", "blue", "purple", "magenta", "cyan", "white", "black", "yellow",
		"fg:red", "bg:blue", "#ff8800", "bg:#00ff00", "42", "255", "0",
	}
	for _, tok := range tokens {
		out, err := renderFormat("[x](" + tok + ")")
		require.NoErrorf(t, err, "token %q", tok)
		assert.Containsf(t, out, "x", "token %q", tok)
		if tok != "none" {
			assert.Containsf(t, out, "\x1b[", "token %q should emit ANSI", tok)
		}
	}
}

// Combined tokens render (bold red matches the off default's style).
func TestStyleCombined(t *testing.T) {
	out, err := renderFormat("[x](bold red)")
	require.NoError(t, err)
	assert.Equal(t, "\x1b[1;31mx\x1b[m", out)
}

// Malformed groups and unsupported tokens are hard errors naming the input.
func TestRenderFormatErrors(t *testing.T) {
	cases := []string{
		"[unclosed group",
		"[text]no parens",
		"[text](unclosed style",
		"[x](notacolour)",
		"[x](#12)",
		"[x](999)",
		"[x](-1)",
	}
	for _, c := range cases {
		_, err := renderFormat(c)
		require.Errorf(t, err, "expected error for %q", c)
		assert.Containsf(t, err.Error(), strings.TrimSpace(c[:1]), "error should reference input for %q", c)
	}
}
