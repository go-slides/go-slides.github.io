package playground

import (
	"strings"
	"testing"

	"github.com/go-widgets/toolkit"
)

func newScene(t *testing.T) (*Scene, *ViewModel, *fakeTransport, *fakeHost) {
	t.Helper()
	SetupText(1)
	tr := &fakeTransport{}
	h := newHost()
	vm := NewViewModel(tr, h, "")
	answer(tr, vm)
	s := NewScene(vm, h, 1280, 800)
	// the tests drive every pane; a window this narrow starts without two
	vm.FilesPane.Set(true)
	vm.HelpPane.Set(true)
	return s, vm, tr, h
}

func TestSceneEditsTheDeck(t *testing.T) {
	s, vm, _, _ := newScene(t)
	buf := make([]byte, 4*1280*800)
	s.Draw(buf)
	if s.AnimationStep(0.016) {
		t.Fatal("dirty right after a draw")
	}
	if w, h := s.Size(); w != 1280 || h != 800 {
		t.Fatal(w, h)
	}
	if p := s.PreviewRect(); p.W < 200 || p.H < 300 {
		t.Fatalf("preview %+v", p)
	}

	// Click at the end of the heading line and type.
	lines := strings.Split(vm.Source.Get(), "\n")
	hl := 12
	x, y := s.CaretPixel(hl, len([]rune(lines[hl])))
	s.Click(x+2, y+4)
	s.Release(x+2, y+4)
	if !s.editing() {
		t.Fatal("the editor did not take the keyboard")
	}
	for _, r := range " ok" {
		s.Char(string(r))
	}
	if !strings.Contains(vm.Source.Get(), "# Hello, {{ .name }}! ok") {
		t.Fatalf("source: %q", strings.Split(vm.Source.Get(), "\n")[hl])
	}
	if l, _ := s.EditorCursor(); l != hl || vm.CaretLine.Get() != hl {
		t.Fatal("caret not followed")
	}
	s.KeyDown("Backspace")
	if s.KeyDown("v") {
		t.Fatal("a shortcut was taken from the browser")
	}
	if !s.Paste("PASTED") || !strings.Contains(vm.Source.Get(), "PASTED") {
		t.Fatal("paste")
	}
	if s.Copy() != "" {
		t.Fatal("copy without a selection")
	}
	s.Move(x, y)
	s.Scroll(x, y, 0, 3)
	if s.Context(x, y) {
		t.Fatal("context menu")
	}

	// A template error: highlighted on its line.
	vm.Source.Set("---\nslides: true\n---\n\n# {{ .nope }}\n")
	hs := s.ProblemHighlights()
	if len(hs) != 1 || hs[0].StartLine != 4 || hs[0].EndCol <= hs[0].StartCol {
		t.Fatalf("%+v", hs)
	}
	vm.Source.Set("---\nslides: true\n---\n\n# fine\n")
	if len(s.ProblemHighlights()) != 0 {
		t.Fatal("the highlight stayed")
	}

	// A click in the preview moves the editor to a line.
	s.GoToLine(3)
	if l, c := s.EditorCursor(); l != 3 || c != 0 {
		t.Fatal(l, c)
	}
	s.GoToLine(99)
	if l, _ := s.EditorCursor(); l != 5 {
		t.Fatal(l)
	}
	vm.CaretLine.Set(1)
	if l, _ := s.EditorCursor(); l != 1 {
		t.Fatal("the view-model's caret did not move the editor")
	}

	// Leaving the editor: keys go to the controls, paste is the browser's.
	s.Click(2, 2)
	s.Release(2, 2)
	if s.editing() || s.Paste("x") || s.Copy() != "" {
		t.Fatal("still editing")
	}
	s.Char("a")
	s.Release(0, 0)
}

func TestSceneLayoutAndChrome(t *testing.T) {
	s, vm, _, h := newScene(t)
	if w, hh := s.Resize(900, 600); w != 900 || hh != 600 {
		t.Fatal(w, hh)
	}
	s.SetPalette(DarkPalette)
	if s.Theme().Background != DarkPalette.Background || !s.AnimationStep(0) {
		t.Fatal("palette")
	}
	r := s.Rects()
	for _, k := range []string{"ribbon", "tab-Home", "tab-View", "examples", "theme", "share", "draft", "present", "status", "editor", "preview", "help", "problems"} {
		if r[k].W == 0 {
			t.Errorf("%s has no size", k)
		}
	}
	// The help wraps to its pane.
	vm.Help.Set(strings.Repeat("word ", 200))
	for _, line := range strings.Split(s.help.Text().Get(), "\n") {
		if len(line) > 200 {
			t.Fatal("the help is not wrapped")
		}
	}
	// A documentation link opens its page.
	b := s.links[0].Bounds()
	s.Click(b.X+2, b.Y+b.H/2)
	if h.opened != DocLinks[0].URL {
		t.Fatalf("opened %q", h.opened)
	}
	// The share button.
	sb := r["share"]
	s.Click(sb.X+sb.W/2, sb.Y+sb.H/2)
	if !strings.HasPrefix(h.fragment, "deck=") {
		t.Fatal("share")
	}
	// The View tab's mode drop-down opens a list.
	vt := r["tab-View"]
	s.Click(vt.X+vt.W/2, vt.Y+vt.H/2)
	mb := s.Rects()["mode"]
	if mb.W == 0 {
		t.Fatal("the View tab does not show the preview mode")
	}
	s.Click(mb.X+mb.W/2, mb.Y+mb.H/2)
	if _, ok := s.Rects()["popover"]; !ok {
		t.Fatal("no list")
	}
}

func TestCompletionItems(t *testing.T) {
	_, vm, _, _ := newScene(t)
	items := completionItems(vm.Complete(3, len("paginate: ")))
	if len(items) == 0 || items[0].Kind != toolkit.CompletionEnum {
		t.Fatalf("%+v", items)
	}
}
