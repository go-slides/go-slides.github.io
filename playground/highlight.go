package playground

import (
	"strings"

	"github.com/go-widgets/toolkit"
)

// deckSyntax colours a deck in the editor: the front matter's keys, the
// slide separators, headings, directive comments, template actions and
// fenced code. It is a toolkit.Highlighter.
type deckSyntax struct{}

// palette picks the colours for a theme: deeper on a light background,
// lighter on a dark one.
func syntaxColors(theme *toolkit.Theme) (key, sep, heading, comment, action, code toolkit.RGBA) {
	dark := theme != nil && int(theme.Background.R)+int(theme.Background.G)+int(theme.Background.B) < 384
	if dark {
		return toolkit.RGB(0x79, 0xc0, 0xff), toolkit.RGB(0xff, 0xa6, 0x57), toolkit.RGB(0xd2, 0xa8, 0xff),
			toolkit.RGB(0x7e, 0xe7, 0x87), toolkit.RGB(0xff, 0x7b, 0x72), toolkit.RGB(0x8b, 0x94, 0x9e)
	}
	return toolkit.RGB(0x05, 0x50, 0xae), toolkit.RGB(0xbc, 0x4c, 0x00), toolkit.RGB(0x6f, 0x42, 0xc1),
		toolkit.RGB(0x11, 0x63, 0x29), toolkit.RGB(0xcf, 0x22, 0x2e), toolkit.RGB(0x57, 0x60, 0x6a)
}

// Highlight implements toolkit.Highlighter. Spans are in rune columns.
func (deckSyntax) Highlight(_ string, lines []string, theme *toolkit.Theme) [][]toolkit.TextSpan {
	key, sep, heading, comment, action, code := syntaxColors(theme)
	out := make([][]toolkit.TextSpan, len(lines))
	inFront := len(lines) > 0 && strings.HasPrefix(lines[0], "---")
	hcl := inFront && strings.TrimSpace(lines[0]) == "---hcl"
	inComment, fence := false, ""
	for i, line := range lines {
		n := len([]rune(line))
		trimmed := strings.TrimSpace(line)
		var spans []toolkit.TextSpan
		switch {
		case i == 0 && inFront:
			spans = append(spans, toolkit.TextSpan{Start: 0, End: n, Color: sep})
		case inFront:
			if strings.HasPrefix(trimmed, "---") {
				inFront = false
				spans = append(spans, toolkit.TextSpan{Start: 0, End: n, Color: sep})
				break
			}
			spans = append(spans, frontKeys(line, hcl, key, code)...)
		case fence != "":
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
			spans = append(spans, toolkit.TextSpan{Start: 0, End: n, Color: code})
		case strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~"):
			fence = trimmed[:3]
			spans = append(spans, toolkit.TextSpan{Start: 0, End: n, Color: code})
		case trimmed == "---" || trimmed == "***" || trimmed == "___":
			spans = append(spans, toolkit.TextSpan{Start: 0, End: n, Color: sep})
		case strings.HasPrefix(trimmed, "#"):
			spans = append(spans, toolkit.TextSpan{Start: 0, End: n, Color: heading})
		}
		if fence == "" && !strings.HasPrefix(trimmed, "```") {
			spans = append(spans, delimited(line, "<!--", "-->", comment, &inComment)...)
			var inAction bool
			spans = append(spans, delimited(line, "{{", "}}", action, &inAction)...)
		}
		out[i] = spans
	}
	return out
}

