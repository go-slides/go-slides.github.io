package playground

import (
	"testing"
	"unicode/utf8"
)

// FuzzFormat runs every Format command over arbitrary text, caret and selection:
// the deck is the person's, so none may panic or write a change that Replace cannot
// apply, whatever the text is (unbalanced marks, ragged tables, a caret past the end,
// a selection backwards, invalid UTF-8). The seeds are the shapes the commands read.
func FuzzFormat(f *testing.F) {
	for _, s := range []string{
		"say **hello** _there_ ~~x~~ `c` <u>u</u> and more",
		"| a | b |\n| :-: | --: |\n| 1 | 2 |\n| 3 |",
		"---\ntheme: x\n---\n\n# T\n\n- a\n1. b\n> q\n\n```\ncode **x**\n```\n\n---\n\nnext",
		"a | b\n- | -\n1 | 2",
		"<style scoped>\nsection { text-align: center; }\n</style>\n\n# T",
		"[link](http://x) ![img](a.png) [broken](",
		"***bold italic*** __b__ *i* \\*esc\\* snake_case_name",
		"", "\n", "|", "| |\n|-|", "\xff\xfe", "é*ñ**ü~~",
	} {
		f.Add(s, 0, 0, 0, 0)
		f.Add(s, 1, 2, 3, 4)
		f.Add(s, 2, 5, 0, 1)
	}
	// found by fuzzing: a link over a selection that starts on one line and ends on
	// another read the text from the caret's line
	f.Add("one\ntwo three", 0, 0, 1, 3)
	f.Add("a b\n[c](d)", 1, 1, 0, 0)
	f.Fuzz(func(t *testing.T, src string, l1, c1, l2, c2 int) {
		if len(src) > 4000 || !utf8.ValidString(src) {
			return
		}
		caret := Pos{l1, c1}
		sel := Range{Pos{l1, c1}, Pos{l2, c2}}
		apply := func(c Change) {
			_ = Replace(src, c)
		}
		for _, m := range []Marker{MarkBold, MarkItalic, MarkStrike, MarkCode, MarkUnderline} {
			apply(ToggleInline(src, caret, sel, m))
			apply(ToggleInline(src, caret, Range{}, m))
			InlineActive(src, caret, sel, m)
		}
		for n := 1; n <= 3; n++ {
			c, _ := ToggleHeading(src, caret, sel, n)
			apply(c)
		}
		for _, numbered := range []bool{false, true} {
			c, _ := ToggleList(src, caret, sel, numbered)
			apply(c)
		}
		c, _ := ToggleQuote(src, caret, sel)
		apply(c)
		BlockState(src, caret)
		apply(InsertRule(src, caret, sel))
		for _, op := range []func(string, Pos) (Change, bool){
			func(s string, p Pos) (Change, bool) { return TableAddRow(s, p, true) },
			func(s string, p Pos) (Change, bool) { return TableAddRow(s, p, false) },
			func(s string, p Pos) (Change, bool) { return TableAddColumn(s, p, true) },
			func(s string, p Pos) (Change, bool) { return TableAddColumn(s, p, false) },
			TableDeleteRow, TableDeleteColumn, TableFormat,
			func(s string, p Pos) (Change, bool) { return TableAlign(s, p, AlignCenter) },
			func(s string, p Pos) (Change, bool) { return TableAlign(s, p, AlignRight) },
		} {
			c, _ := op(src, caret)
			apply(c)
		}
		TableState(src, caret)
		apply(InsertTable(src, caret, sel, l2%20, c2%40))
		for _, a := range []SlideAlignment{SlideLeft, SlideCenter, SlideRight} {
			apply(SetSlideAlignment(src, caret, a))
		}
		SlideAlignmentAt(src, caret)
		apply(InsertLink(src, caret, sel))
		InLink(src, caret)
	})
}
