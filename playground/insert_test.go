package playground

import (
	"strings"
	"testing"
)

// edit applies a change and returns the new text, checking the caret it names is
// a place in it.
func edit(t *testing.T, src string, c Change) string {
	t.Helper()
	out := Replace(src, c)
	lines := splitLines(out)
	if c.Caret.Line < 0 || c.Caret.Line >= len(lines) || c.Caret.Col < 0 || c.Caret.Col > len([]rune(lines[c.Caret.Line])) {
		t.Errorf("caret %+v is not inside %q", c.Caret, out)
	}
	return out
}

func TestReplaceAndAdvance(t *testing.T) {
	src := "héllo\nworld\nend"
	if got := Replace(src, Change{From: Pos{0, 1}, To: Pos{1, 2}, Text: "X"}); got != "hXrld\nend" {
		t.Errorf("a multi-line replace (runes, not bytes): %q", got)
	}
	if got := Replace(src, Change{From: Pos{2, 3}, To: Pos{2, 3}, Text: "\nmore"}); got != "héllo\nworld\nend\nmore" {
		t.Errorf("an append at the end: %q", got)
	}
	if got := Replace(src, Change{From: Pos{99, 99}, To: Pos{99, 99}, Text: "!"}); got != "héllo\nworld\nend!" {
		t.Errorf("positions past the text clamp to its end: %q", got)
	}
	if got := Replace(src, Change{From: Pos{-3, -3}, To: Pos{0, 0}, Text: ">"}); got != ">héllo\nworld\nend" {
		t.Errorf("negative positions clamp to the start: %q", got)
	}
	if p := advance(Pos{2, 3}, "ab"); p != (Pos{2, 5}) {
		t.Errorf("advance over one line: %+v", p)
	}
	if p := advance(Pos{2, 3}, "ab\ncdé"); p != (Pos{3, 3}) {
		t.Errorf("advance over two lines: %+v", p)
	}
	if (Range{Pos{1, 1}, Pos{1, 1}}).Empty() != true || (Range{Pos{1, 1}, Pos{1, 2}}).Empty() {
		t.Error("Range.Empty")
	}
	lines := splitLines("one\ntwo\nthree")
	if got := textIn(lines, Range{Pos{0, 1}, Pos{0, 3}}); got != "ne" {
		t.Errorf("textIn on a line: %q", got)
	}
	if got := textIn(lines, Range{Pos{0, 1}, Pos{2, 2}}); got != "ne\ntwo\nth" {
		t.Errorf("textIn across lines: %q", got)
	}
}

func TestSlideBounds(t *testing.T) {
	lines := splitLines("---\ntheme: gaia\n---\n# One\n\n---\n# Two\n```\n---\n```\n---\n# Three")
	// 0..2 front matter, slide 1 = 3..5(sep), slide 2 = 6..10(sep; the --- inside the fence is not one), slide 3 = 11..
	cases := []struct{ line, start, end int }{
		{0, 3, 5}, {1, 3, 5}, {3, 3, 5}, {5, 3, 5}, {6, 6, 10}, {8, 6, 10}, {10, 6, 10}, {11, 11, 12}, {99, 11, 12},
	}
	for _, c := range cases {
		if s, e := slideBounds(lines, c.line); s != c.start || e != c.end {
			t.Errorf("slideBounds(line %d) = %d,%d want %d,%d", c.line, s, e, c.start, c.end)
		}
	}
	// no front matter; an unterminated front matter is all front matter
	if s, e := slideBounds(splitLines("a\n---\nb"), 2); s != 2 || e != 3 {
		t.Errorf("no front matter: %d,%d", s, e)
	}
	if s, e := slideBounds(splitLines("---\ntitle: x"), 1); s != 2 || e != 2 {
		t.Errorf("a front matter that never closes: %d,%d", s, e)
	}
	if s, _ := slideBounds(splitLines("---\na: b\n...\nbody"), 3); s != 3 {
		t.Errorf("a front matter closed by ...: %d", s)
	}
	if s, e := slideBounds(nil, 0); s != 0 || e != 0 {
		t.Errorf("no lines: %d,%d", s, e)
	}
}

func TestInsertSlide(t *testing.T) {
	deck := "---\ntheme: gaia\n---\n# One\n\ntext\n\n\n---\n# Two\n"
	// after the first slide's last line of text, before its separator
	got := edit(t, deck, InsertSlide(deck, Pos{4, 0}))
	want := "---\ntheme: gaia\n---\n# One\n\ntext\n\n---\n\n\n\n\n---\n# Two\n"
	if got != want {
		t.Errorf("middle slide:\n got %q\nwant %q", got, want)
	}
	c := InsertSlide(deck, Pos{4, 0})
	if c.Caret != (Pos{9, 0}) {
		t.Errorf("the caret goes onto the new slide: %+v", c.Caret)
	}
	if slides := countSlides(got); slides != 3 {
		t.Errorf("a slide was added: %d slides in %q", slides, got)
	}
	// the last slide, text ends with a newline
	got = edit(t, deck, InsertSlide(deck, Pos{9, 0}))
	if !strings.HasPrefix(got, deck[:len(deck)-1]) || countSlides(got) != 3 {
		t.Errorf("last slide: %q", got)
	}
	// no front matter, no trailing newline
	got = edit(t, "# A", InsertSlide("# A", Pos{0, 3}))
	if got != "# A\n\n---\n\n" {
		t.Errorf("a one-slide deck: %q", got)
	}
	// a slide with nothing in it: the separator goes where it would start
	got = edit(t, "a\n---\n", InsertSlide("a\n---\n", Pos{2, 0}))
	if countSlides(got) != 3 {
		t.Errorf("after an empty last slide: %q", got)
	}
	// an empty deck
	got = edit(t, "", InsertSlide("", Pos{}))
	if got != "---\n\n" {
		t.Errorf("an empty deck: %q", got)
	}
	// a separator inside a fenced block does not end the slide
	fenced := "# A\n```\n---\n```\ntail\n---\n# B"
	got = edit(t, fenced, InsertSlide(fenced, Pos{1, 0}))
	if !strings.Contains(got, "tail\n\n---\n\n\n---\n# B") {
		t.Errorf("fenced separator: %q", got)
	}
}

