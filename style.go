package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// renderFormat renders a starship-style format string into an ANSI string.
//
// The grammar is deliberately small: literal text is emitted verbatim, and a
// `[text](style)` group renders text with the given space-separated style
// tokens. There is no nesting and there are no variables. A malformed group or
// an unsupported style token is a hard error whose message names the offending
// input, so callers can surface it and exit non-zero.
//
// Rendering uses lipgloss v2's Style.Render, which always emits full-fidelity
// ANSI; the result is written raw (no colour-profile writer), so boxed emits
// exactly the requested colours regardless of NO_COLOR/CLICOLOR/FORCE_COLOR or
// whether stdout is a TTY.
func renderFormat(format string) (string, error) {
	var b strings.Builder
	rest := format
	for len(rest) > 0 {
		i := strings.IndexByte(rest, '[')
		if i < 0 {
			b.WriteString(rest)
			break
		}
		b.WriteString(rest[:i])
		rest = rest[i:]

		close := strings.IndexByte(rest, ']')
		if close < 0 {
			return "", fmt.Errorf("malformed format %q: unclosed '[' group", format)
		}
		text := rest[1:close]
		rest = rest[close+1:]
		if len(rest) == 0 || rest[0] != '(' {
			return "", fmt.Errorf("malformed format %q: '[...]' must be followed by '(style)'", format)
		}
		styleEnd := strings.IndexByte(rest, ')')
		if styleEnd < 0 {
			return "", fmt.Errorf("malformed format %q: unclosed '(' style", format)
		}
		spec := rest[1:styleEnd]
		rest = rest[styleEnd+1:]

		style, err := parseStyle(spec, format)
		if err != nil {
			return "", err
		}
		b.WriteString(style.Render(text))
	}
	return b.String(), nil
}

// parseStyle builds a lipgloss style from a space-separated token list.
func parseStyle(spec, format string) (lipgloss.Style, error) {
	style := lipgloss.NewStyle()
	for tok := range strings.FieldsSeq(spec) {
		switch tok {
		case "bold":
			style = style.Bold(true)
		case "italic":
			style = style.Italic(true)
		case "underline":
			style = style.Underline(true)
		case "dimmed":
			style = style.Faint(true)
		case "inverted":
			style = style.Reverse(true)
		case "strikethrough":
			style = style.Strikethrough(true)
		case "none":
			// explicit "no styling" token; contributes nothing
		default:
			c, bg, err := parseColor(tok, format)
			if err != nil {
				return style, err
			}
			if bg {
				style = style.Background(c)
			} else {
				style = style.Foreground(c)
			}
		}
	}
	return style, nil
}

// basic ANSI colour names → 3/4-bit palette index.
var colorNames = map[string]int{
	"black": 0, "red": 1, "green": 2, "yellow": 3,
	"blue": 4, "purple": 5, "magenta": 5, "cyan": 6, "white": 7,
}

// parseColor resolves a single colour token. An optional "fg:"/"bg:" prefix
// selects the layer (foreground default). The colour itself is a name, a
// "bright-" name, a 0–255 palette number, or a "#rrggbb" hex value.
func parseColor(tok, format string) (c color.Color, bg bool, err error) {
	switch {
	case strings.HasPrefix(tok, "fg:"):
		tok = tok[3:]
	case strings.HasPrefix(tok, "bg:"):
		tok, bg = tok[3:], true
	}

	// #rrggbb hex.
	if strings.HasPrefix(tok, "#") {
		if len(tok) != 7 {
			return nil, bg, fmt.Errorf("invalid hex colour %q in format %q", tok, format)
		}
		if _, e := strconv.ParseUint(tok[1:], 16, 32); e != nil {
			return nil, bg, fmt.Errorf("invalid hex colour %q in format %q", tok, format)
		}
		return lipgloss.Color(tok), bg, nil
	}

	// 0–255 palette number.
	if n, e := strconv.Atoi(tok); e == nil {
		if n < 0 || n > 255 {
			return nil, bg, fmt.Errorf("colour number %q out of range 0–255 in format %q", tok, format)
		}
		return lipgloss.Color(strconv.Itoa(n)), bg, nil
	}

	// name / bright-name.
	name := tok
	bright := false
	if strings.HasPrefix(name, "bright-") {
		name, bright = name[len("bright-"):], true
	}
	idx, ok := colorNames[name]
	if !ok {
		return nil, bg, fmt.Errorf("unsupported colour %q in format %q", tok, format)
	}
	if bright {
		idx += 8
	}
	return lipgloss.Color(strconv.Itoa(idx)), bg, nil
}
