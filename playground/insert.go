package playground

import (
	"strings"
)

// The Insert commands of the ribbon are text edits on the deck's Markdown: a
// slide separator, a block of math, a fenced code block, an image or a
// background, a kind. Each is computed here as a pure Change of the source
// (positions in lines and rune columns, as the editor counts them), so that a
// test can check the result without a widget, and the scene can apply the same
// change through the editor, where it joins the undo history.

// Pos is a position in the deck: a 0-based line and a 0-based rune column.
type Pos struct{ Line, Col int }

// Range is a span of the deck, Start <= End; it is empty when they are equal.
type Range struct{ Start, End Pos }

// Empty reports whether the range selects nothing.
func (r Range) Empty() bool { return r.Start == r.End }

// Change replaces the text from From to To (the same position when it only
// inserts) by Text, and says where the caret goes afterwards.
type Change struct {
	From, To Pos
	Text     string
	Caret    Pos
	// Select, when set, is the range left selected afterwards (the text a toggle
	// marked, a link's url to type over); its end is where the caret is.
	Select *Range
}

// splitLines is the deck's lines, the last one possibly empty.
func splitLines(src string) []string { return strings.Split(src, "\n") }

// clampPos puts p inside the text: its line within the lines, its column within
// the line.
func clampPos(lines []string, p Pos) Pos {
	p.Line = min(max(p.Line, 0), len(lines)-1)
	p.Col = min(max(p.Col, 0), len([]rune(lines[p.Line])))
	return p
}

// advance is the position reached by writing s at from.
func advance(from Pos, s string) Pos {
	k := strings.Count(s, "\n")
	if k == 0 {
		return Pos{from.Line, from.Col + len([]rune(s))}
	}
	return Pos{from.Line + k, len([]rune(s[strings.LastIndexByte(s, '\n')+1:]))}
}

// Replace applies a change to a source.
func Replace(src string, c Change) string {
	lines := splitLines(src)
	from, to := clampPos(lines, c.From), clampPos(lines, c.To)
	head := string([]rune(lines[from.Line])[:from.Col])
	tail := string([]rune(lines[to.Line])[to.Col:])
	return strings.Join(lines[:from.Line], "\n") + prefixNL(from.Line) + head + c.Text + tail + suffixNL(lines, to.Line)
}

func prefixNL(n int) string {
	if n == 0 {
		return ""
	}
	return "\n"
}

func suffixNL(lines []string, last int) string {
	if last >= len(lines)-1 {
		return ""
	}
	return "\n" + strings.Join(lines[last+1:], "\n")
}

// textIn is the text of a range.
func textIn(lines []string, r Range) string {
	a, b := clampPos(lines, r.Start), clampPos(lines, r.End)
	if a.Line == b.Line {
		return string([]rune(lines[a.Line])[a.Col:b.Col])
	}
	out := []string{string([]rune(lines[a.Line])[a.Col:])}
	out = append(out, lines[a.Line+1:b.Line]...)
	out = append(out, string([]rune(lines[b.Line])[:b.Col]))
	return strings.Join(out, "\n")
}

// separator reports whether a line is a slide separator.
func separator(line string) bool { return strings.TrimRight(line, " \t") == "---" }

// slideBounds is the span of the slide the line is in: its first line and the
// line of the separator that ends it (len(lines) for the last slide). The front
// matter is not a slide, and a separator inside a fenced code block is not one.
func slideBounds(lines []string, line int) (start, end int) {
	first := 0
	if len(lines) > 0 && separator(lines[0]) {
		first = len(lines) // no closing line: all front matter
		for j := 1; j < len(lines); j++ {
			if separator(lines[j]) || strings.TrimRight(lines[j], " \t") == "..." {
				first = j + 1
				break
			}
		}
	}
	start, fenced := first, false
	for i := first; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fenced = !fenced
		}
		if fenced || !separator(lines[i]) {
			continue
		}
		if line <= i {
			return start, i
		}
		start = i + 1
	}
	return start, len(lines)
}

