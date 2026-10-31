package playground

import (
	"strings"
	"testing"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-slides/go-slides.github.io/playground/versions"
)

// promptTab is a page that answers Save version…'s question.
type promptTab struct {
	*filesTab
	name   string
	cancel bool
}

func (p *promptTab) Prompt(_, value string) (string, bool) {
	if p.cancel {
		return "", false
	}
	if p.name == "" {
		return value, true
	}
	return p.name, true
}

// versionScene is a page over a store, with a clock the test moves.
func versionScene(t *testing.T) (*Scene, *ViewModel, *promptTab, *time.Time) {
	t.Helper()
	SetupText(1)
	m := store.NewMemory()
	h := &promptTab{filesTab: &filesTab{tab: &tab{fakeHost: newHost(), answer: true}, saved: map[string][]byte{}}}
	h.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, "")
	tr.vm = vm
	answer(tr, vm)
	clock := time.Date(2026, 10, 31, 14, 0, 0, 0, time.Local)
	vm.drafts.now = func() time.Time { return clock }
	s := NewScene(vm, h, 1600, 900)
	return s, vm, h, &clock
}

func kinds(vm *ViewModel) []versions.Kind {
	var out []versions.Kind
	for _, e := range vm.vers.v.Entries() {
		out = append(out, e.Kind)
	}
	return out
}

func selectKind(t *testing.T, vm *ViewModel, k versions.Kind) versions.Entry {
	t.Helper()
	for i, e := range vm.vers.shown {
		if e.Kind == k {
			vm.VersionSelected.Set(i)
			return e
		}
	}
	t.Fatalf("no %s version in %v", k, kinds(vm))
	return versions.Entry{}
}

func TestVersionsAreKeptAutomaticallyAndByName(t *testing.T) {
	s, vm, h, clock := versionScene(t)
	if vm.vers.v == nil || vm.VersionsNote() != T(KVersionsEmpty) {
		t.Fatal("no versions, or no note for an empty list")
	}
	typeAt(t, s, "A")
	vm.SaveDraft()
	if k := kinds(vm); len(k) != 1 || k[0] != versions.Auto {
		t.Fatalf("after the first save: %v", k)
	}
	// not again so soon, then again after AutoEvery
	typeAt(t, s, "B")
	vm.SaveDraft()
	if len(kinds(vm)) != 1 {
		t.Fatal("an automatic version too soon")
	}
	*clock = clock.Add(AutoEvery)
	typeAt(t, s, "C")
	vm.SaveDraft()
	if len(kinds(vm)) != 2 {
		t.Fatalf("after AutoEvery: %v", kinds(vm))
	}
	// a whole-project step keeps what it replaced (unless the newest
	// version holds it already)
	typeAt(t, s, "D")
	vm.Reset.Execute()
	if k := kinds(vm); k[0] != versions.Before || vm.vers.v.Entries()[0].Name != T(KStepReset) {
		t.Fatalf("Reset: %v", k)
	}
	// Save version…, named by the reader; cancelled: nothing
	h.cancel = true
	vm.SaveVersion.Execute()
	h.cancel, h.name = false, "For Ada"
	n := len(kinds(vm))
	vm.SaveVersion.Execute()
	if len(kinds(vm)) != n+1 || vm.vers.v.Entries()[0].Name != "For Ada" || !vm.VersionsPane.Get() {
		t.Fatalf("Save version…: %v", kinds(vm))
	}
	if !strings.Contains(vm.Status.Get(), "For Ada") {
		t.Fatalf("status %q", vm.Status.Get())
	}
	// the pane lists them, the named ones alone on demand
	if len(vm.VersionItems.Get()) != len(kinds(vm)) || !strings.Contains(vm.VersionItems.Get()[0], "For Ada") {
		t.Fatalf("items %q", vm.VersionItems.Get())
	}
	vm.OnlyNamed.Set(true)
	if len(vm.VersionItems.Get()) != 1 {
		t.Fatalf("only named: %q", vm.VersionItems.Get())
	}
	vm.OnlyNamed.Set(false)
	// a share link keeps one too (here a changed deck: an unchanged one is
	// the newest version already)
	typeAt(t, s, "E")
	vm.Share.Execute()
	if kinds(vm)[0] != versions.Share {
		t.Fatalf("share: %v", kinds(vm))
	}
	r := s.Rects()
	if _, ok := r["versions-list"]; !ok {
		t.Fatal("no rect for the list")
	}
}