// frontKeys colours the keys of a front matter line, and its comment.
//
// YAML (YAML 1.2.2, §7.4 flow collections and §8.2 block collections): a
// key is a plain or quoted scalar followed by ':' and a space, a line end,
// or a flow indicator. It starts the line (after the indentation and any
// "- " of a block sequence entry) or follows '{' or ',' inside a flow
// mapping, at any depth: `data:`, `  name: World`, `- id: 1`,
// `data: {name: World, n: 2}`.
//
// HCL (hclsyntax spec: attributes, blocks, object constructors): an
// identifier followed by '=' (not '==') is an attribute's name, an
// identifier that starts a line ending in '{' a block's type, and inside
// an object constructor a key is followed by '=' or ':'.
//
// A comment ('#', and '//' in HCL) runs to the end of the line; strings
// are skipped, so a ':' or '#' in one is neither.
func frontKeys(line string, hcl bool, key, comment toolkit.RGBA) []toolkit.TextSpan {
	r := []rune(line)
	var spans []toolkit.TextSpan
	isSpace := func(c rune) bool { return c == ' ' || c == '\t' }
	skip := func(i int) int {
		for i < len(r) && isSpace(r[i]) {
			i++
		}
		return i
	}
	// atKey is true where a key may start: the line's start, after a
	// block sequence's "- ", after '{' or ','.
	atKey, depth := true, 0
	i := skip(0)
	if !hcl {
		for i+1 < len(r) && r[i] == '-' && isSpace(r[i+1]) {
			i = skip(i + 1)
		}
		if i == len(r)-1 && r[i] == '-' {
			return nil
		}
	}
	lineStart := i
	for i < len(r) {
		c := r[i]
		if c == '#' || (hcl && c == '/' && i+1 < len(r) && r[i+1] == '/') {
			if i == 0 || isSpace(r[i-1]) || i == lineStart {
				spans = append(spans, toolkit.TextSpan{Start: i, End: len(r), Color: comment})
				break
			}
		}
		switch {
		case isSpace(c):
			i++
			continue
		case c == '{' || c == '[':
			depth++
			atKey = c == '{'
			i++
			continue
		case c == '}' || c == ']':
			if depth > 0 {
				depth--
			}
			atKey = false
			i++
			continue
		case c == ',':
			atKey = depth > 0
			i++
			continue
		}
		// a token: a quoted string or a run up to a delimiter
		start := i
		if c == '"' || c == '\'' {
			i++
			for i < len(r) && r[i] != c {
				if r[i] == '\\' && c == '"' {
					i++
				}
				i++
			}
			if i < len(r) {
				i++
			}
		} else {
			for i < len(r) && !isSpace(r[i]) && !strings.ContainsRune(":=,{}[]\"'", r[i]) {
				i++
			}
			if i < len(r) && r[i] == ':' && !hcl && i+1 < len(r) && !isSpace(r[i+1]) && !strings.ContainsRune(",{}[]", r[i+1]) {
				// a ':' inside a plain scalar (a URL, a time): not a key's
				for i < len(r) && !isSpace(r[i]) && !strings.ContainsRune(",{}[]", r[i]) {
					i++
				}
			}
			if i == start {
				i++ // a lone ':' or '='
				atKey = false
				continue
			}
		}
		end := i
		j := skip(i)
		isKey := false
		if atKey && j < len(r) {
			switch {
			case r[j] == ':' && (!hcl || depth > 0):
				isKey = j+1 == len(r) || isSpace(r[j+1]) || strings.ContainsRune(",{}[]", r[j+1]) || hcl
			case hcl && r[j] == '=' && (j+1 == len(r) || r[j+1] != '='):
				isKey = true
			}
		}
		if hcl && !isKey && start == lineStart && depth == 0 && strings.HasSuffix(strings.TrimSpace(line), "{") && r[start] != '"' {
			isKey = true // a block's type: kind "big" {
		}
		if isKey {
			spans = append(spans, toolkit.TextSpan{Start: start, End: end, Color: key})
			i = j + 1
		}
		atKey = false
	}
	return spans
}

// delimited colours the open...close stretches of a line; open carries a
// stretch over from the previous line (a comment over several lines).
func delimited(line, start, end string, c toolkit.RGBA, open *bool) []toolkit.TextSpan {
	var spans []toolkit.TextSpan
	runes := func(s string) int { return len([]rune(s)) }
	pos := 0
	for {
		from, search := pos, pos
		if !*open {
			k := strings.Index(line[pos:], start)
			if k < 0 {
				break
			}
			from = pos + k
			search = from + len(start)
			*open = true
		}
		to := len(line)
		if k := strings.Index(line[search:], end); k >= 0 {
			to = search + k + len(end)
			*open = false
		}
		spans = append(spans, toolkit.TextSpan{Start: runes(line[:from]), End: runes(line[:to]), Color: c})
		if *open || to >= len(line) {
			break
		}
		pos = to
	}
	return spans
}