// InsertSlide adds an empty slide after the one the caret is in: a separator
// and a blank line at the end of that slide's text, the caret on the new slide.
func InsertSlide(src string, caret Pos) Change {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	start, end := slideBounds(lines, caret.Line)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	// after the last line of the slide's text; an empty slide has none, and the
	// separator goes where it would start
	// (a blank line before the separator: right under text, "---" would make it
	// a setext heading)
	text := "\n\n---\n\n"
	at := Pos{end, 0}
	if end == start {
		text = "---\n\n"
	} else {
		at = Pos{end - 1, len([]rune(lines[end-1]))}
	}
	return Change{From: at, To: at, Text: text, Caret: advance(at, text)}
}

// padFor is the newlines a block needs around it to stand on lines of its own:
// one before when the line has text before the position, one after when it has
// text after.
func padFor(lines []string, from, to Pos) (before, after string) {
	if strings.TrimSpace(string([]rune(lines[from.Line])[:from.Col])) != "" {
		before = "\n"
	}
	if strings.TrimSpace(string([]rune(lines[to.Line])[to.Col:])) != "" {
		after = "\n"
	}
	return before, after
}

// Block inserts a block that opens and closes with a line (math between $$, code
// between fences): the selection, when there is one, becomes its content;
// otherwise the content is empty and the caret is put on it.
func Block(src string, caret Pos, sel Range, open, close string) Change {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	from, to := caret, caret
	content := ""
	if !sel.Empty() {
		from, to = clampPos(lines, sel.Start), clampPos(lines, sel.End)
		content = textIn(lines, Range{from, to})
	}
	before, after := padFor(lines, from, to)
	head := before + open + "\n"
	text := head + content + "\n" + close + after
	c := Change{From: from, To: to, Text: text}
	if sel.Empty() {
		c.Caret = advance(from, head)
	} else {
		c.Caret = advance(from, before+open+"\n"+content+"\n"+close)
	}
	return c
}

// Line inserts a line of text (an image, a background) on a line of its own at the
// caret, replacing the selection if there is one, the caret after it.
func Line(src string, caret Pos, sel Range, text string) Change {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	from, to := caret, caret
	if !sel.Empty() {
		from, to = clampPos(lines, sel.Start), clampPos(lines, sel.End)
	}
	before, after := padFor(lines, from, to)
	out := before + text + after
	return Change{From: from, To: to, Text: out, Caret: advance(from, before+text)}
}

// KindComment is the comment that gives a slide its kind.
func KindComment(kind string) string { return "<!-- _kind: " + kind + " -->" }

// SetKind gives the slide the caret is in a kind: its first line becomes the
// comment (a kind it already has is replaced), the caret staying on what it was
// on.
func SetKind(src string, caret Pos, kind string) Change {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	start, _ := slideBounds(lines, caret.Line)
	if start >= len(lines) {
		start = len(lines) - 1
	}
	if strings.HasPrefix(strings.TrimSpace(lines[start]), "<!-- _kind:") {
		from, to := Pos{start, 0}, Pos{start, len([]rune(lines[start]))}
		return Change{From: from, To: to, Text: KindComment(kind), Caret: caret}
	}
	at := Pos{start, 0}
	c := Change{From: at, To: at, Text: KindComment(kind) + "\n", Caret: caret}
	if caret.Line >= start {
		c.Caret.Line++
	}
	return c
}

// RetagImage rewrites the alternative text of an image's Markdown, ![alt](url),
// to alt: "bg" makes it a background, "bg left:40%" a split background (Marpit's
// image syntax). Markdown that is not an image is returned as it is.
func RetagImage(md, alt string) string {
	if !strings.HasPrefix(md, "![") {
		return md
	}
	i := strings.Index(md, "](")
	if i < 0 {
		return md
	}
	return "![" + alt + md[i:]
}

// ImageMode is what an image picked for the deck becomes.
type ImageMode string

// The modes of the Insert tab's Image and Background controls.
const (
	ModeImage      ImageMode = ""             // an image on the slide
	ModeBackground ImageMode = "bg"           // a background
	ModeBgLeft     ImageMode = "bg left:40%"  // a background filling the left 40%
	ModeBgRight    ImageMode = "bg right:40%" // a background filling the right 40%
)