func TestCompareRestoreAndUndoTheRestore(t *testing.T) {
	s, vm, h, _ := versionScene(t)
	if err := vm.CreateFile("notes.md", "one\ntwo\n"); err != nil {
		t.Fatal(err)
	}
	vm.OpenFile(DeckFile)
	h.name = "with notes"
	vm.SaveVersion.Execute()
	saved := vm.Source.Get()
	// change the deck and the notes
	typeAt(t, s, "NEW ")
	vm.OpenFile("notes.md")
	typeAt(t, s, "zero\n")
	vm.OpenFile(DeckFile)

	e := selectKind(t, vm, versions.Named)
	if !vm.CompareVersion.CanExecute() {
		t.Fatal("Compare is off with a version picked")
	}
	vm.CompareVersion.Execute()
	c := vm.Compare.Get()
	if c == nil || len(c.Files) != 2 || s.leftStack.Visible().Get() != "compare" {
		t.Fatalf("comparison %+v", c)
	}
	if f := c.Files[0]; f.Path != DeckFile || f.Added != 1 || f.Removed != 1 {
		t.Fatalf("deck %+v", f)
	}
	if f := c.Files[1]; f.Path != "notes.md" || f.Added != 1 || f.Removed != 0 {
		t.Fatalf("notes %+v", f)
	}
	if items := vm.CompareItems(); len(items) != 2 || !strings.Contains(items[1], "notes.md (+1 −0)") {
		t.Fatalf("items %q", items)
	}
	vm.CompareFile.Set(1)
	if s.compare.diff.Changes() != 1 {
		t.Fatalf("the diff shows %d changes", s.compare.diff.Changes())
	}
	if s.compare.diff.SideBySide {
		t.Fatal("the comparison does not start inline")
	}
	vm.CompareSideBySide.Set(true)
	if !s.compare.diff.SideBySide {
		t.Fatal("the layout did not follow")
	}
	if _, ok := s.Rects()["compare-diff"]; !ok {
		t.Fatal("no rect for the diff")
	}
	vm.BackToNow.Execute()
	if vm.Compare.Get() != nil || s.leftStack.Visible().Get() != "editor" {
		t.Fatal("Back to now kept the comparison")
	}

	// preview: the render request carries the version's deck, an edit ends it
	vm.PreviewVersion.Execute()
	if !vm.Previewing() || vm.last.Source != saved {
		t.Fatalf("preview renders %q", vm.last.Source[:min(20, len(vm.last.Source))])
	}
	typeAt(t, s, "x")
	if vm.Previewing() {
		t.Fatal("an edit kept the preview")
	}
	vm.Undo.Execute()

	// restore, then Undo the restore
	now := vm.captureState()
	vm.RestoreVersion.Execute()
	if vm.Source.Get() != saved || vm.project().Text("notes.md") != "one\ntwo\n" {
		t.Fatal("not restored")
	}
	if !strings.Contains(vm.Status.Get(), "with notes") || vm.UndoLabel.Get() != T(KStepRestoreVersion, "with notes") {
		t.Fatalf("status %q label %q", vm.Status.Get(), vm.UndoLabel.Get())
	}
	vm.Undo.Execute()
	if !vm.captureState().Same(now) {
		t.Fatal("Undo did not take the restore back")
	}

	// delete, after asking
	h.answer = false
	vm.VersionSelected.Set(0)
	vm.withSelected(func(x versions.Entry) { e = x })
	vm.DeleteVersion.Execute()
	if _, ok := vm.vers.v.Entry(e.ID); !ok {
		t.Fatal("deleted without a yes")
	}
	h.answer = true
	vm.DeleteVersion.Execute()
	if _, ok := vm.vers.v.Entry(e.ID); ok {
		t.Fatal("not deleted")
	}
}

func TestCompareSameAndBinary(t *testing.T) {
	a := ProjectState{Files: map[string][]byte{DeckFile: []byte("x"), "media/p.png": png}}
	b := ProjectState{Files: map[string][]byte{DeckFile: []byte("x"), "media/p.png": append([]byte(nil), png[:20]...)}}
	if c := Compare(a, a, "t", 1); len(c.Files) != 0 {
		t.Fatal("differences in the same project")
	}
	c := Compare(a, b, "t", 1)
	if len(c.Files) != 1 || !c.Files[0].Binary {
		t.Fatalf("%+v", c.Files)
	}
}

// Forget my drafts leaves nothing of the deck in the store, the versions
// included, also once the page is loaded again (draft.cjs checks the same in
// Chrome).
func TestForgetForgetsTheVersions(t *testing.T) {
	m := store.NewMemory()
	vm, _ := open(t, m, "")
	vm.Source.Set("# mine")
	vm.SaveDraft()
	vm.Reset.Execute()
	vm.Source.Set("# mine again")
	vm.SaveDraft()
	if len(vm.vers.v.Entries()) == 0 {
		t.Fatal("no version to forget")
	}
	vm.ForgetDrafts()
	again, _ := open(t, m, "")
	again.SaveDraft()
	for _, n := range m.Names() {
		if n != UIProject {
			t.Fatalf("after Forget and a reload the store holds %v", m.Names())
		}
	}
}

func TestNoStoreNoVersions(t *testing.T) {
	tr := &fakeTransport{}
	vm := NewViewModel(tr, newHost(), "")
	tr.vm = vm
	answer(tr, vm)
	if vm.vers.v != nil || vm.VersionsNote() != T(KVersionsOff) || vm.SaveVersion.CanExecute() {
		t.Fatal("versions without a store")
	}
	vm.SaveVersion.Execute()
	vm.autoVersion()
	vm.backToNow()
}
