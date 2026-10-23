package playground

import (
	"strings"
	"testing"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// TestJournalKeepsTheProjectsFiles: an edit to a theme, journaled when the
// page is hidden but never saved to IndexedDB, comes back at the next load
// with the deck; a file deleted since is deleted again; media stay out of
// the journal.
func TestJournalKeepsTheProjectsFiles(t *testing.T) {
	m, j := store.NewMemory(), &memJournal{}
	vm := openJ(t, m, j, "")
	vm.NewTheme("default")
	vm.NewKind()
	if _, err := vm.AddFile("", "dot.png", png); err != nil {
		t.Fatal(err)
	}
	vm.JournalDeck()
	vm.SaveDraft()
	if j.ok {
		t.Fatal("the journal outlived the save")
	}
	// the last edits: the theme typed in, the kind deleted; journaled, not saved
	time.Sleep(2 * time.Millisecond)
	vm.OpenFile("themes/my-default.css")
	vm.FileText.Set(vm.FileText.Get() + "h2 { color: teal; }\n")
	if err := vm.DeletePath("kinds/chapter.yaml"); err != nil {
		t.Fatal(err)
	}
	vm.JournalDeck()
	if !j.ok || !strings.Contains(j.v, "teal") || strings.Contains(j.v, "dot.png") || strings.Contains(j.v, "chapter.yaml\"") {
		t.Fatalf("journal = %.300s", j.v)
	}
	if p := load(m, DraftProject); strings.Contains(p.Text("themes/my-default.css"), "teal") {
		t.Fatal("the test needs the store to lag")
	}
	vm2 := openJ(t, m, j, "")
	p := vm2.project()
	if !strings.Contains(p.Text("themes/my-default.css"), "teal") {
		t.Fatal("the theme's last edit was lost")
	}
	if _, ok := p.File("kinds/chapter.yaml"); ok {
		t.Fatal("the deleted kind came back")
	}
	if _, ok := p.File("media/dot.png"); !ok {
		t.Fatal("the medium, kept by IndexedDB, was lost")
	}
	if j.ok || !strings.Contains(load(m, DraftProject).Text("themes/my-default.css"), "teal") {
		t.Fatal("the replay was not stored, or the journal not cleared")
	}
	// a hostile entry (a medium, a path out of the project) is dropped whole
	j.v, j.ok = `{"v":2,"source":"# x","files":{"../evil":"x"},"saved":9999999999999}`, true
	openJ(t, m, j, "")
	if j.ok {
		t.Fatal("a journal naming ../evil was kept")
	}
	j.v, j.ok = `{"v":2,"source":"# x","files":{"media/a.png":"x"},"saved":9999999999999}`, true
	openJ(t, m, j, "")
	if j.ok {
		t.Fatal("a journal holding a medium was kept")
	}
	// without IndexedDB the journal is all there is, files included
	j.v, j.ok = `{"v":2,"source":"# alone","files":{"themes/t.css":"/* @theme t */"},"saved":1}`, true
	vm3 := openJ(t, nil, j, "")
	if vm3.Source.Get() != "# alone" || vm3.project().Text("themes/t.css") == "" {
		t.Fatalf("no store: %q", vm3.Source.Get())
	}
}

// TestJournalBudgetCountsTheFiles: the budget is the deck and its files.
func TestJournalBudgetCountsTheFiles(t *testing.T) {
	j := &memJournal{}
	vm := openJ(t, store.NewMemory(), j, "")
	vm.CreateFile("data/big.yaml", strings.Repeat("x", JournalBudget))
	vm.JournalDeck()
	if j.ok || !strings.Contains(vm.Status.Get(), "too large for the quick save") {
		t.Fatalf("journal %v, status %q", j.ok, vm.Status.Get())
	}
}
