package playground

import (
	"strconv"
	"strings"
	"unicode"
)

// The Format tab's paragraph commands: headings, bulleted and numbered lists and a
// quote work on the lines of the selection (or the caret's line) and toggle, as in
// the editors whose toolbars the bibliography read (Markdown All in One, Typora:
// Ctrl+1..3, Ctrl+Shift+[ and ]), and a horizontal rule is the <hr> tag, because a
// Markdown rule (`---`, `***`, `___`) separates slides in Marp.

// linePrefix is a line split into what a block-level mark is made of: the indent,
// the quote marks ("> "), a heading, bullet or number mark, and the rest.
type linePrefix struct {
	indent string
	quote  string
	// kind is 0 for none, 'h' for a heading (level in n), 'b' for a bullet, 'n'
	// for a number.
	kind byte
	n    int
	mark string // as written, with its space
	rest string
}

// parseLine splits a line.
func parseLine(s string) linePrefix {
	var p linePrefix
	r := []rune(s)
	i := 0
	for i < len(r) && (r[i] == ' ' || r[i] == '\t') {
		i++
	}
	p.indent = string(r[:i])
	q := i
	for i < len(r) && r[i] == '>' {
		i++
		if i < len(r) && r[i] == ' ' {
			i++
		}
	}
	p.quote = string(r[q:i])
	rest := r[i:]
	switch {
	case len(rest) > 0 && rest[0] == '#':
		n := runAt(rest, 0, '#')
		if n <= 6 && (n == len(rest) || rest[n] == ' ') {
			p.kind, p.n = 'h', n
			m := n
			if m < len(rest) {
				m++
			}
			p.mark, rest = string(rest[:m]), rest[m:]
		}
	case len(rest) > 1 && (rest[0] == '-' || rest[0] == '*' || rest[0] == '+') && rest[1] == ' ':
		p.kind, p.mark, rest = 'b', string(rest[:2]), rest[2:]
	default:
		d := 0
		for d < len(rest) && d < 9 && rest[d] >= '0' && rest[d] <= '9' {
			d++
		}
		if d > 0 && d+1 < len(rest) && (rest[d] == '.' || rest[d] == ')') && rest[d+1] == ' ' {
			n, _ := strconv.Atoi(string(rest[:d]))
			p.kind, p.n, p.mark, rest = 'n', n, string(rest[:d+2]), rest[d+2:]
		}
	}
	p.rest = string(rest)
	return p
}

func (p linePrefix) String() string { return p.indent + p.quote + p.mark + p.rest }

// prefixLen is the number of runes before the text of the line.
func (p linePrefix) prefixLen() int { return len([]rune(p.indent + p.quote + p.mark)) }

// blockPrefixLen is the number of runes of a line that are its indent and its
// quote, heading or list marks.
func blockPrefixLen(r []rune) int { return parseLine(string(r)).prefixLen() }

// frontMatterEnd is the index of the line after the front matter (0 when the deck
// has none).
func frontMatterEnd(lines []string) int {
	if len(lines) == 0 || !separator(lines[0]) {
		return 0
	}
	for j := 1; j < len(lines); j++ {
		if separator(lines[j]) || strings.TrimRight(lines[j], " \t") == "..." {
			return j + 1
		}
	}
	return len(lines)
}

// plainLines says, for each line, whether it is Markdown text a block mark may be
// put on: not the front matter, a slide separator, a code fence or a line inside
// one.
func plainLines(lines []string) []bool {
	ok := make([]bool, len(lines))
	fm := frontMatterEnd(lines)
	fenced := false
	for i, l := range lines {
		t := strings.TrimSpace(l)
		isFence := strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")
		if isFence {
			fenced = !fenced
		}
		ok[i] = i >= fm && !fenced && !isFence && !separator(l)
	}
	return ok
}

// lineSpan is the first and last line a block command works on: those of the
// selection (not the last one when the selection ends at its start), else the
// caret's.
func lineSpan(lines []string, caret Pos, sel Range) (first, last int) {
	sel = normalize(lines, sel)
	if sel.Empty() {
		c := clampPos(lines, caret)
		return c.Line, c.Line
	}
	first, last = sel.Start.Line, sel.End.Line
	if sel.End.Col == 0 && last > first {
		last--
	}
	return first, last
}

