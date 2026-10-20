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
			if k := strings.IndexAny(line, ":="); k > 0 && line[0] != ' ' && line[0] != '\t' && !strings.HasPrefix(trimmed, "#") {
				spans = append(spans, toolkit.TextSpan{Start: 0, End: len([]rune(line[:k])), Color: key})
			}
			if strings.HasPrefix(trimmed, "#") {
				spans = append(spans, toolkit.TextSpan{Start: 0, End: n, Color: code})
			}
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
