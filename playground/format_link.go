package playground

import (
	"regexp"
	"strings"
)

// The Format tab's Link command, Ctrl/Cmd+K as in Typora, GitHub's editor and
// Google Docs: it turns the selection (or the word at the caret) into [text](url)
// with the url selected for typing over, edits the link the caret is in by
// selecting its url, and with nothing to link writes [text](url) with the text
// selected.

var linkRe = regexp.MustCompile(`!?\[([^\]\n]*)\]\(([^)\n]*)\)`)

// linkAt finds the link (not an image) of the line that holds the columns a..b,
// as rune columns of its text and of its url.
func linkAt(line string, a, b int) (textA, textB, urlA, urlB int, ok bool) {
	for _, m := range linkRe.FindAllStringSubmatchIndex(line, -1) {
		if line[m[0]] == '!' {
			continue
		}
		// byte offsets to rune columns
		col := func(i int) int { return len([]rune(line[:i])) }
		start, end := col(m[0]), col(m[1])
		if a >= start && b <= end {
			return col(m[2]), col(m[3]), col(m[4]), col(m[5]), true
		}
	}
	return 0, 0, 0, 0, false
}

// inImage reports whether the columns a..b are inside an image, ![alt](src).
func inImage(line string, a, b int) bool {
	for _, m := range linkRe.FindAllStringIndex(line, -1) {
		if line[m[0]] != '!' {
			continue
		}
		if start, end := len([]rune(line[:m[0]])), len([]rune(line[:m[1]])); a >= start && b <= end {
			return true
		}
	}
	return false
}

func looksLikeURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "mailto:")
}

// LinkPlaceholder is the url a new link starts with, selected so that it is typed
// over.
const LinkPlaceholder = "https://"

// InsertLink makes or edits a link at the caret.
func InsertLink(src string, caret Pos, sel Range) Change {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	sel = normalize(lines, sel)
	// inside a link: select its url
	a, b := caret.Col, caret.Col
	if !sel.Empty() && sel.Start.Line == sel.End.Line {
		a, b = sel.Start.Col, sel.End.Col
	}
	if _, _, ua, ub, ok := linkAt(lines[caret.Line], a, b); ok {
		r := Range{Pos{caret.Line, ua}, Pos{caret.Line, ub}}
		return Change{From: caret, To: caret, Caret: r.End, Select: &r}
	}
	if inImage(lines[caret.Line], a, b) {
		return Change{From: caret, To: caret, Caret: caret} // an image is not a link
	}
	segs := segments(lines, caret, sel)
	if len(segs) == 1 && segs[0].a < segs[0].b {
		s := segs[0]
		text := string([]rune(lines[s.line])[s.a:s.b])
		if looksLikeURL(text) {
			// the url is what was selected: the text is typed
			out := "[](" + text + ")"
			r := Range{Pos{s.line, s.a + 1}, Pos{s.line, s.a + 1}}
			return Change{From: Pos{s.line, s.a}, To: Pos{s.line, s.b}, Text: out, Caret: r.End, Select: &r}
		}
		out := "[" + text + "](" + LinkPlaceholder + ")"
		ua := s.a + 1 + len([]rune(text)) + 2
		r := Range{Pos{s.line, ua}, Pos{s.line, ua + len([]rune(LinkPlaceholder))}}
		return Change{From: Pos{s.line, s.a}, To: Pos{s.line, s.b}, Text: out, Caret: r.End, Select: &r}
	}
	out := "[text](" + LinkPlaceholder + ")"
	r := Range{Pos{caret.Line, caret.Col + 1}, Pos{caret.Line, caret.Col + 5}}
	return Change{From: caret, To: caret, Text: out, Caret: r.End, Select: &r}
}

// InLink reports whether the caret is in a link.
func InLink(src string, caret Pos) bool {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	_, _, _, _, ok := linkAt(lines[caret.Line], caret.Col, caret.Col)
	return ok
}
