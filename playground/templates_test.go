package playground

import (
	"strings"
	"testing"

	"github.com/go-slides/slides/editor"
)

func TestListInConfig(t *testing.T) {
	cases := []struct{ in, key, value, want string }{
		{"", "themes", "themes", ConfigTemplate + "themes: [themes]\n"},
		{"# mine\nthemes: themes\n", "themes", "themes", "# mine\nthemes: themes\n"},
		{"themes: other\n", "themes", "themes", "themes: [other, themes]\n"},
		{"kinds:\n", "kinds", "kinds/a.yaml", "kinds: [kinds/a.yaml]\n"},
		{"kinds:\n  - kinds/a.yaml # first\n", "kinds", "kinds/b.yaml", "kinds:\n  - kinds/a.yaml # first\n  - kinds/b.yaml\n"},
		{"kinds: [kinds/a.yaml]\n", "kinds", "kinds/a.yaml", "kinds: [kinds/a.yaml]\n"},
		{"lang: fr\n", "data", "data/v.yaml", "lang: fr\ndata: [data/v.yaml]\n"},
	}
	for _, c := range cases {
		got, err := ListInConfig(c.in, c.key, c.value)
		if err != nil || got != c.want {
			t.Errorf("ListInConfig(%q, %s, %s) = %q, %v\nwant %q", c.in, c.key, c.value, got, err, c.want)
		}
	}
	for _, bad := range []string{"kinds: {a: {class: x}}\n", "- a\n", "a: [\n"} {
		if _, err := ListInConfig(bad, "kinds", "kinds/x.yaml"); err == nil {
			t.Errorf("%q: no error", bad)
		}
	}
}

func TestRenameInConfig(t *testing.T) {
	src := "themes: themes\nkinds: [kinds/a.yaml, kinds/b.yaml]\ndata: data/v.yaml\nlang: fr\n"
	got, _ := RenameInConfig(src, "kinds/a.yaml", "kinds/z.yaml", false)
	if !strings.Contains(got, "kinds: [kinds/z.yaml, kinds/b.yaml]") {
		t.Fatalf("rename: %q", got)
	}
	got, _ = RenameInConfig(src, "kinds", "styles", false) // a folder
	if !strings.Contains(got, "kinds: [styles/a.yaml, styles/b.yaml]") {
		t.Fatalf("folder: %q", got)
	}
	got, _ = RenameInConfig(src, "data/v.yaml", "", true)
	if !strings.Contains(got, "data: []") {
		t.Fatalf("remove scalar: %q", got)
	}
	got, _ = RenameInConfig(src, "kinds/b.yaml", "", true)
	if !strings.Contains(got, "kinds: [kinds/a.yaml]") {
		t.Fatalf("remove: %q", got)
	}
	if got, _ := RenameInConfig(src, "media/x.png", "media/y.png", false); got != src {
		t.Fatalf("unrelated: %q", got)
	}
	if got, err := RenameInConfig("", "a", "b", false); got != "" || err != nil {
		t.Fatal("empty")
	}
	if _, err := RenameInConfig("a: [\n", "a", "b", false); err == nil {
		t.Fatal("bad YAML")
	}
}

// The templates build, the way the slides command reads a project.
func TestTemplatesRender(t *testing.T) {
	cfg, _ := ListInConfig("", "themes", "themes")
	cfg, _ = ListInConfig(cfg, "kinds", "kinds/chapter.yaml")
	cfg, _ = ListInConfig(cfg, "data", "data/values.yaml")
	for _, base := range BuiltinThemes {
		r := editor.Render(editor.Request{
			Path:   "/p/deck.md",
			Source: "---\ntheme: mine\n---\n\n<!-- _kind: chapter -->\n# {{ .title }}\n",
			Files: map[string]string{
				"/p/slides.yaml": cfg, "/p/themes/mine.css": ThemeTemplate("mine", base),
				"/p/kinds/chapter.yaml": KindTemplate("chapter"), "/p/data/values.yaml": DataTemplate(),
			},
		})
		if len(r.Diagnostics) > 0 || !strings.Contains(r.HTML, "My talk") || !strings.Contains(r.HTML, "invert") {
			t.Fatalf("%s: %+v\n%.300s", base, r.Diagnostics, r.HTML)
		}
	}
}
