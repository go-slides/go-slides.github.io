package playground

import (
	"strings"
	"testing"

	"github.com/go-slides/slides/editor"
)

// The Format commands write Markdown; these tests render what they write with the
// engine the playground uses (go-slides' editor.Render, the default HTML
// allow-list) and look at the HTML, so that "underline works with raw HTML off" and
// "a scoped style centres the slide" are facts and not hopes.

func rendered(t *testing.T, src string) string {
	t.Helper()
	r := editor.Render(editor.Request{Path: DeckPath, Source: src})
	if r.HTML == "" {
		t.Fatalf("no output for %q: %v", src, r.Diagnostics)
	}
	return strings.Join(r.SlideHTML, "")
}

// renderedHead is the style sheet of a deck's preview.
func renderedHead(t *testing.T, src string) string {
	t.Helper()
	return editor.Render(editor.Request{Path: DeckPath, Source: src}).Head
}

func TestFormattedMarkdownRenders(t *testing.T) {
	apply := func(in string, f func(src string, caret Pos, sel Range) Change) string {
		src, caret, sel := parseMarked(in)
		return Replace(src, f(src, caret, sel))
	}
	mark := func(m Marker) func(string, Pos, Range) Change {
		return func(s string, c Pos, r Range) Change { return ToggleInline(s, c, r, m) }
	}
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"bold", apply("say «hi» now", mark(MarkBold)), "<strong>hi</strong>"},
		{"italic", apply("say «hi» now", mark(MarkItalic)), "<em>hi</em>"},
		{"strike", apply("say «hi» now", mark(MarkStrike)), "<s>hi</s>"},
		{"code", apply("say «hi» now", mark(MarkCode)), "<code>hi</code>"},
		// underline has no Markdown: the <u> tag passes the default HTML allow-list
		{"underline", apply("say «hi» now", mark(MarkUnderline)), "<u>hi</u>"},
	}
	for _, c := range cases {
		if html := rendered(t, c.src); !strings.Contains(html, c.want) {
			t.Errorf("%s: %q renders without %s", c.name, c.src, c.want)
		}
	}
	block := func(f func(string, Pos, Range) (Change, bool), in string) string {
		src, caret, sel := parseMarked(in)
		c, _ := f(src, caret, sel)
		return Replace(src, c)
	}
	h2 := func(s string, c Pos, r Range) (Change, bool) { return ToggleHeading(s, c, r, 2) }
	ul := func(s string, c Pos, r Range) (Change, bool) { return ToggleList(s, c, r, false) }
	ol := func(s string, c Pos, r Range) (Change, bool) { return ToggleList(s, c, r, true) }
	q := func(s string, c Pos, r Range) (Change, bool) { return ToggleQuote(s, c, r) }
	for name, c := range map[string]struct{ src, want string }{
		"heading":  {block(h2, "Ti‸tle"), "<h2"},
		"bullets":  {block(ul, "«a\nb»"), "<ul"},
		"numbered": {block(ol, "«a\nb»"), "<ol"},
		"quote":    {block(q, "te‸xt"), "<blockquote"},
		"rule":     {apply("one‸\n\ntwo", func(s string, c Pos, r Range) Change { return InsertRule(s, c, r) }), "<hr"},
	} {
		if html := rendered(t, c.src); !strings.Contains(html, c.want) {
			t.Errorf("%s: %q renders without %s", name, c.src, c.want)
		}
	}
	// the rule stays on one slide: every Markdown thematic break would have made two
	rule := apply("one‸\n\ntwo", func(s string, c Pos, r Range) Change { return InsertRule(s, c, r) })
	if r := editor.Render(editor.Request{Path: DeckPath, Source: rule}); len(r.SlideHTML) != 1 {
		t.Errorf("a rule made %d slides", len(r.SlideHTML))
	}
	for _, md := range []string{"---", "***", "___"} {
		if r := editor.Render(editor.Request{Path: DeckPath, Source: "one\n\n" + md + "\n\ntwo"}); len(r.SlideHTML) != 2 {
			t.Errorf("%s was expected to separate slides in Marp: %d slides", md, len(r.SlideHTML))
		}
	}
}

func TestTableAlignmentRenders(t *testing.T) {
	src, caret, sel := parseMarked("‸")
	table := Replace(src, InsertTable(src, caret, sel, 3, 3))
	html := rendered(t, table)
	if !strings.Contains(html, "<table") || strings.Count(html, "<th>") != 3 {
		t.Fatalf("an inserted table renders as 3 header cells: %s", html)
	}
	// centre the first column from a caret in it
	c, ok := TableAlign(table, Pos{0, 4}, AlignCenter)
	if !ok {
		t.Fatal("no table")
	}
	aligned := Replace(table, c)
	html = rendered(t, aligned)
	if !strings.Contains(html, "text-align:center") && !strings.Contains(html, `align="center"`) {
		t.Fatalf("a centred column renders centred: %s", html)
	}
	// each edit keeps a valid table: rows and columns added and removed
	at := Pos{2, 4}
	for _, op := range []func(string, Pos) (Change, bool){
		func(s string, p Pos) (Change, bool) { return TableAddRow(s, p, true) },
		func(s string, p Pos) (Change, bool) { return TableAddColumn(s, p, true) },
		TableDeleteRow, TableDeleteColumn,
	} {
		c, ok := op(aligned, at)
		if !ok {
			t.Fatal("an edit was refused")
		}
		aligned = Replace(aligned, c)
		if !strings.Contains(rendered(t, aligned), "<table") {
			t.Fatalf("not a table any more: %q", aligned)
		}
	}
}

func TestSlideAlignmentRendersInThePreview(t *testing.T) {
	deck := "# One\n\ntext\n\n---\n\n# Two\n\ntext"
	c := SetSlideAlignment(deck, Pos{2, 0}, SlideCenter)
	html := rendered(t, Replace(deck, c))
	head := renderedHead(t, Replace(deck, c))
	if !strings.Contains(head, "text-align: center") && !strings.Contains(head, "text-align:center") {
		t.Fatalf("the scoped style is not in the style sheet")
	}
	if !strings.Contains(html, "data-marpit-scope") {
		t.Fatalf("the slide is not scoped: %s", html)
	}
	if !strings.Contains(html, "<h1") {
		t.Fatal("the slide still renders")
	}
}

func TestHTMLPolicyKeepsTheUnderlineTag(t *testing.T) {
	// the default policy keeps <u> and drops the style attribute: the reason the
	// Format tab underlines with a tag and aligns with a scoped style
	html := rendered(t, `a <u>b</u> and <div style="text-align:center">c</div>`)
	if !strings.Contains(html, "<u>b</u>") {
		t.Error("<u> must pass the allow-list")
	}
	if strings.Contains(html, `style="text-align:center"`) {
		t.Error("a style attribute must not pass the allow-list: alignment needs a scoped style")
	}
}
