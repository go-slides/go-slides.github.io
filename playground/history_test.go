package playground

import (
	"strings"
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// historyScene is a page over a store, with its scene: the editor the undo
// works on.
func historyScene(t *testing.T) (*Scene, *ViewModel, *filesTab) {
	t.Helper()
	SetupText(1)
	m := store.NewMemory()
	h := &filesTab{tab: &tab{fakeHost: newHost(), answer: true}, saved: map[string][]byte{}}
	h.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, "")
	tr.vm = vm
	answer(tr, vm)
	s := NewScene(vm, h, 1280, 800)
	return s, vm, h
}

// typeAt puts the caret at the start of the editor's text and types s.
func typeAt(t *testing.T, s *Scene, text string) {
	t.Helper()
	s.editor.Focused().Set(true)
	s.editor.CursorLine().Set(0)
	s.editor.CursorCol().Set(0)
	for _, r := range text {
		if !s.Char(string(r)) {
			t.Fatal("the editor took no character")
		}
	}
}

func TestUndoKeepsEachFileItsHistory(t *testing.T) {
	s, vm, _ := historyScene(t)
	deck0 := vm.Source.Get()
	if vm.CanUndo() || vm.CanRedo() || vm.Undo.CanExecute() {
		t.Fatal("something to undo at the start")
	}
	typeAt(t, s, "D")
	if !strings.HasPrefix(vm.Source.Get(), "D") || !vm.Undo.CanExecute() {
		t.Fatal("typing in the deck gave nothing to undo")
	}
	if err := vm.CreateFile("notes.md", "notes\n"); err != nil {
		t.Fatal(err)
	}
	if vm.Open.Get() != "notes.md" || vm.CanUndo() {
		t.Fatal("the new file shows the deck's history")
	}
	typeAt(t, s, "N")
	if vm.FileText.Get() != "Nnotes\n" {
		t.Fatalf("notes: %q", vm.FileText.Get())
	}
	// back to the deck: its own step, not the notes'
	vm.OpenFile(DeckFile)
	vm.Undo.Execute()
	if vm.Source.Get() != deck0 {
		t.Fatal("Undo in the deck did not take the deck's edit back")
	}
	if p := vm.project(); p.Text("notes.md") != "Nnotes\n" {
		t.Fatalf("Undo in the deck changed the notes: %q", p.Text("notes.md"))
	}
	if vm.CanUndo() {
		t.Fatal("the deck has another step")
	}
	// the notes kept theirs
	vm.OpenFile("notes.md")
	if !vm.CanUndo() {
		t.Fatal("the notes lost their history")
	}
	vm.Undo.Execute()
	if vm.FileText.Get() != "notes\n" {
		t.Fatalf("notes after Undo: %q", vm.FileText.Get())
	}
	vm.Redo.Execute()
	if vm.FileText.Get() != "Nnotes\n" {
		t.Fatalf("notes after Redo: %q", vm.FileText.Get())
	}
}

func TestUndoTakesAWholeProjectStepBack(t *testing.T) {
	s, vm, _ := historyScene(t)
	typeAt(t, s, "A")
	typed := vm.Source.Get()
	vm.Example.Set(1)
	if vm.Source.Get() != Examples[1].Source {
		t.Fatal("the example did not load")
	}
	want := T(KStepExample, ExampleNames[1])
	if vm.UndoLabel.Get() != want || UndoTip(vm.UndoLabel.Get()) != T(KTipUndoStep, want) {
		t.Fatalf("label %q", vm.UndoLabel.Get())
	}
	// the deck's typing history went with the deck it was typed in
	vm.Undo.Execute()
	if vm.Source.Get() != typed {
		t.Fatal("Undo did not bring the deck back")
	}
	if !strings.Contains(vm.Status.Get(), want) {
		t.Fatalf("status %q", vm.Status.Get())
	}
	if vm.Example.Get() != 1 {
		t.Fatal("undoing ran the example's own load")
	}
	// ... and its history came back with it
	vm.Undo.Execute()
	if strings.HasPrefix(vm.Source.Get(), "A") {
		t.Fatal("the typing before the example was not undone")
	}
	vm.Redo.Execute() // the typing
	vm.Redo.Execute() // the example
	if vm.Source.Get() != Examples[1].Source || vm.RedoLabel.Get() != "" {
		t.Fatal("Redo did not bring the example back")
	}
	// a new edit drops what could be redone
	vm.Undo.Execute()
	typeAt(t, s, "B")
	if vm.CanRedo() {
		t.Fatal("an edit kept the project's redo")
	}
}

