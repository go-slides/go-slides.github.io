package playground

import (
	"strings"
	"unicode"
)

// The Format tab's inline commands: bold, italic, underline, strikethrough and
// code. What Marp's Markdown can say (bibliography, opened for this change):
//
//   - CommonMark and GFM: **bold** (or __bold__), *italic* (or _italic_), ~~strike~~
//     and `code`;
//   - underline has no Markdown syntax: it is <u>…</u>, and `u` is in the HTML
//     allow-list of marp-core 4.4 that go-slides keeps (deck/sanitize.go, the
//     default policy), so it works with raw HTML off. <mark>, <sub> and <sup> are
//     in it too.
//
// Each command toggles, as the Markdown All in One extension (Ctrl/Cmd+B, Ctrl/Cmd+I)
// and Typora (Ctrl/Cmd+B, I, U, K) do: with a selection it marks the selected text
// (or unmarks it when it is already marked), with none it marks the word at the
// caret, and with neither it puts an empty pair of marks and the caret between
// them. A mark is found again by pairing the marks of the line, so that a caret
// anywhere in **two words** sees the bold, and toggling there removes it.

// Marker is an inline mark: the text it puts before and after what it marks.
type Marker struct {
	Name        string
	Open, Close string
}

// The inline marks of the Format tab. Italic is written with underscores, as the
// playground's own examples do; a *star* italic is found and removed as well.
var (
	MarkBold      = Marker{"bold", "**", "**"}
	MarkItalic    = Marker{"italic", "_", "_"}
	MarkStrike    = Marker{"strike", "~~", "~~"}
	MarkCode      = Marker{"code", "`", "`"}
	MarkUnderline = Marker{"underline", "<u>", "</u>"}
)

// markerSpan is one marked stretch of a line: the open mark is [os, oe) and the
// close mark [cs, ce), in rune columns.
type markerSpan struct{ os, oe, cs, ce int }

// isWordRune is what a word is made of when the caret stands in one.
func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// codeCover marks the runes of a line that are inside a code span (the backtick
// runs of the same length pair up) or are the backticks themselves.
func codeCover(line []rune) []bool {
	cover := make([]bool, len(line))
	for i := 0; i < len(line); {
		if line[i] != '`' || i > 0 && line[i-1] == '\\' {
			i++
			continue
		}
		j := i
		for j < len(line) && line[j] == '`' {
			j++
		}
		n := j - i
		// the closing run of the same length
		k := j
		closeAt := -1
		for k < len(line) {
			if line[k] != '`' {
				k++
				continue
			}
			e := k
			for e < len(line) && line[e] == '`' {
				e++
			}
			if e-k == n {
				closeAt = k
				break
			}
			k = e
		}
		if closeAt < 0 {
			i = j
			continue
		}
		for x := i; x < closeAt+n; x++ {
			cover[x] = true
		}
		i = closeAt + n
	}
	return cover
}

// runAt is the length of the run of r starting at i.
func runAt(line []rune, i int, r rune) int {
	n := 0
	for i+n < len(line) && line[i+n] == r {
		n++
	}
	return n
}

// intraword reports whether the underscore run [i, i+n) sits between two letters
// or digits (snake_case is not an emphasis).
func intraword(line []rune, i, n int) bool {
	return i > 0 && i+n < len(line) && isWordRune(line[i-1]) && isWordRune(line[i+n])
}

// token is a mark found in a line.
type token struct{ s, e int }

// markerTokens are the marks of m in the line, in order, left out when they sit in
// a code span (except the code mark itself) or are escaped.
func markerTokens(line []rune, m Marker) []token {
	if m.Name == "underline" {
		return nil // pairs of tags, found by htmlSpans
	}
	cover := codeCover(line)
	var out []token
	escaped := func(i int) bool { return i > 0 && line[i-1] == '\\' }
	for i := 0; i < len(line); {
		c := line[i]
		switch {
		case m.Name == "code" && c == '`':
			n := runAt(line, i, '`')
			if n == 1 && !escaped(i) {
				out = append(out, token{i, i + 1})
			}
			i += n
		case c == '`':
			i += runAt(line, i, '`')
		case cover[i] || escaped(i):
			i++
		case (c == '*' || c == '_') && (m.Name == "bold" || m.Name == "italic"):
			n := runAt(line, i, c)
			opener := len(out)%2 == 0
			switch {
			case c == '_' && intraword(line, i, n):
			case m.Name == "bold" && n == 2:
				out = append(out, token{i, i + 2})
			case m.Name == "bold" && n == 3 && c == '*':
				if opener {
					out = append(out, token{i + 1, i + 3})
				} else {
					out = append(out, token{i, i + 2})
				}
			case m.Name == "italic" && n == 1:
				out = append(out, token{i, i + 1})
			case m.Name == "italic" && n == 3 && c == '*':
				if opener {
					out = append(out, token{i, i + 1})
				} else {
					out = append(out, token{i + 2, i + 3})
				}
			}
			i += n
		case c == '~' && m.Name == "strike":
			n := runAt(line, i, '~')
			if n == 2 {
				out = append(out, token{i, i + 2})
			}
			i += n
		default:
			i++
		}
	}
	return out
}

