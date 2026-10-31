package playground

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/mvvm"
)

// Projects is the host's project store (playground/store): IndexedDB in
// the page, the projects it holds loaded before the view-model starts. A
// host without one still gives a working playground; nothing is kept.
type Projects interface {
	Projects() *store.Store
	// Announce tells the page's other tabs that this one saved a project
	// (BroadcastChannel); they call DraftSavedElsewhere.
	Announce(name string, saved time.Time)
}

// Confirmer asks the reader a yes/no question (window.confirm).
type Confirmer interface {
	Confirm(question string) bool
}

// The projects the playground keeps: the draft (the deck being edited,
// deck.md, with the files the file panel adds), the draft it last replaced,
// and the UI state, apart from the deck.
const (
	DraftProject    = "draft"
	PreviousProject = "previous"
	UIProject       = "ui"
)

// drafts keeps the deck across reloads.
type drafts struct {
	st  *store.Store // nil: nothing is kept
	ann func(string, time.Time)
	now func() time.Time
	// mine is this tab's draft project; foreign says the stored draft is
	// another tab's (or the one a shared link set aside), to be kept as the
	// previous draft before this tab's first save replaces it.
	mine         *store.Project
	foreign      bool
	foreignSaved time.Time
	dirty        bool // the source changed since the last save
	// jr is the synchronous journal (nil: none); journaled the text last
	// written to it; jrWarned whether its refusal was said.
	jr        Journal
	journaled string
	jrWarned  bool
	failed    bool // the last save failed (said, and marked on renders)
	// Items are the Draft menu's items (item 0 is its title), actions what
	// each does (nil: nothing).
	Items   []string
	actions []func()
}

// NewTabID is a random id for a page load: the store stamps a project
// with it, and two tabs tell their drafts apart by it.
func NewTabID() string {
	var b [9]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}

// when says when a draft was saved, the way the status line says it.
func (d *drafts) when(t time.Time) string {
	now := d.now()
	if now.Sub(t) < 20*time.Hour && now.Day() == t.Day() {
		return T(KDateToday, clock(t))
	}
	return T(KDateAt, dateShort(t), clock(t))
}

// short is when, for the Draft menu: the time today, else the day.
func (d *drafts) short(t time.Time) string {
	now := d.now()
	if now.Sub(t) < 20*time.Hour && now.Day() == t.Day() {
		return clock(t)
	}
	return dateShort(t)
}

// startDrafts restores the deck at load: a shared link's deck wins and the
// draft is kept and offered; else the draft, when there is one. It returns
// the deck to start with and the status to show ("" for none).
func (vm *ViewModel) startDrafts(shared string, hasShared bool) (src, status string) {
	d := vm.drafts
	// a journal newer than the stored draft becomes the draft first
	jtext, jat, jonly := vm.replayJournal()
	if d.st == nil {
		switch {
		case hasShared:
			return shared, ""
		case jonly:
			return jtext, T(KStatusRestoredDraft, d.when(jat))
		}
		return Examples[0].Source, ""
	}
	saved := d.st.Cached(DraftProject)
	if saved != nil && saved.Text(store.DeckPath) == "" {
		saved = nil
	}
	switch {
	case hasShared:
		d.mine = store.NewProject(DraftProject)
		if saved != nil {
			d.foreign, d.foreignSaved = true, saved.Saved
			if saved.Text(store.DeckPath) != shared {
				status = T(KStatusSharedKeptDraft, d.when(saved.Saved))
			}
		}
		return shared, status
	case saved != nil:
		// this tab takes the draft over
		d.mine = saved
		return saved.Text(store.DeckPath), T(KStatusRestoredDraftReset, d.when(saved.Saved))
	}
	d.mine = store.NewProject(DraftProject)
	return Examples[0].Source, ""
}

// SaveDraft stores the deck when it changed: the host calls it a moment
// after the typing stops, and when the page is hidden. A draft another tab
// saved is kept as the previous draft, not overwritten.
func (vm *ViewModel) SaveDraft() {
	d := vm.drafts
	// the panel's deck size and the usage line follow the text on every
	// save tick, whatever changed it (typing, a ribbon command, an example,
	// Reset, an import), and with no store as well
	vm.refreshFiles()
	if !d.dirty || d.st == nil {
		return
	}
	if d.foreign {
		vm.setAside(d.foreignSaved)
	}
	if err := d.mine.Write(store.DeckPath, []byte(vm.Source.Get()), d.now()); err != nil {
		vm.saveFailed(err)
		return
	}
	d.dirty = false
	p := d.mine
	sig := vm.journalSig() // what this save stores: the deck and the text files
	d.st.Save(p, d.now(), func(err error) {
		if err != nil {
			d.dirty = true
			vm.saveFailed(err)
			return
		}
		vm.journalSaved(sig)
		if d.ann != nil {
			d.ann(DraftProject, p.Saved)
		}
		if d.failed {
			d.failed = false
			vm.Status.Set(T(KStatusDraftSavedAgain))
		}
		vm.refreshDraftMenu()
	})
}

