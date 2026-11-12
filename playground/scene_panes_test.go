package playground

import (
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit"
)

// paneSwitches are the observables behind each pane, by index.
func paneSwitches(vm *ViewModel) [paneCount]*mvvm.Observable[bool] {
	return [paneCount]*mvvm.Observable[bool]{
		paneFiles: vm.FilesPane, paneEditor: vm.EditorPane, paneThumbs: vm.Thumbs, panePreview: vm.PreviewPane,
		paneHelp: vm.HelpPane, paneVersions: vm.VersionsPane, paneGit: vm.GitPane, paneTheme: vm.ThemePane,
	}
}

// clickClose presses pane i's close button as the page's mouse does.
func clickClose(t *testing.T, s *Scene, i int) {
	t.Helper()
	r := s.panes[i].Button.Bounds()
	if r.W <= 0 {
		t.Fatalf("pane %s has no close button in reach", paneIDs[i])
	}
	s.Click(r.X+r.W/2, r.Y+r.H/2)
	s.Release(r.X+r.W/2, r.Y+r.H/2)
}

// Every pane closes from its button and the others fill the room; the View tab's
// observable brings it back at the width it had.
func TestEveryPaneClosesFromItsButtonAndReopens(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	// every pane is on, the three task panes too
	vm.VersionsPane.Set(true)
	vm.GitPane.Set(true)
	vm.ThemePane.Set(true)
	sw := paneSwitches(vm)
	for i, o := range sw {
		if !o.Get() {
			t.Fatalf("pane %s is not on at the start", paneIDs[i])
		}
	}
	for i := 0; i < paneCount; i++ {
		var before [paneCount]int
		for j := range before {
			before[j] = s.split.PaneSize(j)
		}
		clickClose(t, s, i)
		if sw[i].Get() || s.split.PaneSize(i) != 0 || s.panes[i].Bounds().W != 0 {
			t.Fatalf("%s: still open after its close button (%v, %d wide)", paneIDs[i], sw[i].Get(), s.split.PaneSize(i))
		}
		// the room it left is shared by the panes that stay: the flexible ones
		// (the editor and the preview) grow, none shrinks
		grew := false
		for j := 0; j < paneCount; j++ {
			if j == i {
				continue
			}
			if s.split.PaneSize(j) < before[j] {
				t.Fatalf("closing %s shrank %s: %d -> %d", paneIDs[i], paneIDs[j], before[j], s.split.PaneSize(j))
			}
			grew = grew || s.split.PaneSize(j) > before[j]
		}
		if !grew {
			t.Fatalf("closing %s freed room nobody took", paneIDs[i])
		}
		if _, ok := s.Rects()["close-"+paneIDs[i]]; ok {
			t.Fatalf("%s is closed and still has a close button in Rects", paneIDs[i])
		}
		sw[i].Set(true) // the View tab's toggle
		for j := range before {
			if s.split.PaneSize(j) != before[j] {
				t.Fatalf("reopening %s: pane %s is %d wide, was %d", paneIDs[i], paneIDs[j], s.split.PaneSize(j), before[j])
			}
		}
	}
}

// The editor and the preview can each be closed, never both.
func TestTheLastOfEditorAndPreviewRefusesToClose(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	clickClose(t, s, paneEditor)
	if vm.EditorPane.Get() || !vm.PreviewPane.Get() {
		t.Fatal("the editor did not close")
	}
	btn := s.panes[panePreview].Button
	if !btn.Disabled().Get() || btn.A11y().Name == "" {
		t.Fatal("the last pane's close button must be disabled, and named")
	}
	if s.panes[panePreview].Tip() != T(KTipLastPane) {
		t.Fatalf("the disabled button says %q", s.panes[panePreview].Tip())
	}
	clickClose(t, s, panePreview)
	vm.PreviewPane.Set(false) // a View tab toggle is refused too
	if !vm.PreviewPane.Get() || s.split.PaneSize(panePreview) == 0 {
		t.Fatal("the last pane closed")
	}
	// the editor comes back from the View tab's observable, and then the preview may close
	vm.EditorPane.Set(true)
	if btn.Disabled().Get() {
		t.Fatal("the preview's button stayed disabled")
	}
	clickClose(t, s, panePreview)
	if vm.PreviewPane.Get() || s.PreviewRect().W != 0 {
		t.Fatal("the preview did not close")
	}
}

