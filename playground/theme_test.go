package playground

import (
	"slices"
	"testing"

	"github.com/go-widgets/toolkit"
)

func TestDeckTheme(t *testing.T) {
	for src, want := range map[string]string{
		"---\nslides: true\ntheme: gaia\n---\n# x":          "gaia",
		"---\ntheme: \"uncover\" # a comment\n---\n":        "uncover",
		"---\ndata:\n  theme: nested\n---\n":                "",
		"---hcl\nslides = true\ntheme = \"gaia\" // c\n---": "gaia",
		"---hcl\ntheme == x\n---":                           "",
		"# no front matter\ntheme: gaia":                    "",
		"---\ntheme: gaia\n":                                "", // never closed
	} {
		if got, _ := DeckTheme(src); got != want {
			t.Errorf("%q: %q, want %q", src, got, want)
		}
	}
}

func TestSetTheme(t *testing.T) {
	for _, c := range []struct{ src, name, want string }{
		// replaced, the comment kept
		{"---\nslides: true\ntheme: gaia # mine\n---\n# x", "uncover", "---\nslides: true\ntheme: uncover # mine\n---\n# x"},
		// added first in the front matter
		{"---\nslides: true\n---\n# x", "gaia", "---\ntheme: gaia\nslides: true\n---\n# x"},
		// a nested theme key is not the theme
		{"---\ndata:\n  theme: x\n---\n", "gaia", "---\ntheme: gaia\ndata:\n  theme: x\n---\n"},
		// HCL
		{"---hcl\nslides = true\ntheme = \"default\"\n---\n", "gaia", "---hcl\nslides = true\ntheme = \"gaia\"\n---\n"},
		{"---hcl\nslides = true\n---\n", "gaia", "---hcl\ntheme = \"gaia\"\nslides = true\n---\n"},
		// no front matter: one is added
		{"# x", "gaia", "---\ntheme: gaia\n---\n\n# x"},
		// a name that is not a name changes nothing
		{"---\ntheme: gaia\n---\n", "x\nmarp: true", "---\ntheme: gaia\n---\n"},
		{"---\ntheme: gaia\n---\n", `a" + file("/etc/passwd") + "`, "---\ntheme: gaia\n---\n"},
	} {
		if got := SetTheme(c.src, c.name); got != c.want {
			t.Errorf("SetTheme(%q, %q):\n got %q\nwant %q", c.src, c.name, got, c.want)
		}
		if themeName.MatchString(c.name) {
			if got, _ := DeckTheme(SetTheme(c.src, c.name)); got != c.name {
				t.Errorf("round trip %q: %q", c.name, got)
			}
		}
	}
}

// TestThemePicker: the picker follows the deck's text, and writes it.
func TestThemePicker(t *testing.T) {
	tr := &fakeTransport{}
	vm := NewViewModel(tr, newHost(), "")
	tr.vm = vm
	if vm.Themes[vm.Theme.Get()] != "default" {
		t.Fatal("the welcome deck's theme")
	}
	// chosen: written into the deck
	vm.Theme.Set(1)
	if got, _ := DeckTheme(vm.Source.Get()); got != "gaia" {
		t.Fatalf("chose gaia, the deck says %q", got)
	}
	// edited by hand: the picker follows, with the deck's own theme
	vm.Source.Set("---\ntheme: corp\n---\n<style>/* @theme corp */</style>\n# x")
	if vm.Themes[vm.Theme.Get()] != "corp" || vm.ThemesVersion.Get() == 0 {
		t.Fatalf("%v %d", vm.Themes, vm.Theme.Get())
	}
	vm.Source.Set("---hcl\ntheme = \"uncover\"\n---\n# x")
	if vm.Themes[vm.Theme.Get()] != "uncover" || len(vm.Themes) != 3 {
		t.Fatalf("%v %d", vm.Themes, vm.Theme.Get())
	}
	vm.Theme.Set(0)
	if vm.Source.Get() != "---hcl\ntheme = \"default\"\n---\n# x" {
		t.Fatalf("%q", vm.Source.Get())
	}
	// a deck that names no theme shows default; choosing it again writes nothing
	vm.Source.Set("# x")
	vm.Theme.Set(0)
	if vm.Source.Get() != "# x" {
		t.Fatalf("%q", vm.Source.Get())
	}
}

func TestPreviewClip(t *testing.T) {
	frame := toolkit.Rect{X: 1000, Y: 200, W: 1000, H: 800}
	if c := PreviewClip(frame, toolkit.Rect{X: 100, Y: 180, W: 200, H: 100}, 2); c != "none" {
		t.Fatal("a list left of the frame clips it:", c)
	}
	// the theme list over the frame's top-left: a notch, in CSS pixels
	c := PreviewClip(frame, toolkit.Rect{X: 1100, Y: 180, W: 200, H: 100}, 2)
	if c != "polygon(0 0,47.0px 0,47.0px 43.0px,153.0px 43.0px,153.0px 0,100% 0,100% 100%,0 100%)" {
		t.Fatal(c)
	}
}

func TestThemeChoices(t *testing.T) {
	src := "---\ntheme: mine\n---\n<style>\n/* @theme corp */\n</style>\n<style>/* @theme corp */</style>\n/* @theme bad\"name */"
	if got := ThemeChoices(src); !slices.Equal(got, []string{"default", "gaia", "uncover", "corp", "bad", "mine"}) {
		t.Fatal(got)
	}
	if got := ThemeChoices("# x"); !slices.Equal(got, BuiltinThemes) {
		t.Fatal(got)
	}
}