// setAside keeps the stored draft (another tab's, or set aside by a shared
// link) as the previous one, and makes this tab's draft the one saved
// next, all its files with it.
func (vm *ViewModel) setAside(saved time.Time) {
	d := vm.drafts
	d.foreign = false
	d.st.Copy(DraftProject, PreviousProject, saved, vm.storeErr)
	d.st.Delete(DraftProject, vm.storeErr)
	d.mine = d.mine.Clone(DraftProject)
}

func (vm *ViewModel) storeErr(err error) {
	if err != nil {
		vm.saveFailed(err)
	}
}

func (vm *ViewModel) saveFailed(err error) {
	if vm.drafts.failed {
		return
	}
	vm.drafts.failed = true
	vm.Status.Set(T(KStatusDraftNotSaved, err.Error()))
}

// DraftSavedElsewhere is the host's word that another tab saved a project:
// this tab says so, and its own next save keeps the other tab's draft as
// the previous one.
func (vm *ViewModel) DraftSavedElsewhere(name, tab string, saved time.Time) {
	d := vm.drafts
	if d.st == nil || tab == d.st.Tab() || name != DraftProject {
		return
	}
	// this tab's next save (an edit here) sets the other's draft aside
	d.foreign, d.foreignSaved = true, saved
	vm.Status.Set(T(KStatusOtherTab, d.when(saved)))
	vm.refreshDraftMenu()
}

// archiveDraft keeps the draft as the previous one before the deck is
// replaced as a whole (an example, Reset).
func (vm *ViewModel) archiveDraft(next string) {
	d := vm.drafts
	if d.st == nil {
		return
	}
	vm.SaveDraft()
	if d.foreign || d.mine.Text(store.DeckPath) == "" || d.mine.Text(store.DeckPath) == next {
		return
	}
	d.st.Copy(DraftProject, PreviousProject, d.mine.Saved, vm.storeErr)
	d.st.Delete(DraftProject, vm.storeErr)
	d.mine = store.NewProject(DraftProject)
}

// restorePrevious swaps the draft and the previous draft: neither is lost.
func (vm *ViewModel) restorePrevious() {
	d := vm.drafts
	vm.SaveDraft()
	const swap = "swap"
	at, _ := d.st.Exists(PreviousProject)
	d.st.Copy(PreviousProject, swap, at, vm.storeErr)
	d.st.Copy(DraftProject, PreviousProject, d.mine.Saved, vm.storeErr)
	d.st.Copy(swap, DraftProject, at, vm.storeErr)
	d.st.Delete(swap, vm.storeErr)
	d.st.Load(DraftProject, func(p *store.Project, err error) {
		if err != nil || p == nil {
			vm.storeErr(err)
			return
		}
		d.mine, d.foreign = p, false
		vm.notice(T(KStatusRestoredTheDraft, d.when(at)))
		vm.Source.Set(p.Text(store.DeckPath))
		d.dirty = false
		vm.refreshDraftMenu()
	})
}

// restoreDraft puts the stored draft back (a shared link set it aside).
func (vm *ViewModel) restoreDraft() {
	d := vm.drafts
	d.st.Load(DraftProject, func(p *store.Project, err error) {
		if err != nil || p == nil {
			vm.storeErr(err)
			return
		}
		d.mine, d.foreign = p, false
		vm.notice(T(KStatusRestoredTheDraft, d.when(p.Saved)))
		vm.Source.Set(p.Text(store.DeckPath))
		d.dirty = false
		vm.refreshDraftMenu()
	})
}

// ForgetDrafts removes every stored project of the playground.
func (vm *ViewModel) ForgetDrafts() {
	d := vm.drafts
	if d.st == nil {
		return
	}
	for _, n := range []string{DraftProject, PreviousProject, UIProject} {
		d.st.Delete(n, vm.storeErr)
	}
	if d.jr != nil {
		d.jr.JournalClear()
	}
	d.mine, d.foreign, d.dirty = store.NewProject(DraftProject), false, false
	if d.jr != nil {
		// the text on screen is not journaled again until it is edited:
		// leaving the page must not bring a forgotten deck back
		d.journaled = sigOf(vm.Source.Get(), d.mine)
	}
	vm.Status.Set(T(KStatusForgot))
	vm.refreshDraftMenu()
}