func TestUndoBringsADeletedFileBack(t *testing.T) {
	_, vm, _ := historyScene(t)
	if err := vm.CreateFile("kinds/part.yaml", "part:\n  class: lead\n"); err != nil {
		t.Fatal(err)
	}
	if err := vm.RenamePath("kinds/part.yaml", "kinds/chapter.yaml"); err != nil {
		t.Fatal(err)
	}
	if err := vm.DeletePath("kinds"); err != nil {
		t.Fatal(err)
	}
	if _, ok := vm.project().File("kinds/chapter.yaml"); ok {
		t.Fatal("not deleted")
	}
	vm.Undo.Execute()
	if vm.project().Text("kinds/chapter.yaml") != "part:\n  class: lead\n" {
		t.Fatal("Undo did not bring the file back")
	}
	vm.Undo.Execute()
	if _, ok := vm.project().File("kinds/part.yaml"); !ok {
		t.Fatal("Undo did not take the rename back")
	}
	// a refused operation records nothing
	before := vm.undo.project.UndoLabel()
	if err := vm.RenamePath(DeckFile, "x.md"); err == nil || vm.undo.project.UndoLabel() != before {
		t.Fatal("a refused rename was recorded")
	}
}

func TestUndoShortcuts(t *testing.T) {
	s, vm, _ := historyScene(t)
	typeAt(t, s, "Z")
	if !s.Shortcut("KeyZ", false, false) || strings.HasPrefix(vm.Source.Get(), "Z") {
		t.Fatal("Ctrl+Z")
	}
	if !s.Shortcut("KeyZ", true, false) || !strings.HasPrefix(vm.Source.Get(), "Z") {
		t.Fatal("Ctrl+Shift+Z")
	}
	s.Shortcut("KeyZ", false, false)
	if !s.Shortcut("KeyY", false, false) || !strings.HasPrefix(vm.Source.Get(), "Z") {
		t.Fatal("Ctrl+Y")
	}
	if u, r := isUndoKey("KeyY", true); u || r {
		t.Fatal("Ctrl+Shift+Y is not Redo")
	}
	// Alt with Z is not Undo, and the ribbon has both buttons
	if s.Shortcut("KeyZ", false, true) && vm.Format.Shortcut("KeyZ", false, true) {
		t.Fatal("Ctrl+Alt+Z")
	}
	if RibbonLabels["undo"] == "" || RibbonLabels["redo"] == "" {
		t.Fatal("no Undo/Redo labels")
	}
	if RedoTip("") != T(KTipRedo) || RedoTip("x") != T(KTipRedoStep, "x") || UndoTip("") != T(KTipUndo) {
		t.Fatal("tips")
	}
}

func TestProjectStateSame(t *testing.T) {
	a := ProjectState{Files: map[string][]byte{"a": []byte("1")}, Dirs: []string{"d"}}
	b := ProjectState{Files: map[string][]byte{"a": []byte("1")}, Dirs: []string{"d"}, Open: "a"}
	if !a.Same(b) || a.Size() != 1 {
		t.Fatal("same")
	}
	b.Files["a"] = []byte("2")
	if a.Same(b) {
		t.Fatal("different files")
	}
}
