package playground

import (
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/toolkit"
)

// dragDivider presses divider j, moves by dx and releases, as the page's mouse
// events do.
func dragDivider(t *testing.T, s *Scene, j, dx int) {
	t.Helper()
	r, ok := s.Rects()["split-"+string(rune('0'+j))]
	if !ok {
		t.Fatalf("no divider %d: %v", j, s.Rects())
	}
	x, y := r.X+r.W/2, r.Y+r.H/2
	s.Click(x, y)
	s.Move(x+dx/2, y)
	s.Move(x+dx, y)
	s.Release(x+dx, y)
}

func storedUI(vm *ViewModel) map[string]string {
	if p := vm.drafts.st.Cached(UIProject); p != nil {
		return p.UI
	}
	return nil
}

// The panes start at the widths #39 fixed, and a divider changes them live: the
// pane next to it takes what it gives, the rest keep their place.
func TestDividersResizeThePanesAndTheNeighbourAbsorbs(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	if w := s.files.box.Bounds().W; w != 190 {
		t.Fatalf("the file panel starts at %d, want 190", w)
	}
	if w, hw, th := s.files.box.Bounds().W, s.help.Bounds().W+0, s.thumbs.Bounds().W; th != 168 || w != 190 || hw < 100 {
		t.Fatalf("defaults: files %d thumbs %d help %d", w, th, hw)
	}
	if got := float64(s.editor.Bounds().W+0) / float64(s.PreviewRect().W); got < 1.1 || got > 1.3 {
		t.Fatalf("editor:preview starts near 6:5, got %.2f", got)
	}
	editor, preview, thumbs := s.editor.Bounds().W, s.PreviewRect().W, s.thumbs.Bounds().W
	dragDivider(t, s, 0, 60)
	if w := s.files.box.Bounds().W; w != 250 || vm.FilesWidth.Get() != 250 {
		t.Fatalf("after dragging by 60 the panel is %d (view-model %d), want 250", w, vm.FilesWidth.Get())
	}
	if e := s.editor.Bounds().W; e != editor-60 {
		t.Fatalf("the editor gave the 60 px: %d, was %d", e, editor)
	}
	if s.PreviewRect().W != preview || s.thumbs.Bounds().W != thumbs {
		t.Fatal("the other panes must keep their size")
	}
	// the preview's frame follows a drag of its own divider, live
	before := s.PreviewRect()
	r := s.Rects()["split-2"] // thumbnails | preview
	x, y := r.X+r.W/2, r.Y+r.H/2
	s.Click(x, y)
	s.Move(x+30, y)
	if got := s.PreviewRect(); got.X != before.X+30 || got.W != before.W-30 {
		t.Fatalf("mid-drag the preview is at %+v, was %+v", got, before)
	}
	s.Release(x+30, y)
	if vm.ThumbsWidth.Get() != 168+30 {
		t.Fatalf("the thumbnails are %d wide, want 198", vm.ThumbsWidth.Get())
	}
}

// What the person dragged is stored, what they did not is not; a new page reads
// it back and starts with it.
func TestDraggedSizesAreStoredAndRestored(t *testing.T) {
	SetupText(1)
	m := store.NewMemory()
	vm, h := open(t, m, "")
	s := NewScene(vm, h, 1920, 900)
	if _, ok := storedUI(vm)["w-files"]; ok {
		t.Fatal("a default width was stored")
	}
	dragDivider(t, s, 0, 40)
	ui := storedUI(vm)
	if ui["w-files"] != "230" {
		t.Fatalf("the dragged width is stored as %q, want 230", ui["w-files"])
	}
	for _, k := range []string{"w-thumbs", "w-help"} {
		if _, ok := ui[k]; ok {
			t.Errorf("%s was never changed and must not be stored", k)
		}
	}
	// the editor gave its 40 px, so its share changed with the panel's width:
	// both are kept, and a restart at this window size is the same page
	sharedByDrag := ui["w-split"]
	if sharedByDrag == "" || sharedByDrag == "545" {
		t.Fatalf("the share the drag left is stored: %q", sharedByDrag)
	}
	vm.Thumbs.Set(false) // so that the editor and the preview are neighbours
	dragDivider(t, s, 1, -80)
	if v := storedUI(vm)["w-split"]; v == "" || v == sharedByDrag {
		t.Fatalf("the share is stored again once changed: %q then %q", sharedByDrag, v)
	}
	share := vm.EditorShare()

	vm2, h2 := open(t, m, "")
	s2 := NewScene(vm2, h2, 1920, 900)
	if vm2.FilesWidth.Get() != 230 || s2.files.box.Bounds().W != 230 {
		t.Fatalf("restored files width %d (panel %d), want 230", vm2.FilesWidth.Get(), s2.files.box.Bounds().W)
	}
	if got := vm2.EditorShare(); got != share {
		t.Fatalf("restored share %d, want %d", got, share)
	}
	if vm2.ThumbsWidth.Get() != 168 || vm2.HelpWidth.Get() != 280 {
		t.Fatal("the widths the person did not change stay at their defaults")
	}
	// the stored width wins over a later default; a width out of range is ignored
	p := h2.st.Cached(UIProject)
	p.SetUI("w-help", "5000")
	p.SetUI("w-thumbs", "-3")
	p.SetUI("w-split", "999")
	h2.st.Save(p, vm2.drafts.now(), func(error) {})
	vm3, _ := open(t, m, "")
	if vm3.HelpWidth.Get() != 280 || vm3.ThumbsWidth.Get() != 168 || vm3.EditorShare() != 545 {
		t.Fatalf("out-of-range sizes must be ignored: help %d thumbs %d share %d", vm3.HelpWidth.Get(), vm3.ThumbsWidth.Get(), vm3.EditorShare())
	}
}

