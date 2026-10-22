package playground

import (
	"strings"
	"testing"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// tab is a page of the playground over a browser's storage (a Memory
// shared by the tabs), answering confirmations with answer, and recording
// what it announced to the other tabs.
type tab struct {
	*fakeHost
	st        *store.Store
	answer    bool
	asked     int
	announced []string
}

func (t *tab) Projects() *store.Store            { return t.st }
func (t *tab) Announce(name string, _ time.Time) { t.announced = append(t.announced, name) }
func (t *tab) Confirm(string) bool               { t.asked++; return t.answer }
func load(m *store.Memory, name string) (p *store.Project) {
	m.Load(name, func(q *store.Project, _ error) { p = q })
	return p
}

// open loads a page as the host does: the stored projects first, then the
// view-model.
func open(t *testing.T, m *store.Memory, fragment string) (*ViewModel, *tab) {
	t.Helper()
	h := &tab{fakeHost: newHost(), answer: true}
	h.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, fragment)
	tr.vm = vm
	answer(tr, vm)
	return vm, h
}

func item(vm *ViewModel, prefix string) int {
	for i, s := range vm.DraftItems() {
		if strings.HasPrefix(s, prefix) {
			return i
		}
	}
	return -1
}

func deckOf(m *store.Memory, name string) string {
	if p := load(m, name); p != nil {
		return p.Text(store.DeckPath)
	}
	return ""
}

func TestDraftSurvivesAReload(t *testing.T) {
	m := store.NewMemory()
	vm, h := open(t, m, "")
	if vm.Source.Get() != Examples[0].Source || item(vm, "Nothing saved yet") < 0 {
		t.Fatal("a first visit starts on the example")
	}
	vm.Source.Set("---\nslides: true\n---\n# mine\n")
	vm.SaveDraft()
	if deckOf(m, DraftProject) != "---\nslides: true\n---\n# mine\n" || len(h.announced) != 1 {
		t.Fatal("not saved, or not announced")
	}
	vm2, _ := open(t, m, "")
	if vm2.Source.Get() != "---\nslides: true\n---\n# mine\n" || !strings.Contains(vm2.Status.Get(), "Restored your draft of today at") {
		t.Fatalf("%q / %q", vm2.Source.Get(), vm2.Status.Get())
	}
	// the draft is a project: deck.md, and other files beside it survive the saves
	p := load(m, DraftProject)
	if p.Paths()[0] != store.DeckPath {
		t.Fatal(p.Paths())
	}
}

func TestSharedLinkWinsTheDraftIsKept(t *testing.T) {
	m := store.NewMemory()
	vm, _ := open(t, m, "")
	vm.Source.Set("# my draft\n")
	vm.SaveDraft()
	frag, err := EncodeShare("# shared\n")
	if err != nil {
		t.Fatal(err)
	}
	vm2, _ := open(t, m, frag)
	if vm2.Source.Get() != "# shared\n" || !strings.Contains(vm2.Status.Get(), "is kept: Draft ▸ Restore my draft") {
		t.Fatalf("%q / %q", vm2.Source.Get(), vm2.Status.Get())
	}
	i := item(vm2, "Restore my draft")
	if i < 0 {
		t.Fatalf("menu %v", vm2.DraftItems())
	}
	vm2.DraftAction.Set(i)
	if vm2.Source.Get() != "# my draft\n" || vm2.DraftAction.Get() != 0 {
		t.Fatalf("restored %q", vm2.Source.Get())
	}
	// editing the shared deck keeps the old draft as the previous one
	vm3, _ := open(t, m, frag)
	vm3.Source.Set("# shared, edited\n")
	vm3.SaveDraft()
	if deckOf(m, PreviousProject) != "# my draft\n" || deckOf(m, DraftProject) != "# shared, edited\n" {
		t.Fatalf("previous %q, draft %q", deckOf(m, PreviousProject), deckOf(m, DraftProject))
	}
}

func TestTwoTabsDoNotEraseEachOther(t *testing.T) {
	m := store.NewMemory()
	a, ha := open(t, m, "")
	b, _ := open(t, m, "")
	a.Source.Set("# from A\n")
	a.SaveDraft()
	b.DraftSavedElsewhere(DraftProject, ha.st.Tab(), time.Now())
	if !strings.Contains(b.Status.Get(), "Another tab saved its draft") {
		t.Fatal(b.Status.Get())
	}
	b.Source.Set("# from B\n")
	b.SaveDraft()
	if deckOf(m, DraftProject) != "# from B\n" || deckOf(m, PreviousProject) != "# from A\n" {
		t.Fatal("A's draft was erased")
	}
	// restoring the replaced draft swaps the two: nothing is lost
	b.DraftAction.Set(item(b, "Restore replaced"))
	if b.Source.Get() != "# from A\n" || deckOf(m, PreviousProject) != "# from B\n" || deckOf(m, DraftProject) != "# from A\n" {
		t.Fatalf("swap: %q / %q / %q", b.Source.Get(), deckOf(m, DraftProject), deckOf(m, PreviousProject))
	}
	// a tab's own save, or another project, raises no alarm
	b.Status.Set("quiet")
	b.DraftSavedElsewhere(DraftProject, b.drafts.st.Tab(), time.Now())
	b.DraftSavedElsewhere(UIProject, ha.st.Tab(), time.Now())
	if b.Status.Get() != "quiet" {
		t.Fatal(b.Status.Get())
	}
}

