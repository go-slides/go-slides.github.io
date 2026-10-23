package playground

import (
	"strings"
	"testing"
)

// The tests write a deck with its caret as ‸ and its selection between « and »,
// and show the result the same way, so that a case reads as before and after.

// parseMarked reads a marked deck: its text, caret and selection.
func parseMarked(s string) (string, Pos, Range) {
	lines := strings.Split(s, "\n")
	var caret Pos
	var sel Range
	hasSel := false
	for i, l := range lines {
		r := []rune(l)
		var out []rune
		for _, c := range r {
			switch c {
			case '‸':
				caret = Pos{i, len(out)}
			case '«':
				sel.Start = Pos{i, len(out)}
				hasSel = true
			case '»':
				sel.End = Pos{i, len(out)}
			default:
				out = append(out, c)
			}
		}
		lines[i] = string(out)
	}
	if hasSel {
		caret = sel.End
	}
	return strings.Join(lines, "\n"), caret, sel
}

// showChange applies a change and writes the result marked: the selection it
// leaves, or else the caret.
func showChange(src string, c Change) string {
	text := Replace(src, c)
	lines := strings.Split(text, "\n")
	put := func(p Pos, mark string, lines []string) {
		r := []rune(lines[p.Line])
		lines[p.Line] = string(r[:p.Col]) + mark + string(r[p.Col:])
	}
	if c.Select != nil {
		put(c.Select.End, "»", lines)
		put(c.Select.Start, "«", lines)
	} else {
		put(c.Caret, "‸", lines)
	}
	return strings.Join(lines, "\n")
}

func inline(m Marker, in string) string {
	src, caret, sel := parseMarked(in)
	return showChange(src, ToggleInline(src, caret, sel, m))
}