// countSlides is the number of slides: the separators outside the front matter
// and fences, plus one.
func countSlides(src string) int {
	lines := splitLines(src)
	n := 1
	for s, e := slideBounds(lines, 0); e < len(lines); s, e = slideBounds(lines, e+1) {
		n++
		_ = s
	}
	return n
}

func TestBlock(t *testing.T) {
	// no selection, caret on an empty line: the block with the caret inside
	src := "# T\n\nafter"
	c := Block(src, Pos{1, 0}, Range{}, "$$", "$$")
	got := edit(t, src, c)
	if got != "# T\n$$\n\n$$\nafter" {
		t.Errorf("math at an empty line: %q", got)
	}
	if c.Caret != (Pos{2, 0}) {
		t.Errorf("the caret is inside the block: %+v", c.Caret)
	}
	// the caret in the middle of a line: the block goes on lines of its own
	src = "ab cd"
	got = edit(t, src, Block(src, Pos{0, 2}, Range{}, "```go {lineNumbers}", "```"))
	if got != "ab\n```go {lineNumbers}\n\n```\n cd" {
		t.Errorf("mid-line: %q", got)
	}
	// a selection becomes the content
	src = "x\nfoo\nbar\ny"
	sel := Range{Pos{1, 0}, Pos{2, 3}}
	c = Block(src, Pos{2, 3}, sel, "$$", "$$")
	got = edit(t, src, c)
	if got != "x\n$$\nfoo\nbar\n$$\ny" {
		t.Errorf("selection: %q", got)
	}
	if c.Caret != (Pos{4, 2}) {
		t.Errorf("the caret follows the block: %+v", c.Caret)
	}
	// a selection inside a line
	src = "a b c"
	got = edit(t, src, Block(src, Pos{0, 3}, Range{Pos{0, 2}, Pos{0, 3}}, "```go {lineNumbers}", "```"))
	if got != "a \n```go {lineNumbers}\nb\n```\n c" {
		t.Errorf("selection inside a line: %q", got)
	}
}

func TestLine(t *testing.T) {
	src := "text here\nnext"
	got := edit(t, src, Line(src, Pos{0, 9}, Range{}, "![bg](x)"))
	if got != "text here\n![bg](x)\nnext" {
		t.Errorf("at the end of a line: %q", got)
	}
	got = edit(t, src, Line(src, Pos{1, 0}, Range{}, "![bg](x)"))
	if got != "text here\n![bg](x)next" {
		// text after the caret goes on a line of its own
		if got != "text here\n![bg](x)\nnext" {
			t.Errorf("at the start of a line: %q", got)
		}
	}
	// a selection is replaced
	got = edit(t, "see logo now", Line("see logo now", Pos{0, 8}, Range{Pos{0, 4}, Pos{0, 8}}, "![logo](u)"))
	if !strings.Contains(got, "![logo](u)") || strings.Contains(got, "logo now") && !strings.Contains(got, "\n now") {
		t.Errorf("selection replaced: %q", got)
	}
}

func TestSetKind(t *testing.T) {
	src := "---\ntheme: gaia\n---\n# One\n\n---\n# Two\ntext"
	// the kind goes on the first line of the slide the caret is in
	c := SetKind(src, Pos{7, 2}, "listing")
	got := edit(t, src, c)
	if got != "---\ntheme: gaia\n---\n# One\n\n---\n<!-- _kind: listing -->\n# Two\ntext" {
		t.Errorf("new kind: %q", got)
	}
	if c.Caret != (Pos{8, 2}) {
		t.Errorf("the caret stays on its text: %+v", c.Caret)
	}
	// a kind the slide has is replaced, the caret unmoved
	c = SetKind(got, Pos{8, 2}, "title")
	got2 := edit(t, got, c)
	if got2 != "---\ntheme: gaia\n---\n# One\n\n---\n<!-- _kind: title -->\n# Two\ntext" || c.Caret != (Pos{8, 2}) {
		t.Errorf("replaced kind: %q caret %+v", got2, c.Caret)
	}
	// a caret above the slide start (in the front matter) is moved, not shifted
	c = SetKind(src, Pos{1, 0}, "x")
	if c.Caret != (Pos{1, 0}) {
		t.Errorf("a caret before the slide: %+v", c.Caret)
	}
	// a deck of only front matter
	only := "---\na: b\n---"
	if got := edit(t, only, SetKind(only, Pos{0, 0}, "k")); !strings.Contains(got, "<!-- _kind: k -->") {
		t.Errorf("front matter only: %q", got)
	}
}

func TestRetagImage(t *testing.T) {
	cases := map[[2]string]string{
		{"![logo](data:image/png;base64,AA==)", "bg"}: "![bg](data:image/png;base64,AA==)",
		{"![logo](https://x/y.png)", "bg left:40%"}:   "![bg left:40%](https://x/y.png)",
		{"![](u)", "my picture"}:                      "![my picture](u)",
		{"plain text", "bg"}:                          "plain text",
		{"![unterminated", "bg"}:                      "![unterminated",
	}
	for in, want := range cases {
		if got := RetagImage(in[0], in[1]); got != want {
			t.Errorf("RetagImage(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
	if KindComment("a") != "<!-- _kind: a -->" {
		t.Error("KindComment")
	}
}