func TestQuotaIsSaidNotSwallowed(t *testing.T) {
	m := store.NewMemory()
	vm, _ := open(t, m, "")
	m.Quota = 1000
	vm.Source.Set(strings.Repeat("x", 2000))
	vm.SaveDraft()
	if !strings.HasPrefix(vm.Status.Get(), "⚠ This browser did not keep your draft (QuotaExceededError)") {
		t.Fatal(vm.Status.Get())
	}
	// while it fails, every render's status says so, briefly
	vm.Source.Set(strings.Repeat("y", 2000))
	vm.SaveDraft()
	if st := vm.Status.Get(); strings.HasPrefix(st, "⚠") || !strings.HasSuffix(st, "⚠ draft not kept in this browser") {
		t.Fatal(st)
	}
	vm.Source.Set("small")
	vm.SaveDraft()
	if !strings.Contains(vm.Status.Get(), "saved in this browser again") || deckOf(m, DraftProject) != "small" {
		t.Fatal(vm.Status.Get())
	}
}

func TestBlockedStorageStillWorks(t *testing.T) {
	m := store.NewMemory()
	m.Blocked = true
	vm, _ := open(t, m, "")
	if vm.Source.Get() != Examples[0].Source || vm.Doc.Get() == "" {
		t.Fatal("no playground without storage")
	}
	vm.Source.Set("# x")
	vm.SaveDraft()
	if !strings.Contains(vm.Status.Get(), "SecurityError") {
		t.Fatal(vm.Status.Get())
	}
	// and a host with no store at all: drafts are off, nothing breaks
	tr := &fakeTransport{}
	plain := NewViewModel(tr, newHost(), "")
	plain.Source.Set("# y")
	plain.SaveDraft()
	plain.ForgetDrafts()
	if item(plain, "Nothing saved yet") < 0 {
		t.Fatal(plain.DraftItems())
	}
}

func TestResetAsksAndKeeps(t *testing.T) {
	m := store.NewMemory()
	vm, h := open(t, m, "")
	vm.Source.Set("# precious\n")
	h.answer = false
	vm.Reset.Execute()
	if vm.Source.Get() != "# precious\n" || h.asked != 1 {
		t.Fatal("Reset did not ask, or did not listen")
	}
	h.answer = true
	vm.Reset.Execute()
	if vm.Source.Get() != Examples[0].Source || deckOf(m, PreviousProject) != "# precious\n" {
		t.Fatal("Reset lost the draft")
	}
	// an unchanged deck: nothing to ask
	vm.Reset.Execute()
	if h.asked != 2 {
		t.Fatal(h.asked)
	}
	// an example replaces the deck the same way
	vm.Source.Set("# again\n")
	vm.Example.Set(2)
	if deckOf(m, PreviousProject) != "# again\n" {
		t.Fatal("an example lost the draft")
	}
	vm.SaveDraft()
	if deckOf(m, DraftProject) != Examples[2].Source {
		t.Fatal(deckOf(m, DraftProject))
	}
}

func TestForgetAndUIState(t *testing.T) {
	m := store.NewMemory()
	vm, _ := open(t, m, "")
	vm.Mode.Set(int(ModeSlideshow))
	vm.Source.Set("# x\n")
	vm.SaveDraft()
	vm2, _ := open(t, m, "")
	if vm2.Mode.Get() != int(ModeSlideshow) {
		t.Fatal("the preview mode was not kept")
	}
	vm2.DraftAction.Set(item(vm2, "Forget my drafts"))
	for _, n := range []string{DraftProject, PreviousProject, UIProject} {
		if load(m, n) != nil {
			t.Fatal(n, "kept")
		}
	}
	if !strings.Contains(vm2.Status.Get(), "Forgot your drafts") || item(vm2, "Nothing saved yet") < 0 {
		t.Fatalf("%q %v", vm2.Status.Get(), vm2.DraftItems())
	}
	vm3, _ := open(t, m, "")
	if vm3.Source.Get() != Examples[0].Source || vm3.Mode.Get() != int(ModeOverview) {
		t.Fatal("forgotten, yet restored")
	}
}
