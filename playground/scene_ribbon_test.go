package playground

import (
	"strings"
	"testing"

	"github.com/go-widgets/toolkit"
)

// clickRect clicks the middle of a named rectangle and releases.
func clickRect(s *Scene, name string) toolkit.Rect {
	r := s.Rects()[name]
	s.Click(r.X+r.W/2, r.Y+r.H/2)
	s.Release(r.X+r.W/2, r.Y+r.H/2)
	return r
}

// gotoTab shows a ribbon tab by clicking it.
func gotoTab(t *testing.T, s *Scene, name string) {
	t.Helper()
	if clickRect(s, "tab-"+name).W == 0 {
		t.Fatalf("no tab %s", name)
	}
}

func TestRibbonShowsEachTabsControls(t *testing.T) {
	s, _, _, _ := newScene(t)
	s.Resize(1500, 800)
	want := map[string][]string{
		"Home":   {"examples", "reset", "theme", "present", "share", "draft"},
		"Insert": {"newslide", "kind", "image", "background", "math", "code"},
		"View":   {"mode", "present", "minimapToggle", "thumbsToggle"},
		"Export": {"save-html", "save-pdf", "save-pptx", "save-odp"},
	}
	for tab, keys := range want {
		gotoTab(t, s, tab)
		for _, k := range keys {
			if s.Rects()[k].W == 0 {
				t.Errorf("%s tab: no %s among %v", tab, k, s.ribbon.Levels())
			}
		}
	}
	// A narrow window folds groups into pop-ups but keeps the tabs.
	s.Resize(560, 700)
	gotoTab(t, s, "Home")
	if s.Rects()["tab-Export"].W == 0 {
		t.Error("the tabs go with the width")
	}
}

func TestRibbonCommandsRun(t *testing.T) {
	s, vm, _, h := newScene(t)
	s.Resize(1500, 800)
	clickRect(s, "share")
	if !strings.HasPrefix(h.fragment, "deck=") {
		t.Error("Share link")
	}
	clickRect(s, "present")
	if h.presenters != 1 || vm.Mode.Get() != int(ModeSlideshow) {
		t.Errorf("Present: presenters %d mode %d", h.presenters, vm.Mode.Get())
	}
	gotoTab(t, s, "Export")
	clickRect(s, "save-html")
	if h.downloads["deck.html"] == "" {
		t.Error("HTML page")
	}
	// the View toggles are the view-model's observables
	gotoTab(t, s, "View")
	before := vm.Minimap.Get()
	clickRect(s, "minimapToggle")
	if vm.Minimap.Get() == before || s.editor.ShowMinimap != vm.Minimap.Get() {
		t.Error("the Minimap toggle did not reach the view-model and the editor")
	}
	before = vm.Thumbs.Get()
	clickRect(s, "thumbsToggle")
	if vm.Thumbs.Get() == before {
		t.Error("the Slides toggle did not reach the view-model")
	}
	// the Theme list writes the theme into the deck
	gotoTab(t, s, "Home")
	clickRect(s, "theme")
	if _, ok := s.Rects()["popover"]; !ok {
		t.Fatal("the theme list did not open")
	}
	pop := s.Rects()["popover"]
	s.Click(pop.X+pop.W/2, pop.Y+pop.H-3) // the last theme
	s.Release(pop.X+pop.W/2, pop.Y+pop.H-3)
	if vm.Theme.Get() == 0 {
		t.Error("choosing a theme in the ribbon did not change it")
	}
}

func TestRibbonKeytipsTakeAndGiveBackTheKeyboard(t *testing.T) {
	s, vm, _, _ := newScene(t)
	s.Resize(1500, 800)
	x, y := s.CaretPixel(12, 0)
	s.Click(x+2, y+4)
	s.Release(x+2, y+4)
	if !s.editing() {
		t.Fatal("the editor should hold the keyboard")
	}
	// Alt takes the keyboard for the ribbon; letters are keytips: E is the Export tab.
	s.KeyDown("Alt")
	if s.editing() || !s.ribbon.Focused() {
		t.Fatal("Alt should give the keyboard to the ribbon")
	}
	s.Char("e")
	s.Char("h") // the HTML page's tip
	if vm.Status.Get() == "" || s.ribbon.Focused() || !s.editing() {
		t.Errorf("a keytip that ran a command returns the keyboard to the editor: focused=%v editing=%v", s.ribbon.Focused(), s.editing())
	}
	// Escape on the tabs gives it back too.
	s.KeyDown("Alt")
	s.KeyDown("Escape")
	if s.ribbon.Focused() || !s.editing() {
		t.Error("Escape out of the keytips returns to the editor")
	}
	// Typing in the editor is not stolen by the ribbon.
	before := vm.Source.Get()
	s.Char("z")
	if vm.Source.Get() == before {
		t.Error("the editor stopped receiving text")
	}
}

func TestRibbonInsertEditsThroughTheEditor(t *testing.T) {
	s, vm, _, _ := newScene(t)
	s.Resize(1500, 800)
	vm.Source.Set("# One\n\ntext\n")
	gotoTab(t, s, "Insert")
	// New slide: a separator and a blank slide, the editor holding the keyboard
	slides := strings.Count(vm.Source.Get(), "---")
	clickRect(s, "newslide")
	if strings.Count(vm.Source.Get(), "---") != slides+1 || !s.editing() {
		t.Errorf("New slide: %q editing=%v", vm.Source.Get(), s.editing())
	}
	// ...and one Undo takes it back: the edit is in the editor's history
	s.editor.Undo()
	if strings.Count(vm.Source.Get(), "---") != slides {
		t.Errorf("Undo did not take the slide back: %q", vm.Source.Get())
	}
	// Math around the selection
	vm.Source.Set("x\nfoo\ny")
	s.editor.SetSelection(toolkit.Selection{StartLine: 1, StartCol: 0, EndLine: 1, EndCol: 3})
	clickRect(s, "math")
	if got := vm.Source.Get(); got != "x\n$$\nfoo\n$$\ny" {
		t.Errorf("Math around the selection: %q", got)
	}
	// The code block with line numbers, at the caret
	vm.Source.Set("a")
	s.GoToLine(0)
	clickRect(s, "code")
	if !strings.Contains(vm.Source.Get(), "```go {lineNumbers}") {
		t.Errorf("Code block: %q", vm.Source.Get())
	}
	// Image: the host is asked for a file; with none, the status says so
	clickRect(s, "image")
	if !strings.HasPrefix(vm.Status.Get(), "⚠") {
		t.Errorf("Image without a file picker: %q", vm.Status.Get())
	}
}

func TestRibbonCollapseGivesTheRoomBack(t *testing.T) {
	s, _, _, _ := newScene(t)
	open := s.PreviewRect().H
	s.ribbon.Collapsed().Set(true)
	if s.PreviewRect().H <= open {
		t.Errorf("a collapsed ribbon should leave the preview taller: %d then %d", open, s.PreviewRect().H)
	}
	s.ribbon.Collapsed().Set(false)
	if s.PreviewRect().H != open {
		t.Errorf("expanding gives the room back: %d, want %d", s.PreviewRect().H, open)
	}
}

func TestThemeOptions(t *testing.T) {
	got := ThemeOptions([]string{"default", "gaia"})
	if len(got) != 2 || got[0] != "Theme: default" || got[1] != "Theme: gaia" {
		t.Errorf("%q", got)
	}
}