// htmlSpans pairs the open tags of m with the next close tag, ignoring case.
func htmlSpans(line []rune, m Marker) []markerSpan {
	low := []rune(strings.ToLower(string(line)))
	open, close := []rune(m.Open), []rune(m.Close)
	at := func(from int, w []rune) int {
		for i := from; i+len(w) <= len(low); i++ {
			if string(low[i:i+len(w)]) == string(w) {
				return i
			}
		}
		return -1
	}
	var out []markerSpan
	for i := at(0, open); i >= 0; {
		j := at(i+len(open), close)
		if j < 0 {
			break
		}
		out = append(out, markerSpan{i, i + len(open), j, j + len(close)})
		i = at(j+len(close), open)
	}
	return out
}

// markSpans are the marked stretches of a line: the marks paired up in order.
func markSpans(line []rune, m Marker) []markerSpan {
	if m.Name == "underline" {
		return htmlSpans(line, m)
	}
	toks := markerTokens(line, m)
	var out []markerSpan
	for i := 0; i+1 < len(toks); i += 2 {
		out = append(out, markerSpan{toks[i].s, toks[i].e, toks[i+1].s, toks[i+1].e})
	}
	return out
}

// seg is a stretch [a, b) of a line to mark, in rune columns.
type seg struct{ line, a, b int }

// normalize puts a range in order and inside the text.
func normalize(lines []string, r Range) Range {
	a, b := clampPos(lines, r.Start), clampPos(lines, r.End)
	if b.Line < a.Line || b.Line == a.Line && b.Col < a.Col {
		a, b = b, a
	}
	return Range{a, b}
}

// segments are what a mark applies to: each non-blank line of the selection
// (trimmed, and after a list, quote or heading mark the line starts with), or
// else the word at the caret, or else the caret's own place.
func segments(lines []string, caret Pos, sel Range) []seg {
	if sel = normalize(lines, sel); !sel.Empty() {
		var out []seg
		for l := sel.Start.Line; l <= sel.End.Line; l++ {
			r := []rune(lines[l])
			a, b := 0, len(r)
			if l == sel.Start.Line {
				a = sel.Start.Col
			}
			if l == sel.End.Line {
				b = sel.End.Col
			}
			a = max(a, blockPrefixLen(r))
			for a < b && unicode.IsSpace(r[a]) {
				a++
			}
			for b > a && unicode.IsSpace(r[b-1]) {
				b--
			}
			if a < b {
				out = append(out, seg{l, a, b})
			}
		}
		return out
	}
	caret = clampPos(lines, caret)
	r := []rune(lines[caret.Line])
	a, b := caret.Col, caret.Col
	for a > 0 && inWord(r, a-1) {
		a--
	}
	for b < len(r) && inWord(r, b) {
		b++
	}
	return []seg{{caret.Line, a, b}}
}

// inWord reports whether r[i] belongs to a word: a letter or a digit, or an
// underscore between two of them (snake_case is one word).
func inWord(r []rune, i int) bool {
	return isWordRune(r[i]) || r[i] == '_' && i > 0 && i+1 < len(r) && isWordRune(r[i-1]) && isWordRune(r[i+1])
}

// enclosingSpan is the innermost span of the line that holds [a, b) between its
// marks, found by the marks next to it first, then by pairing.
func enclosingSpan(line []rune, m Marker, a, b int) (markerSpan, bool) {
	open, close := []rune(m.Open), []rune(m.Close)
	if a >= len(open) && b+len(close) <= len(line) &&
		string(line[a-len(open):a]) == m.Open && string(line[b:b+len(close)]) == m.Close &&
		!snakeCase(line, m, a-len(open), b) {
		return markerSpan{a - len(open), a, b, b + len(close)}, true
	}
	best, found := markerSpan{}, false
	for _, sp := range markSpans(line, m) {
		if sp.oe <= a && b <= sp.cs && (!found || sp.os > best.os) {
			best, found = sp, true
		}
	}
	return best, found
}

// snakeCase reports whether the underscores of an italic pair at os and cs are the
// inside of words (snake_case_name), which are not emphasis.
func snakeCase(line []rune, m Marker, os, cs int) bool {
	if m.Name != "italic" || m.Open != "_" {
		return false
	}
	return os > 0 && isWordRune(line[os-1]) || cs+1 < len(line) && isWordRune(line[cs+1])
}