// A pane pulled shut by its divider is stored as 0, and the thumbnails' frame
// is hidden with it.
func TestCollapsedPaneIsStoredAndHidesItsFrame(t *testing.T) {
	SetupText(1)
	m := store.NewMemory()
	vm, h := open(t, m, "")
	s := NewScene(vm, h, 1920, 900)
	vm.Source.Set("# a\n\n---\n\n# b\n\n---\n\n# c")
	if _, ok := s.ThumbsRect(); !ok {
		t.Fatal("thumbnails are on")
	}
	dragDivider(t, s, 2, -300) // pull the thumbnails shut
	if vm.ThumbsWidth.Get() != 0 || s.thumbs.Bounds().W != 0 {
		t.Fatalf("thumbnails should be shut: %d / %d", vm.ThumbsWidth.Get(), s.thumbs.Bounds().W)
	}
	if _, ok := s.ThumbsRect(); ok {
		t.Fatal("a shut column has no frame over it")
	}
	if storedUI(vm)["w-thumbs"] != "0" {
		t.Fatalf("stored %q, want 0", storedUI(vm)["w-thumbs"])
	}
	// the View tab's toggles still switch the panes
	vm.FilesPane.Set(false)
	if s.files.box.Bounds().W != 0 {
		t.Fatal("the file panel switched off takes no room")
	}
	vm.FilesPane.Set(true)
	if s.files.box.Bounds().W != 190 {
		t.Fatalf("the file panel is back at %d, want 190", s.files.box.Bounds().W)
	}
	// and a restart keeps the pulled-shut column
	vm2, h2 := open(t, m, "")
	s2 := NewScene(vm2, h2, 1920, 900)
	if s2.thumbs.Bounds().W != 0 {
		t.Fatal("restored shut")
	}
}

// F6 walks the dividers, the arrows move the one that holds the keyboard, Enter
// pulls its pane shut and brings it back, Escape returns to the editor.
func TestDividersWorkFromTheKeyboard(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	s.Click(s.editor.Bounds().X+20, s.editor.Bounds().Y+20)
	if !s.editing() {
		t.Fatal("the editor holds the keyboard")
	}
	s.KeyDown("F6")
	if s.split.FocusedHandle() != 0 || s.editing() {
		t.Fatalf("F6 goes to the first divider: %d, editing %v", s.split.FocusedHandle(), s.editing())
	}
	s.KeyDown("ArrowRight")
	if vm.FilesWidth.Get() != 190+toolkit.SplitterStep {
		t.Fatalf("ArrowRight: %d", vm.FilesWidth.Get())
	}
	if storedUI(vm)["w-files"] == "" {
		t.Fatal("a key press is a choice and is stored")
	}
	s.KeyDown("Tab")
	if s.split.FocusedHandle() != 1 {
		t.Fatalf("Tab goes to the next divider: %d", s.split.FocusedHandle())
	}
	s.KeyDown("Shift+Tab")
	s.KeyDown("Enter")
	if vm.FilesWidth.Get() != 0 {
		t.Fatalf("Enter shuts the panel: %d", vm.FilesWidth.Get())
	}
	s.KeyDown("Enter")
	if vm.FilesWidth.Get() != 190+toolkit.SplitterStep {
		t.Fatalf("Enter brings it back to what it was: %d", vm.FilesWidth.Get())
	}
	s.KeyDown("Escape")
	if s.split.FocusedHandle() != -1 || !s.editing() {
		t.Fatal("Escape returns to the editor")
	}
	// F6 all the way round: every divider that is on, then the editor
	n := 0
	for i := 0; i < 10; i++ {
		s.KeyDown("F6")
		if s.split.FocusedHandle() < 0 {
			break
		}
		n++
	}
	if n != s.split.Handles()-0 || !s.editing() {
		t.Fatalf("F6 visited %d dividers (of %d) then the editor: editing %v", n, s.split.Handles(), s.editing())
	}
	// Shift+F6 goes the other way; a click elsewhere takes the keyboard off
	s.KeyDown("Shift+F6")
	if s.split.FocusedHandle() != s.split.Handles()-1 {
		t.Fatalf("Shift+F6 starts at the last divider: %d", s.split.FocusedHandle())
	}
	s.Click(s.editor.Bounds().X+20, s.editor.Bounds().Y+20)
	if s.split.FocusedHandle() != -1 {
		t.Fatal("a press elsewhere takes the keyboard off the divider")
	}
	// Alt hands the keyboard to the ribbon
	s.KeyDown("F6")
	s.KeyDown("Alt")
	if s.split.FocusedHandle() != -1 {
		t.Fatal("Alt takes the keyboard for the ribbon")
	}
}

func TestResizeCursorOverADivider(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	r := s.Rects()["split-0"]
	if c := s.Cursor(r.X+r.W/2, r.Y+40); c != "col-resize" {
		t.Fatalf("over a divider: %q", c)
	}
	if c := s.Cursor(s.editor.Bounds().X+30, 300); c != "" {
		t.Fatalf("over the editor: %q", c)
	}
	// a hover paints the divider (no panic, no stuck state)
	s.Move(r.X+r.W/2, r.Y+40)
	buf := make([]byte, 1920*900*4)
	s.Draw(buf)
}

// A pane's text wraps to the width a divider gave it.
func TestHelpRewrapsWhenItsDividerMoves(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	buf := make([]byte, 1920*900*4)
	s.Draw(buf)
	wide := s.helpW
	dragDivider(t, s, 3, 150) // the help's divider moves right: it gets narrower
	s.Draw(buf)
	if s.help.Bounds().W >= wide || s.helpW != s.help.Bounds().W {
		t.Fatalf("the help is %d wide, wrapped at %d, was %d", s.help.Bounds().W, s.helpW, wide)
	}
}