func TestInlineToggles(t *testing.T) {
	cases := []struct {
		name string
		m    Marker
		in   string
		want string
	}{
		{"bold the selection", MarkBold, "say «hello» there", "say «**hello**» there"},
		{"bold the word at the caret", MarkBold, "say hel‸lo there", "say **hel‸lo** there"},
		{"bold at the end of a word", MarkBold, "say hello‸ there", "say **hello‸** there"},
		{"unbold the selection", MarkBold, "say «**hello**» there", "say «hello» there"},
		{"unbold a word", MarkBold, "say **hel‸lo** there", "say hel‸lo there"},
		{"unbold the selection inside the marks", MarkBold, "say **«hello»** there", "say «hello» there"},
		{"unbold anywhere in a bold phrase", MarkBold, "say **two wo‸rds** here", "say two wo‸rds here"},
		{"bold in the gap makes an empty pair", MarkBold, "say ‸ there", "say **‸** there"},
		{"an empty pair is removed again", MarkBold, "say **‸** there", "say ‸ there"},
		{"italic is an underscore pair", MarkItalic, "a «b» c", "a «_b_» c"},
		{"italic removes the star kind too", MarkItalic, "a *b‸* c", "a b‸ c"},
		{"snake_case is not italic", MarkItalic, "use snake_ca‸se_name now", "use _snake_ca‸se_name_ now"},
		{"strike", MarkStrike, "a ‸word b", "a ~~‸word~~ b"},
		{"unstrike", MarkStrike, "a ~~wo‸rd~~ b", "a wo‸rd b"},
		{"code", MarkCode, "call «Open» now", "call «`Open`» now"},
		{"uncode", MarkCode, "call `Op‸en` now", "call Op‸en now"},
		{"underline is a tag", MarkUnderline, "a «b» c", "a «<u>b</u>» c"},
		{"un-underline", MarkUnderline, "a <u>b‸</u> c", "a b‸ c"},
		{"bold inside italic", MarkBold, "_a «b» c_", "_a «**b**» c_"},
		{"italic inside bold is not confused", MarkItalic, "**a «b» c**", "**a «_b_» c**"},
		{"unbold in bold-italic", MarkBold, "***a‸***", "*a‸*"},
		{"unitalic in bold-italic", MarkItalic, "***a‸***", "**a‸**"},
		{"a mark in code does not count", MarkBold, "`a **b**` x ‸y", "`a **b**` x **‸y**"},
		{"an escaped star is not a mark", MarkBold, "a \\**b ‸c\\** d", "a \\**b **‸c**\\** d"},
		{"bold a heading's text, not its #", MarkBold, "«## Title»", "«## **Title**»"},
		{"bold a list item's text, not its dash", MarkBold, "«- item one»", "«- **item one**»"},
		{"each line of a selection", MarkBold, "«one\ntwo»", "«**one**\n**two**»"},
		{"blank lines are skipped", MarkBold, "«one\n\ntwo»", "«**one**\n\n**two**»"},
		{"unbold all lines", MarkBold, "«**one**\n**two**»", "«one\ntwo»"},
		{"a line unmarked makes the toggle mark them all", MarkBold, "«**one**\ntwo»", "«**one**\n**two**»"},
		{"trailing space stays out", MarkBold, "«hello »x", "«**hello** »x"},
		{"unicode", MarkBold, "un «café» noir", "un «**café**» noir"},
	}
	for _, c := range cases {
		if got := inline(c.m, c.in); got != c.want {
			t.Errorf("%s:\n  in   %q\n  got  %q\n  want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestInlineToggleTwiceIsTheIdentity(t *testing.T) {
	for _, m := range []Marker{MarkBold, MarkItalic, MarkStrike, MarkCode, MarkUnderline} {
		for _, in := range []string{"say «hello» there", "say hel‸lo there", "«one\ntwo»", "- «item»", "a b ‸"} {
			src, caret, sel := parseMarked(in)
			c := ToggleInline(src, caret, sel, m)
			text := Replace(src, c)
			sel2 := Range{}
			caret2 := c.Caret
			if c.Select != nil {
				sel2 = *c.Select
			}
			c2 := ToggleInline(text, caret2, sel2, m)
			if got := Replace(text, c2); got != src {
				t.Errorf("%s twice on %q: %q", m.Name, in, got)
			}
		}
	}
}

func TestInlineState(t *testing.T) {
	cases := []struct {
		m    Marker
		in   string
		want bool
	}{
		{MarkBold, "a **b‸c** d", true},
		{MarkBold, "a **bc** ‸d", false},
		{MarkBold, "a **b c** d‸", false},
		{MarkBold, "a **b‸ c** d", true},
		{MarkBold, "a «**bc**» d", true},
		{MarkBold, "a **b** «c» d", false},
		{MarkItalic, "a _b‸_ c", true},
		{MarkItalic, "a *b‸* c", true},
		{MarkUnderline, "<u>a‸</u>", true},
		{MarkCode, "`a‸`", true},
		{MarkStrike, "~~a‸~~", true},
		{MarkBold, "‸", false},
		{MarkBold, "«one\n**two**»", false},
		{MarkBold, "«**one**\n**two**»", true},
	}
	for _, c := range cases {
		src, caret, sel := parseMarked(c.in)
		if got := InlineActive(src, caret, sel, c.m); got != c.want {
			t.Errorf("%s active at %q: %v, want %v", c.m.Name, c.in, got, c.want)
		}
	}
}

func TestNothingToMark(t *testing.T) {
	c := ToggleInline("", Pos{}, Range{}, MarkBold)
	if got := Replace("", c); got != "****" || c.Caret != (Pos{0, 2}) {
		t.Fatalf("an empty deck gets an empty pair: %q caret %v", got, c.Caret)
	}
	// a selection of blanks has no segment: the caret's own place is marked
	src, caret, sel := parseMarked("a«  »b")
	if got := showChange(src, ToggleInline(src, caret, sel, MarkBold)); got == "" {
		t.Fatal("unreachable")
	}
}

func block(f func(string, Pos, Range) (Change, bool), in string) (string, bool) {
	src, caret, sel := parseMarked(in)
	c, ok := f(src, caret, sel)
	return showChange(src, c), ok
}

func TestHeadings(t *testing.T) {
	h := func(n int) func(string, Pos, Range) (Change, bool) {
		return func(s string, c Pos, r Range) (Change, bool) { return ToggleHeading(s, c, r, n) }
	}
	cases := []struct {
		n        int
		in, want string
	}{
		{1, "Ti‸tle", "# Ti‸tle"},
		{2, "Ti‸tle", "## Ti‸tle"},
		{1, "# Ti‸tle", "Ti‸tle"},
		{1, "## Ti‸tle", "# Ti‸tle"},
		{3, "- it‸em", "### it‸em"},
		{1, "‸", "# ‸"},
		{1, "> quote ‸text", "> # quote ‸text"},
		{1, "«a\n\nb»", "«# a\n\n# b»"},
		{1, "«# a\n# b»", "«a\nb»"},
		{1, "«# a\nb»", "«# a\n# b»"},
		{2, "x\n---\n‸y", "x\n---\n## ‸y"},
	}
	for _, c := range cases {
		got, ok := block(h(c.n), c.in)
		if !ok || got != c.want {
			t.Errorf("H%d on %q: got %q (%v), want %q", c.n, c.in, got, ok, c.want)
		}
	}
	// not in the front matter, a separator or a code block
	for _, in := range []string{"---\ntheme: ‸x\n---\n\n# a", "a\n‸---\nb", "```\nco‸de\n```"} {
		if _, ok := block(h(1), in); ok {
			t.Errorf("a heading must not go on %q", in)
		}
	}
}

func TestLists(t *testing.T) {
	bullets := func(s string, c Pos, r Range) (Change, bool) { return ToggleList(s, c, r, false) }
	numbered := func(s string, c Pos, r Range) (Change, bool) { return ToggleList(s, c, r, true) }
	cases := []struct {
		f        func(string, Pos, Range) (Change, bool)
		in, want string
	}{
		{bullets, "it‸em", "- it‸em"},
		{bullets, "- it‸em", "it‸em"},
		{bullets, "1. it‸em", "- it‸em"},
		{bullets, "# hea‸ding", "- hea‸ding"},
		{bullets, "«a\nb»", "«- a\n- b»"},
		{bullets, "«- a\n- b»", "«a\nb»"},
		{bullets, "«- a\nb»", "«- a\n- b»"},
		{bullets, "  it‸em", "  - it‸em"},
		{numbered, "«a\nb\nc»", "«1. a\n2. b\n3. c»"},
		{numbered, "«1. a\n2. b»", "«a\nb»"},
		{numbered, "- it‸em", "1. it‸em"},
		{numbered, "«a\n\nb»", "«1. a\n\n2. b»"},
		{bullets, "> q‸", "> - q‸"},
	}
	for _, c := range cases {
		if got, ok := block(c.f, c.in); !ok || got != c.want {
			t.Errorf("list on %q: got %q (%v), want %q", c.in, got, ok, c.want)
		}
	}
}

func TestQuote(t *testing.T) {
	q := func(s string, c Pos, r Range) (Change, bool) { return ToggleQuote(s, c, r) }
	cases := []struct{ in, want string }{
		{"te‸xt", "> te‸xt"},
		{"> te‸xt", "te‸xt"},
		{">> te‸xt", "> te‸xt"},
		{"«a\nb»", "«> a\n> b»"},
		{"«> a\n> b»", "«a\nb»"},
		{"- ‸it", "> - ‸it"},
		{"«a\n> b»", "«> a\n> > b»"},
	}
	for _, c := range cases {
		if got, ok := block(q, c.in); !ok || got != c.want {
			t.Errorf("quote on %q: got %q (%v), want %q", c.in, got, ok, c.want)
		}
	}
}

func TestBlockState(t *testing.T) {
	cases := []struct {
		in                    string
		h                     int
		bullets, numbered, qu bool
	}{
		{"## a‸", 2, false, false, false},
		{"- a‸", 0, true, false, false},
		{"3. a‸", 0, false, true, false},
		{"> - a‸", 0, true, false, true},
		{"a‸", 0, false, false, false},
		{"```\n# a‸\n```", 0, false, false, false},
	}
	for _, c := range cases {
		src, caret, _ := parseMarked(c.in)
		h, b, n, q := BlockState(src, caret)
		if h != c.h || b != c.bullets || n != c.numbered || q != c.qu {
			t.Errorf("state of %q: %d %v %v %v", c.in, h, b, n, q)
		}
	}
}

func TestRuleIsNotASlideSeparator(t *testing.T) {
	src, caret, sel := parseMarked("text‸")
	got := Replace(src, InsertRule(src, caret, sel))
	if got != "text\n<hr>" && got != "text\n<hr>\n" {
		t.Fatalf("rule: %q", got)
	}
	if strings.Contains(got, "---") || strings.Contains(got, "***") || strings.Contains(got, "___") {
		t.Fatal("a Markdown rule would start a new slide")
	}
}

func table(f func(string, Pos) (Change, bool), in string) (string, bool) {
	src, caret, _ := parseMarked(in)
	c, ok := f(src, caret)
	return showChange(src, c), ok
}

const sample = "| Name | Qty |\n| --- | ---: |\n| ‸apple | 3 |\n| pear | 12 |"

func TestTableFormatAligns(t *testing.T) {
	got, ok := table(TableFormat, sample)
	want := "| Name  | Qty |\n| ----- | --: |\n| ‸apple |   3 |\n| pear  |  12 |"
	// the caret follows its cell: after "apple", which is cell 0 of row 1
	want = "| Name  | Qty |\n| ----- | --: |\n| apple‸ |   3 |\n| pear  |  12 |"
	if !ok || got != want {
		t.Fatalf("format:\n got  %q\n want %q", got, want)
	}
	// a table that is aligned already is left as it is
	src := "| a   | b   |\n| --- | --- |\n| 1   | 2   |"
	if c, ok := TableFormat(src, Pos{2, 3}); !ok || Replace(src, c) != src {
		t.Fatal("an aligned table is unchanged")
	}
}

func TestTableAlignment(t *testing.T) {
	src := "| a | b |\n| --- | --- |\n| 1 | ‸2 |"
	got, _ := table(func(s string, c Pos) (Change, bool) { return TableAlign(s, c, AlignCenter) }, src)
	want := "| a   |  b  |\n| --- | :-: |\n| 1   | 2‸   |"
	want = "| a   |  b  |\n| --- | :-: |\n| 1   |  2‸  |"
	if got != want {
		t.Fatalf("centre:\n got  %q\n want %q", got, want)
	}
	text, _, _ := parseMarked(got)
	if in, a := TableState(text, Pos{2, 10}); !in || a != AlignCenter {
		t.Fatalf("state %v %c", in, a)
	}
	// the same alignment again clears it
	back, _ := TableAlign(text, Pos{2, 10}, AlignCenter)
	if strings.Contains(Replace(text, back), ":") {
		t.Fatalf("toggled off: %q", Replace(text, back))
	}
	for a, mark := range map[byte]string{AlignLeft: ":--", AlignRight: "--:"} {
		c, _ := TableAlign(src2(), Pos{2, 2}, a)
		if !strings.Contains(Replace(src2(), c), mark) {
			t.Errorf("align %c should write %s: %q", a, mark, Replace(src2(), c))
		}
	}
}

func src2() string { return "| a | b |\n| --- | --- |\n| 1 | 2 |" }

func TestTableRowsAndColumns(t *testing.T) {
	cases := []struct {
		name string
		f    func(string, Pos) (Change, bool)
		in   string
		want string
	}{
		{"row below", func(s string, c Pos) (Change, bool) { return TableAddRow(s, c, true) },
			"| a   | b   |\n| --- | --- |\n| 1‸   | 2   |",
			"| a   | b   |\n| --- | --- |\n| 1   | 2   |\n| ‸    |     |"},
		{"row above", func(s string, c Pos) (Change, bool) { return TableAddRow(s, c, false) },
			"| a   | b   |\n| --- | --- |\n| 1‸   | 2   |",
			"| a   | b   |\n| --- | --- |\n| ‸    |     |\n| 1   | 2   |"},
		{"row from the header goes under it", func(s string, c Pos) (Change, bool) { return TableAddRow(s, c, false) },
			"| a‸   | b   |\n| --- | --- |\n| 1   | 2   |",
			"| a   | b   |\n| --- | --- |\n| ‸    |     |\n| 1   | 2   |"},
		{"row from the delimiter", func(s string, c Pos) (Change, bool) { return TableAddRow(s, c, true) },
			"| a   | b   |\n| -‸-- | --- |\n| 1   | 2   |",
			"| a   | b   |\n| --- | --- |\n| ‸    |     |\n| 1   | 2   |"},
		{"delete row", TableDeleteRow,
			"| a   | b   |\n| --- | --- |\n| 1‸   | 2   |\n| 3   | 4   |",
			"| a   | b   |\n| --- | --- |\n| 3‸   | 4   |"},
		{"delete the last row", TableDeleteRow,
			"| a   | b   |\n| --- | --- |\n| 1   | 2   |\n| 3‸   | 4   |",
			"| a   | b   |\n| --- | --- |\n| 1‸   | 2   |"},
		{"column right", func(s string, c Pos) (Change, bool) { return TableAddColumn(s, c, true) },
			"| a‸   | b   |\n| --- | --- |\n| 1   | 2   |",
			"| a   | Column 2 | b   |\n| --- | -------- | --- |\n| 1   |          | 2   |"},
		{"column left", func(s string, c Pos) (Change, bool) { return TableAddColumn(s, c, false) },
			"| a   | b‸   |\n| --- | --- |\n| 1   | 2   |",
			"| a   | Column 2 | b   |\n| --- | -------- | --- |\n| 1   |          | 2   |"},
		{"delete column", TableDeleteColumn,
			"| a   | b‸   |\n| --- | --- |\n| 1   | 2   |",
			"| a‸   |\n| --- |\n| 1   |"},
	}
	for _, c := range cases {
		got, ok := table(c.f, c.in)
		// the caret's own place is checked by position in TestTableCaret; here the
		// text decides
		strip := func(s string) string { return strings.NewReplacer("‸", "", "«", "", "»", "").Replace(s) }
		if !ok || strip(got) != strip(c.want) {
			t.Errorf("%s:\n got  %q (%v)\n want %q", c.name, strip(got), ok, strip(c.want))
		}
	}
	// a header is not a body row, and a table keeps a column
	if _, ok := table(TableDeleteRow, "| a | b |\n| --- | --- |\n| 1 | 2 |"); ok {
		t.Error("the caret's line is not in the table")
	}
	if _, ok := table(TableDeleteRow, "| a‸ | b |\n| --- | --- |\n| 1 | 2 |"); ok {
		t.Error("the header stays")
	}
	if _, ok := table(TableDeleteColumn, "| a‸ |\n| --- |\n| 1 |"); ok {
		t.Error("a table keeps one column")
	}
	if _, ok := table(TableFormat, "just a‸ line"); ok {
		t.Error("not in a table")
	}
}

func TestTableCaretFollowsItsCell(t *testing.T) {
	src, caret, _ := parseMarked("| a | b |\n| --- | --- |\n| 1 | ‸2 |")
	c, _ := TableAddRow(src, caret, true)
	// the new row is the 4th line of the table, the caret in its second cell
	if c.Caret.Line != 3 {
		t.Fatalf("caret line %d, want 3", c.Caret.Line)
	}
	out := strings.Split(Replace(src, c), "\n")
	if !strings.HasPrefix(out[3][c.Caret.Col-1:], "|") && out[3][c.Caret.Col-1] != ' ' {
		t.Fatalf("caret column %d in %q", c.Caret.Col, out[3])
	}
	// a pipe inside a cell is escaped, not a column
	esc := "| a \\| b | c |\n| --- | --- |\n| 1 | ‸2 |"
	s2, c2, _ := parseMarked(esc)
	tb, ok := findTable(splitLines(s2), c2.Line)
	if !ok || len(tb.aligns) != 2 || tb.rows[0][0] != "a \\| b" {
		t.Fatalf("escaped pipe: %+v", tb)
	}
	// a table without outer pipes is read and written with them
	bare := "a | b\n--- | ---\n1 | ‸2"
	s3, c3, _ := parseMarked(bare)
	c4, ok := TableFormat(s3, c3)
	if !ok || !strings.HasPrefix(Replace(s3, c4), "| a   | b   |") {
		t.Fatalf("bare table: %q", Replace(s3, c4))
	}
}

func TestInsertTable(t *testing.T) {
	src, caret, sel := parseMarked("Some text‸")
	c := InsertTable(src, caret, sel, 3, 3)
	got := showChange(src, c)
	want := "Some text\n\n| «Column 1» | Column 2 | Column 3 |\n| -------- | -------- | -------- |\n|          |          |          |\n|          |          |          |"
	if got != want {
		t.Fatalf("table:\n got  %q\n want %q", got, want)
	}
	// in place on an empty line, with the text after it set off by a blank line
	src, caret, sel = parseMarked("‸\nnext")
	got = showChange(src, InsertTable(src, caret, sel, 2, 2))
	if !strings.HasPrefix(got, "| «Column 1» | Column 2 |\n| -------- | -------- |\n|          |          |\n\nnext") {
		t.Fatalf("in place: %q", got)
	}
	// after a paragraph that has more text below it
	src, caret, sel = parseMarked("para‸\nmore")
	got = showChange(src, InsertTable(src, caret, sel, 1, 2))
	if !strings.HasSuffix(got, "|          |\n\nmore") || !strings.HasPrefix(got, "para\n\n| «Column 1» |") {
		t.Fatalf("under text: %q", got)
	}
	// inside a table: after it
	src, caret, sel = parseMarked("| a | b |\n| --- | --- |\n| 1 | ‸2 |")
	got = Replace(src, InsertTable(src, caret, sel, 1, 2))
	if !strings.HasPrefix(got, "| a | b |\n| --- | --- |\n| 1 | 2 |\n\n| Column 1 |") {
		t.Fatalf("inside a table: %q", got)
	}
	// sizes are limited
	c = InsertTable("", Pos{}, Range{}, 99, 999)
	if n := strings.Count(c.Text, "\n"); n != MaxTableRows+1 { // the delimiter row, and the line end
		t.Fatalf("rows: %d lines", n)
	}
	c = InsertTable("", Pos{}, Range{}, 0, 0)
	if !strings.HasPrefix(c.Text, "| Column 1 |") || strings.Count(c.Text, "\n") != 3 {
		t.Fatalf("minimum: %q", c.Text)
	}
	// a selection: the table goes after its last line
	src, caret, sel = parseMarked("«a\nb»\nc")
	got = Replace(src, InsertTable(src, caret, sel, 1, 2))
	if !strings.HasPrefix(got, "a\nb\n\n| Column 1 |") {
		t.Fatalf("after a selection: %q", got)
	}
}

func TestSlideAlignment(t *testing.T) {
	deck := "---\ntheme: default\n---\n\n# One\n\ntext‸\n\n---\n\n# Two"
	src, caret, _ := parseMarked(deck)
	c := SetSlideAlignment(src, caret, SlideCenter)
	got := Replace(src, c)
	want := "---\ntheme: default\n---\n\n<style scoped>\nsection { text-align: center; }\n</style>\n\n# One\n\ntext\n\n---\n\n# Two"
	if got != want {
		t.Fatalf("centre:\n got  %q\n want %q", got, want)
	}
	if c.Caret != (Pos{src2line(got, "text"), 4}) {
		t.Fatalf("the caret stays on its text: %v", c.Caret)
	}
	if SlideAlignmentAt(got, c.Caret) != SlideCenter || SlideAlignmentAt(got, Pos{len(strings.Split(got, "\n")) - 1, 0}) != SlideLeft {
		t.Fatal("the alignment belongs to its slide only")
	}
	// right rewrites the block, the same alignment again clears it
	r := Replace(got, SetSlideAlignment(got, c.Caret, SlideRight))
	if !strings.Contains(r, "text-align: right;") || strings.Contains(r, "center") {
		t.Fatalf("right: %q", r)
	}
	back := Replace(r, SetSlideAlignment(r, c.Caret, SlideRight))
	if back != src {
		t.Fatalf("toggled off:\n got  %q\n want %q", back, src)
	}
	// left on a left slide does nothing; left on a centred one removes the block
	if c := SetSlideAlignment(src, caret, SlideLeft); c.Text != "" || c.From != c.To {
		t.Fatalf("left on left: %+v", c)
	}
	if left := Replace(got, SetSlideAlignment(got, c.Caret, SlideLeft)); left != src {
		t.Fatalf("left on centre: %q", left)
	}
	// another style block the person wrote is left alone
	own := "# T\n\n<style scoped>\nh1 { color: red; }\n</style>\n\nx‸"
	s2, c2, _ := parseMarked(own)
	g2 := Replace(s2, SetSlideAlignment(s2, c2, SlideRight))
	if !strings.Contains(g2, "h1 { color: red; }") || !strings.Contains(g2, "text-align: right;") {
		t.Fatalf("own style: %q", g2)
	}
	// a slide with no text yet and the last line of a deck
	e := Replace("", SetSlideAlignment("", Pos{}, SlideCenter))
	if e != "<style scoped>\nsection { text-align: center; }\n</style>" {
		t.Fatalf("empty deck: %q", e)
	}
	if empty := Replace(e, SetSlideAlignment(e, Pos{}, SlideCenter)); empty != "" {
		t.Fatalf("removing the only block: %q", empty)
	}
}

func src2line(text, prefix string) int {
	for i, l := range strings.Split(text, "\n") {
		if strings.HasPrefix(l, prefix) {
			return i
		}
	}
	return -1
}

func TestLinks(t *testing.T) {
	run := func(in string) string {
		src, caret, sel := parseMarked(in)
		return showChange(src, InsertLink(src, caret, sel))
	}
	if got := run("say «hello» now"); got != "say [hello](«https://») now" {
		t.Errorf("link a selection: %q", got)
	}
	if got := run("say hel‸lo now"); got != "say [hello](«https://») now" {
		t.Errorf("link the word: %q", got)
	}
	if got := run("see «https://example.com» now"); got != "see [«»](https://example.com) now" {
		t.Errorf("a selected url: %q", got)
	}
	if got := run("see ‸ now"); got != "see [«text»](https://) now" {
		t.Errorf("a new link: %q", got)
	}
	if got := run("see [a link](http://x.y‸) now"); got != "see [a link](«http://x.y») now" {
		t.Errorf("edit a link: %q", got)
	}
	if got := run("see ![img](a.png‸) now"); got != "see ![img](a.png‸) now" {
		t.Errorf("an image is not a link: %q", got)
	}
	src, caret, _ := parseMarked("a [b‸](c)")
	if !InLink(src, caret) {
		t.Error("the caret is in a link")
	}
	src, caret, _ = parseMarked("a ‸ [b](c)")
	if InLink(src, caret) {
		t.Error("the caret is not in a link")
	}
}