// blockEdit parses the lines a block command works on, lets f change them, and
// returns the change that writes them back. f gets only the lines that are text
// (and are not blank, unless the command is on a single line, so that a heading can
// be started on an empty line).
func blockEdit(src string, caret Pos, sel Range, f func(ls []*linePrefix)) (Change, bool) {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	sel = normalize(lines, sel)
	first, last := lineSpan(lines, caret, sel)
	plain := plainLines(lines)
	var idx []int
	var ls []*linePrefix
	for l := first; l <= last; l++ {
		if !plain[l] || first != last && strings.TrimSpace(lines[l]) == "" {
			continue
		}
		p := parseLine(lines[l])
		idx, ls = append(idx, l), append(ls, &p)
	}
	if len(ls) == 0 {
		return Change{From: caret, To: caret, Caret: caret}, false
	}
	old := make(map[int]int, len(ls))
	for k, l := range idx {
		old[l] = ls[k].prefixLen()
	}
	f(ls)
	out := make([]string, 0, last-first+1)
	delta := map[int]int{}
	k := 0
	for l := first; l <= last; l++ {
		if k < len(idx) && idx[k] == l {
			out = append(out, ls[k].String())
			delta[l] = ls[k].prefixLen() - old[l]
			k++
		} else {
			out = append(out, lines[l])
		}
	}
	// a column in the marks stays at the start of the text; one in the text moves
	// with it
	move := func(p Pos, start bool) Pos {
		d, ok := delta[p.Line]
		if !ok || start && p.Col == 0 { // a selection that began at the line's start still does
			return p
		}
		p.Col = max(p.Col+d, 0)
		if n := blockPrefixLen([]rune(out[p.Line-first])); p.Col < n && d != 0 {
			p.Col = n
		}
		return p
	}
	c := Change{From: Pos{first, 0}, To: Pos{last, len([]rune(lines[last]))}, Text: strings.Join(out, "\n")}
	c.Caret = move(caret, false)
	if !sel.Empty() {
		c.Select = &Range{move(sel.Start, true), move(sel.End, false)}
		c.Caret = c.Select.End
	}
	return c, true
}

// ToggleHeading makes the lines level-n headings, or plain lines when they all are
// already.
func ToggleHeading(src string, caret Pos, sel Range, n int) (Change, bool) {
	return blockEdit(src, caret, sel, func(ls []*linePrefix) {
		all := true
		for _, p := range ls {
			all = all && p.kind == 'h' && p.n == n
		}
		for _, p := range ls {
			if all {
				p.kind, p.n, p.mark = 0, 0, ""
			} else {
				p.kind, p.n, p.mark = 'h', n, strings.Repeat("#", n)+" "
			}
		}
	})
}

// ToggleList makes the lines a bulleted or a numbered list, or plain lines when
// they all are such a list already.
func ToggleList(src string, caret Pos, sel Range, numbered bool) (Change, bool) {
	kind := byte('b')
	if numbered {
		kind = 'n'
	}
	return blockEdit(src, caret, sel, func(ls []*linePrefix) {
		all := true
		for _, p := range ls {
			all = all && p.kind == kind
		}
		for i, p := range ls {
			switch {
			case all:
				p.kind, p.n, p.mark = 0, 0, ""
			case numbered:
				p.kind, p.n, p.mark = 'n', i+1, strconv.Itoa(i+1)+". "
			default:
				p.kind, p.n, p.mark = 'b', 0, "- "
			}
		}
	})
}

// ToggleQuote puts a quote mark on the lines, or takes one level off when they all
// have one.
func ToggleQuote(src string, caret Pos, sel Range) (Change, bool) {
	return blockEdit(src, caret, sel, func(ls []*linePrefix) {
		all := true
		for _, p := range ls {
			all = all && p.quote != ""
		}
		for _, p := range ls {
			if all {
				r := []rune(p.quote)
				i := 1
				if i < len(r) && r[i] == ' ' {
					i++
				}
				p.quote = string(r[i:])
			} else {
				p.quote = "> " + p.quote
			}
		}
	})
}

// BlockState is what the caret's line is: a heading's level (0 for none), whether it
// is in a bulleted or a numbered list, whether it is quoted.
func BlockState(src string, caret Pos) (heading int, bullets, numbered, quote bool) {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	if !plainLines(lines)[caret.Line] {
		return 0, false, false, false
	}
	p := parseLine(lines[caret.Line])
	if p.kind == 'h' {
		heading = p.n
	}
	return heading, p.kind == 'b', p.kind == 'n', p.quote != ""
}

// RuleLine is the horizontal rule the Format tab writes. Not `---`, and not `***` or
// `___` either: Marpit reads every thematic break as a slide separator (rendered
// and checked in format_render_test.go), so a rule inside a slide is the <hr> tag,
// which the default HTML allow-list keeps.
const RuleLine = "<hr>"

// InsertRule puts a horizontal rule on a line of its own at the caret.
func InsertRule(src string, caret Pos, sel Range) Change { return Line(src, caret, sel, RuleLine) }

// isBlank reports whether a line has no text.
func isBlank(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return !unicode.IsSpace(r) }) < 0
}