// segOn reports whether the segment is marked: inside a span of m, or the marks
// are the first and last thing in it.
func segOn(lines []string, s seg, m Marker) (markerSpan, bool) {
	line := []rune(lines[s.line])
	if sp, ok := enclosingSpan(line, m, s.a, s.b); ok {
		return sp, true
	}
	if s.a < s.b {
		for _, sp := range markSpans(line, m) {
			if sp.os == s.a && sp.ce == s.b {
				return sp, true
			}
		}
	}
	return markerSpan{}, false
}

// InlineActive reports whether the text at the caret (the selection, or the word)
// is marked with m: what a toggle button shows as pressed.
func InlineActive(src string, caret Pos, sel Range, m Marker) bool {
	lines := splitLines(src)
	segs := segments(lines, caret, sel)
	for _, s := range segs {
		if _, ok := segOn(lines, s, m); !ok {
			return false
		}
	}
	return len(segs) > 0
}

// lineEdit replaces del runes at col of a line by ins; open says that an insertion
// is an open mark (a caret standing where it goes ends up after it).
type lineEdit struct {
	col, del int
	ins      string
	open     bool
}

// applyEdits writes the edits of one line into it.
func applyEdits(line []rune, eds []lineEdit) []rune {
	out := append([]rune(nil), line...)
	for i := 0; i < len(eds); i++ {
		for j := i + 1; j < len(eds); j++ {
			if eds[j].col > eds[i].col {
				eds[i], eds[j] = eds[j], eds[i]
			}
		}
	}
	for _, e := range eds {
		out = append(out[:e.col], append([]rune(e.ins), out[e.col+e.del:]...)...)
	}
	return out
}

// How a column is carried over an edit that touches it.
const (
	mapCaret  = iota // after an open mark inserted at it, before a close mark
	mapSelEnd        // after every mark inserted at it
	mapSelBeg        // before every mark inserted at it
)

// mapCol is where column p of a line goes once the edits are written: what is
// before it moves it, what is inserted at it moves it as how says, and a column
// inside a deleted mark goes to where the mark was.
func mapCol(eds []lineEdit, p int, how int) int {
	shift := 0
	for _, e := range eds {
		switch {
		case e.del == 0:
			at := e.col < p || e.col == p && (how == mapSelEnd || how == mapCaret && e.open)
			if at {
				shift += len([]rune(e.ins))
			}
		case e.col+e.del <= p:
			shift += len([]rune(e.ins)) - e.del
		case e.col < p:
			shift += e.col - p
		}
	}
	return p + shift
}

// ToggleInline marks the text at the caret with m, or unmarks it when it is marked
// already, as one change of the deck (so one step of Undo). The selection, or the
// word, stays selected (or the caret where it was in the text) so that pressing the
// button again undoes it.
func ToggleInline(src string, caret Pos, sel Range, m Marker) Change {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	sel = normalize(lines, sel)
	segs := segments(lines, caret, sel)
	if len(segs) == 0 {
		return Change{From: caret, To: caret, Caret: caret}
	}
	all := true
	for _, s := range segs {
		if _, ok := segOn(lines, s, m); !ok {
			all = false
		}
	}
	per := map[int][]lineEdit{}
	zero := false
	for _, s := range segs {
		sp, on := segOn(lines, s, m)
		switch {
		case all && on:
			per[s.line] = append(per[s.line],
				lineEdit{sp.os, sp.oe - sp.os, "", false}, lineEdit{sp.cs, sp.ce - sp.cs, "", false})
		case !all && !on && s.a == s.b:
			per[s.line] = append(per[s.line], lineEdit{s.a, 0, m.Open + m.Close, false})
			zero = true
		case !all && !on:
			per[s.line] = append(per[s.line], lineEdit{s.a, 0, m.Open, true}, lineEdit{s.b, 0, m.Close, false})
		}
	}
	first, last := segs[0].line, segs[len(segs)-1].line
	out := make([]string, 0, last-first+1)
	for l := first; l <= last; l++ {
		out = append(out, string(applyEdits([]rune(lines[l]), per[l])))
	}
	text := strings.Join(out, "\n")
	c := Change{From: Pos{first, 0}, To: Pos{last, len([]rune(lines[last]))}, Text: text}
	switch {
	case zero:
		c.Caret = Pos{caret.Line, caret.Col + len([]rune(m.Open))}
	case !sel.Empty():
		// the selection keeps what it covered, marks included
		a := Pos{sel.Start.Line, mapCol(per[sel.Start.Line], sel.Start.Col, mapSelBeg)}
		b := Pos{sel.End.Line, mapCol(per[sel.End.Line], sel.End.Col, mapSelEnd)}
		c.Select = &Range{a, b}
		c.Caret = b
	default:
		c.Caret = Pos{caret.Line, mapCol(per[caret.Line], caret.Col, mapCaret)}
	}
	return c
}