// refreshDraftMenu rebuilds the Draft menu from what is stored.
func (vm *ViewModel) refreshDraftMenu() {
	d := vm.drafts
	items, acts := []string{T(KLblDraft)}, []func(){nil}
	if d.st != nil {
		// a draft a shared link set aside (loaded as the page started)
		if c := d.st.Cached(DraftProject); d.foreign && c != nil && c != d.mine && c.Text(store.DeckPath) != vm.Source.Get() {
			items = append(items, T(KDraftRestoreMine, d.short(c.Saved)))
			acts = append(acts, vm.restoreDraft)
		}
		if at, ok := d.st.Exists(PreviousProject); ok {
			items = append(items, T(KDraftRestoreReplaced, d.short(at)))
			acts = append(acts, vm.restorePrevious)
		}
		if _, ok := d.st.Exists(DraftProject); ok || len(items) > 1 {
			items = append(items, T(KDraftForget))
			acts = append(acts, vm.ForgetDrafts)
		}
	}
	if vm.Source.Get() != Examples[max(vm.Example.Get(), 0)].Source {
		items = append(items, T(KDraftResetExample))
		acts = append(acts, func() { vm.Reset.Execute() })
	}
	if len(items) == 1 {
		items = append(items, T(KDraftNothing))
		acts = append(acts, nil)
	}
	if fmt.Sprint(items) != fmt.Sprint(d.Items) {
		d.Items = items
		vm.DraftVersion.Set(vm.DraftVersion.Get() + 1)
	}
	d.actions = acts
}

// DraftItems are the Draft menu's items (item 0 is its title).
func (vm *ViewModel) DraftItems() []string { return vm.drafts.Items }

// chooseDraftAction runs the chosen Draft menu item and returns the menu
// to its title.
func (vm *ViewModel) chooseDraftAction(i int) {
	if i <= 0 || i >= len(vm.drafts.actions) {
		return
	}
	act := vm.drafts.actions[i]
	vm.DraftAction.Set(0)
	if act != nil {
		act()
	}
}

// saveUI stores the small UI state, beside the deck.
func (vm *ViewModel) saveUI() {
	d := vm.drafts
	if d.st == nil || vm.quiet {
		return
	}
	p := d.st.Cached(UIProject)
	if p == nil {
		p = store.NewProject(UIProject)
	}
	p.SetUI("mode", strconv.Itoa(vm.Mode.Get()))
	p.SetUI("minimap", strconv.FormatBool(vm.Minimap.Get()))
	p.SetUI("thumbs", strconv.FormatBool(vm.Thumbs.Get()))
	// the panes and the ribbon are stored once chosen: a default is not
	if vm.saved["files"] {
		p.SetUI("files", strconv.FormatBool(vm.FilesPane.Get()))
	}
	if vm.saved["help"] {
		p.SetUI("help", strconv.FormatBool(vm.HelpPane.Get()))
	}
	if vm.saved["ribbon"] {
		p.SetUI("ribbon", strconv.FormatBool(vm.RibbonCollapsed.Get()))
	}
	vm.saveSizes(p.SetUI)                 // the widths the person dragged, not the defaults
	d.st.Save(p, d.now(), func(error) {}) // UI state: best effort
}

// loadUI reads it back, ignoring anything out of range.
func (vm *ViewModel) loadUI() {
	d := vm.drafts
	if d.st == nil {
		return
	}
	if p := d.st.Cached(UIProject); p != nil {
		if m, err := strconv.Atoi(p.UI["mode"]); err == nil && m >= 0 && m < len(ModeNames) {
			vm.Mode.Set(m)
		}
		if b, err := strconv.ParseBool(p.UI["minimap"]); err == nil {
			vm.Minimap.Set(b)
		}
		if b, err := strconv.ParseBool(p.UI["thumbs"]); err == nil {
			vm.Thumbs.Set(b)
		}
		// the panes and the ribbon: only what the person chose is stored
		for key, obs := range map[string]*mvvm.Observable[bool]{"files": vm.FilesPane, "help": vm.HelpPane, "ribbon": vm.RibbonCollapsed} {
			if b, err := strconv.ParseBool(p.UI[key]); err == nil {
				vm.quiet = true
				obs.Set(b)
				vm.quiet = false
				vm.saved[key] = true
			}
		}
		vm.loadSizes(func(key string) string { return p.UI[key] })
	}
}