// Which panes are open is kept: the same page opened again has them as they
// were, at the widths they had.
func TestOpenPanesAreRememberedAcrossAReload(t *testing.T) {
	SetupText(1)
	m := store.NewMemory()
	vm, h := open(t, m, "")
	s := NewScene(vm, h, 1920, 900)
	dragDivider(t, s, 0, 40) // files 230
	clickClose(t, s, panePreview)
	clickClose(t, s, paneThumbs)
	clickClose(t, s, paneFiles)
	ui := storedUI(vm)
	if ui["preview"] != "false" || ui["thumbs"] != "false" || ui["files"] != "false" {
		t.Fatalf("stored panes: %v", ui)
	}
	if _, ok := ui["editor"]; ok {
		t.Fatal("the editor was never chosen and must not be stored")
	}

	vm2, h2 := open(t, m, "")
	s2 := NewScene(vm2, h2, 1920, 900)
	if vm2.PreviewPane.Get() || vm2.Thumbs.Get() || vm2.FilesPane.Get() || !vm2.EditorPane.Get() {
		t.Fatal("a reload did not keep the closed panes closed")
	}
	if s2.PreviewRect().W != 0 || s2.thumbs.Bounds().W != 0 {
		t.Fatal("a pane closed before the reload takes room after it")
	}
	vm2.FilesPane.Set(true)
	if w := s2.files.box.Bounds().W; w != 230 {
		t.Fatalf("the file panel is back at %d, was dragged to 230", w)
	}

	// a stored state with neither the editor nor the preview is not honoured
	p := h2.st.Cached(UIProject)
	p.SetUI("editor", "false")
	p.SetUI("preview", "false")
	h2.st.Save(p, vm2.drafts.now(), func(error) {})
	vm3, h3 := open(t, m, "")
	s3 := NewScene(vm3, h3, 1920, 900)
	if !vm3.EditorPane.Get() || s3.editor.Bounds().W == 0 {
		t.Fatal("a blank page was restored")
	}
}

