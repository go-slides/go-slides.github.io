package playground

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// memJournal is localStorage's journal key, shared by the page loads of a
// browser; it can refuse writes.
type memJournal struct {
	v      string
	ok     bool
	refuse bool
}

// jtab is a tab with a journal.
type jtab struct {
	*tab
	j *memJournal
}

func (t *jtab) JournalWrite(v string) error {
	if t.j.refuse {
		return errors.New("QuotaExceededError")
	}
	t.j.v, t.j.ok = v, true
	return nil
}
func (t *jtab) JournalRead() (string, bool) { return t.j.v, t.j.ok }
func (t *jtab) JournalClear()               { t.j.v, t.j.ok = "", false }

func openJ(t *testing.T, m *store.Memory, j *memJournal, fragment string) *ViewModel {
	t.Helper()
	h := &jtab{tab: &tab{fakeHost: newHost(), answer: true}, j: j}
	if m != nil {
		h.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	}
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, fragment)
	tr.vm = vm
	answer(tr, vm)
	return vm
}

// TestJournalClosesTheGap: an edit journaled but never saved to IndexedDB
// (the page closed before the asynchronous write) comes back at the next
// load, is saved there, and the journal is cleared.
func TestJournalClosesTheGap(t *testing.T) {
	m, j := store.NewMemory(), &memJournal{}
	vm := openJ(t, m, j, "")
	vm.Source.Set("# saved\n")
	vm.JournalDeck()
	vm.SaveDraft()
	if j.ok {
		t.Fatal("the journal outlived the save of its text")
	}
	// the last keystroke: journaled (pagehide), the IndexedDB save never ran
	time.Sleep(2 * time.Millisecond)
	vm.Source.Set("# saved!\n")
	vm.JournalDeck()
	if !j.ok || !strings.Contains(j.v, "# saved!") {
		t.Fatal("not journaled")
	}
	if deckOf(m, DraftProject) != "# saved\n" {
		t.Fatal("the test needs the store to lag")
	}
	vm2 := openJ(t, m, j, "")
	if vm2.Source.Get() != "# saved!\n" || !strings.Contains(vm2.Status.Get(), "Restored your draft") {
		t.Fatalf("%q / %q", vm2.Source.Get(), vm2.Status.Get())
	}
	if deckOf(m, DraftProject) != "# saved!\n" || j.ok {
		t.Fatal("the replay was not stored, or the journal not cleared")
	}
	// the same text is journaled once, not on every tick
	vm2.Source.Set("# again\n")
	vm2.JournalDeck()
	j.v = "sentinel"
	vm2.JournalDeck()
	if j.v != "sentinel" {
		t.Fatal("an unchanged text was written again")
	}
	// Forget clears the journal too, and leaving the page right after does
	// not journal the forgotten deck again
	vm2.ForgetDrafts()
	vm2.JournalDeck()
	if j.ok {
		t.Fatal("Forget left the journal, or leaving wrote it back")
	}
	// a page opened and left untouched journals nothing
	j2 := &memJournal{}
	vm3 := openJ(t, store.NewMemory(), j2, "")
	vm3.JournalDeck()
	if j2.ok {
		t.Fatal("an untouched example was journaled")
	}
}

// TestJournalOlderThanTheDraft: a journal the draft already outran (another
// tab saved since) is dropped, not replayed over it.
func TestJournalOlderThanTheDraft(t *testing.T) {
	m, j := store.NewMemory(), &memJournal{}
	vm := openJ(t, m, j, "")
	vm.Source.Set("# old\n")
	vm.JournalDeck()
	time.Sleep(2 * time.Millisecond)
	vm.Source.Set("# newer\n")
	vm.SaveDraft() // the journal still says "# old" (not written since)
	j.v, j.ok = strings.Replace(j.v, "# newer", "# old", 1), true
	vm2 := openJ(t, m, j, "")
	if vm2.Source.Get() != "# newer\n" || j.ok {
		t.Fatalf("%q, journal kept %v", vm2.Source.Get(), j.ok)
	}
}

// TestJournalBoundsAndRefusal: a deck past the budget is not journaled and
// the status says so once; a refused write is said once; a garbled entry
// is dropped; without IndexedDB the journal alone brings the text back.
func TestJournalBoundsAndRefusal(t *testing.T) {
	m, j := store.NewMemory(), &memJournal{}
	vm := openJ(t, m, j, "")
	vm.Source.Set("# small\n")
	vm.JournalDeck()
	vm.Source.Set(strings.Repeat("x", JournalBudget+1))
	vm.JournalDeck()
	if j.ok || !strings.Contains(vm.Status.Get(), "too large for the quick save") {
		t.Fatalf("over budget: journal %v, %q", j.ok, vm.Status.Get())
	}
	vm.Status.Set("quiet")
	vm.Source.Set(strings.Repeat("y", JournalBudget+1))
	vm.JournalDeck()
	if vm.Status.Get() == "quiet" {
		// a render replaced it: what matters is that the warning is not repeated
	} else if strings.Contains(vm.Status.Get(), "too large") {
		t.Fatal("said twice")
	}

	j2 := &memJournal{refuse: true}
	vm2 := openJ(t, store.NewMemory(), j2, "")
	vm2.Source.Set("# x\n")
	vm2.JournalDeck()
	if !strings.Contains(vm2.Status.Get(), "quick save is refused (QuotaExceededError)") {
		t.Fatal(vm2.Status.Get())
	}

	j3 := &memJournal{v: "{not json", ok: true}
	vm3 := openJ(t, store.NewMemory(), j3, "")
	if vm3.Source.Get() != Examples[0].Source || j3.ok {
		t.Fatal("a garbled journal was used or kept")
	}

	// no IndexedDB: the journal is all there is
	j4 := &memJournal{v: `{"v":1,"source":"# only here\n","saved":1}`, ok: true}
	vm4 := openJ(t, nil, j4, "")
	if vm4.Source.Get() != "# only here\n" || !strings.Contains(vm4.Status.Get(), "Restored your draft") {
		t.Fatalf("%q / %q", vm4.Source.Get(), vm4.Status.Get())
	}
	// a shared link still wins over it
	frag, _ := EncodeShare("# shared\n")
	if vm5 := openJ(t, nil, j4, frag); vm5.Source.Get() != "# shared\n" {
		t.Fatal(vm5.Source.Get())
	}
}

// TestJournalAndSharedLink: a journal replayed while a shared link opens is
// the draft that link sets aside, so Restore my draft brings the last
// keystroke back.
func TestJournalAndSharedLink(t *testing.T) {
	m, j := store.NewMemory(), &memJournal{}
	vm := openJ(t, m, j, "")
	vm.Source.Set("# mine\n")
	vm.SaveDraft()
	time.Sleep(2 * time.Millisecond)
	vm.Source.Set("# mine!\n")
	vm.JournalDeck()
	frag, _ := EncodeShare("# shared\n")
	vm2 := openJ(t, m, j, frag)
	if vm2.Source.Get() != "# shared\n" {
		t.Fatal(vm2.Source.Get())
	}
	vm2.DraftAction.Set(item(vm2, "Restore my draft"))
	if vm2.Source.Get() != "# mine!\n" {
		t.Fatalf("restored %q", vm2.Source.Get())
	}
}
