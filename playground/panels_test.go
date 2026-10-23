package playground

import (
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// A window picks its panes by its width: at 1280 the editor gets the room (the
// file panel and the help start off), at 1920 everything is on, and the editor
// still has about 60 columns (a monospaced column is about 7.8 px at 13 px).
func TestPanesFollowTheWindowWidth(t *testing.T) {
	SetupText(1)
	cases := []struct {
		w          int
		files, hlp bool
	}{{1280, false, false}, {1500, true, false}, {1920, true, true}}
	for _, c := range cases {
		vm := NewViewModel(&fakeTransport{}, newHost(), "")
		s := NewScene(vm, newHost(), c.w, 800)
		if vm.FilesPane.Get() != c.files || vm.HelpPane.Get() != c.hlp {
			t.Errorf("%d px: files %v help %v, want %v %v", c.w, vm.FilesPane.Get(), vm.HelpPane.Get(), c.files, c.hlp)
		}
		cols := (s.editor.Bounds().W - 60) * 10 / 78 // minus the gutter and the scrollbar
		if cols < 60 {
			t.Errorf("%d px: the editor is %d px wide, %d columns", c.w, s.editor.Bounds().W, cols)
		}
		// the panes that are off take one pixel, not a band
		if !c.files && s.files.box.Bounds().W > 2 {
			t.Errorf("%d px: the file panel off still takes %d px", c.w, s.files.box.Bounds().W)
		}
	}
}

// What the person chose is kept and wins over the width; a default is not kept.
func TestPaneChoicesAreStored(t *testing.T) {
	SetupText(1)
	m := store.NewMemory()
	vm, h := open(t, m, "")
	NewScene(vm, h, 1280, 800)
	if vm.saved["files"] || vm.saved["help"] {
		t.Fatal("a default was taken for a choice")
	}
	vm.FilesPane.Set(true)
	vm.RibbonCollapsed.Set(true)
	p := h.st.Cached(UIProject)
	if p == nil || p.UI["files"] != "true" || p.UI["ribbon"] != "true" {
		t.Fatalf("the choices are not stored: %+v", p)
	}
	if _, ok := p.UI["help"]; ok {
		t.Error("help was never chosen and must not be stored")
	}
	// a new page reads them back: files on and the ribbon collapsed at 1280 too
	vm2, h2 := open(t, m, "")
	s2 := NewScene(vm2, h2, 1280, 800)
	if !vm2.FilesPane.Get() || vm2.HelpPane.Get() || !vm2.RibbonCollapsed.Get() || !s2.ribbon.Collapsed().Get() {
		t.Errorf("restored: files %v help %v ribbon %v", vm2.FilesPane.Get(), vm2.HelpPane.Get(), vm2.RibbonCollapsed.Get())
	}
	// the ribbon's own collapse reaches the view-model and so the store
	s2.ribbon.Collapsed().Set(false)
	if vm2.RibbonCollapsed.Get() || h2.st.Cached(UIProject).UI["ribbon"] != "false" {
		t.Error("expanding the ribbon did not reach the view-model and the store")
	}
}
