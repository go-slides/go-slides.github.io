package playground

import (
	"strings"
	"testing"
)

func TestFilePanelInTheScene(t *testing.T) {
	s, vm, _, _ := newScene(t)
	buf := make([]byte, 4*1280*800)
	s.Draw(buf)
	r := s.Rects()
	tree := r["files"]
	if tree.W < 150 || tree.H < 200 || r["project"].W == 0 || r["usage"].W == 0 {
		t.Fatalf("panel %+v", r)
	}
	if tree.X+tree.W > r["editor"].X {
		t.Fatal("the panel overlaps the editor")
	}
	vm.NewTheme("default")
	if !strings.Contains(s.editor.Text().Get(), "@theme my-default") {
		t.Fatal("the editor does not show the new theme")
	}
	// Typing in it edits the file, not the deck.
	deck := vm.Source.Get()
	s.editor.Text().Set(s.editor.Text().Get() + "\n/* more */\n")
	if !strings.Contains(vm.project().Text("themes/my-default.css"), "/* more */") || vm.Source.Get() != deck {
		t.Fatal("the theme's edit went astray")
	}
	// A click on deck.md (the first file, under the three folders, two of
	// them open) shows the deck again.
	s.Draw(buf)
	rowH := tree.H / 20
	for y := tree.Y + 2; y < tree.Y+tree.H; y += max(rowH/2, 4) {
		s.Click(tree.X+60, y)
		s.Release(tree.X+60, y)
		if vm.Open.Get() == DeckFile {
			break
		}
	}
	if vm.Open.Get() != DeckFile || s.editor.Text().Get() != vm.Source.Get() {
		t.Fatalf("open %q", vm.Open.Get())
	}
	// The keyboard goes to the panel after a click on it.
	s.KeyDown("ArrowUp")
	// The menu, over the panel only.
	if s.Context(r["editor"].X+5, r["editor"].Y+5) {
		t.Fatal("a context menu over the editor")
	}
	if !s.Context(tree.X+40, tree.Y+tree.H-10) || !s.files.tree.MenuOpen() {
		t.Fatal("no context menu over the panel")
	}
	s.Click(r["editor"].X+5, r["editor"].Y+5)
	if s.files.tree.MenuOpen() {
		t.Fatal("a click elsewhere kept the menu")
	}
	// Drops: over the panel, a folder (the root below the rows); elsewhere,
	// the editor's.
	if dir, over := s.FileDropDir(tree.X+40, tree.Y+tree.H-10); !over || dir != "" {
		t.Fatalf("drop dir %q %v", dir, over)
	}
	if _, over := s.FileDropDir(r["editor"].X+5, r["editor"].Y+5); over {
		t.Fatal("the editor is not the panel")
	}
	s.FileDragOver(tree.X+40, tree.Y+5, true)
	if _, on := s.files.tree.DropTarget(); !on {
		t.Fatal("no drop target while dragging")
	}
	s.FileDragOver(0, 0, false)
	if _, on := s.files.tree.DropTarget(); on {
		t.Fatal("the drop target stayed")
	}
	// New file… from the menu starts a name in the panel, in the folder of
	// the selection (the deck's: the top).
	s.files.tree.Selected().Set(DeckFile)
	vm.ProjectAction.Set(6)
	if !s.files.tree.Editing() || !s.files.tree.Focused() {
		t.Fatal("New file… did not start a name")
	}
	for _, c := range "notes.md" {
		s.Char(string(c))
	}
	s.KeyDown("Enter")
	if vm.Open.Get() != "notes.md" {
		t.Fatalf("open %q", vm.Open.Get())
	}
	// A medium inserted from its menu goes into the deck.
	p, _ := vm.AddFile("", "dot.png", png)
	s.InsertImage(MediaMarkdown(p))
	if vm.Open.Get() != DeckFile || !strings.Contains(vm.Source.Get(), "![dot](media/dot.png)") {
		t.Fatal("the medium is not in the deck")
	}
	s.Draw(buf)
}