// The close buttons are stops of the keyboard ring (Tab), Enter closes the pane
// and a closed pane never keeps the keyboard.
func TestCloseButtonsAreReachableByTabAndTakeNoFocusWhenClosed(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	s.Click(s.editor.Bounds().X+20, s.editor.Bounds().Y+20)
	s.KeyDown("F6") // the first divider
	if s.split.FocusedHandle() != 0 {
		t.Fatal("F6 goes to the first divider")
	}
	s.KeyDown("Shift+Tab") // the close button of the file panel, before it
	if !s.panes[paneFiles].Button.Focused() || s.split.FocusedHandle() != -1 {
		t.Fatal("Shift+Tab did not reach the file panel's close button")
	}
	s.KeyDown("ArrowRight") // a button takes no arrow
	s.KeyDown("Enter")
	if vm.FilesPane.Get() || !s.editing() || s.panes[paneFiles].Button.Focused() {
		t.Fatal("Enter closed the pane, and the keyboard goes back to the editor")
	}
	// Tab from the first divider walks every stop that is on, buttons included
	s.KeyDown("F6")
	seen := map[string]bool{}
	for i := 0; i < 40 && !s.editing(); i++ {
		for j, p := range s.panes {
			if p.Button.Focused() {
				seen[paneIDs[j]] = true
			}
		}
		s.KeyDown("Tab")
	}
	if !s.editing() || !seen["thumbs"] || !seen["preview"] || !seen["help"] || seen["files"] {
		t.Fatalf("Tab visited %v then the editor", seen)
	}
	// close the editor with the keyboard: it keeps no focus, typed text goes nowhere
	s.KeyDown("F6")
	s.KeyDown("Shift+Tab") // the editor's button, before the first divider (the file panel is closed)
	if !s.panes[paneEditor].Button.Focused() {
		t.Fatal("not on the editor's close button")
	}
	s.KeyDown("Enter")
	if vm.EditorPane.Get() || s.editing() || s.editor.Focused() {
		t.Fatal("the closed editor kept the keyboard")
	}
	// the preview is now the last pane: its button is not a stop
	for _, st := range s.chromeStops() {
		if st.pane == panePreview {
			t.Fatal("the disabled button is a stop")
		}
	}
	s.KeyDown("F6") // no editor to return to: it still lets go
	s.KeyDown("Escape")
	if s.editor.Focused() {
		t.Fatal("Escape gave the keyboard to the closed editor")
	}
	// Space is Enter on a button, Shift+F6 and F6 from a button reach dividers
	vm.EditorPane.Set(true)
	vm.FilesPane.Set(true)
	s.KeyDown("F6")
	s.KeyDown("Shift+Tab")
	s.KeyDown("F6") // from a button: the next divider
	if s.split.FocusedHandle() != 0 {
		t.Fatalf("F6 from a button goes to the next divider: %d", s.split.FocusedHandle())
	}
	s.KeyDown("Shift+Tab")
	s.KeyDown("Shift+F6") // from a button: the previous divider, none: back to the editor
	if s.split.FocusedHandle() != -1 || !s.editing() {
		t.Fatal("Shift+F6 from the first button returns to the editor")
	}
	s.KeyDown("F6")
	s.KeyDown("Shift+Tab")
	s.chromeKey(0, toolkit.Event{Kind: toolkit.EventKeyDown, Code: "Space"})
	if vm.FilesPane.Get() {
		t.Fatal("Space on a close button did not close the pane")
	}
}

// The canvas host hands over "Tab" and the Shift state apart: SetShift makes the
// first Shift+Tab, and Shift+F6 likewise.
func TestShiftFromTheHostWalksBackwards(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	s.KeyDown("F6")
	s.SetShift(true)
	s.KeyDown("Tab")
	if !s.panes[paneFiles].Button.Focused() {
		t.Fatal("Tab with Shift held did not go back to the close button before the divider")
	}
	s.KeyDown("F6") // Shift+F6 from the first stop: no divider before it, the editor
	s.SetShift(false)
	if s.chromeCurrent() >= 0 || !s.editing() {
		t.Fatal("Shift+F6 from the first stop returns to the editor")
	}
}

// A click anywhere else takes the keyboard off a close button.
func TestAClickElsewhereLeavesTheCloseButton(t *testing.T) {
	SetupText(1)
	vm, h := open(t, store.NewMemory(), "")
	s := NewScene(vm, h, 1920, 900)
	s.KeyDown("F6")
	s.KeyDown("Shift+Tab")
	if !s.panes[paneFiles].Button.Focused() {
		t.Fatal("setup")
	}
	s.Click(s.editor.Bounds().X+20, s.editor.Bounds().Y+20)
	if s.chromeCurrent() >= 0 {
		t.Fatal("a click in the editor left the close button focused")
	}
	// the tooltip of a hovered button is drawn over the panes
	r := s.panes[paneEditor].Button.Bounds()
	s.Release(0, 0)
	s.Move(r.X+r.W/2, r.Y+r.H/2)
	s.AnimationStep(1)
	buf := make([]byte, 1920*900*4)
	s.Draw(buf)
	if _, ok := s.OpenPopover(); !ok {
		t.Fatal("a hovered close button shows no tooltip")
	}
	_ = vm
}
