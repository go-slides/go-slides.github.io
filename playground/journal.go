package playground

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// Journal is the page's synchronous write-ahead log of the deck's text
// (localStorage of the top-level page). IndexedDB writes are asynchronous,
// and a page being closed may not wait for one (pagehide does not); a
// localStorage write is done when the call returns. So the text is written
// here first, on every save tick and when the page is hidden, and replayed
// at the next load when it is newer than the stored draft. Media stay in
// IndexedDB: the journal holds the deck's text alone.
type Journal interface {
	JournalWrite(value string) error
	JournalRead() (string, bool)
	JournalClear()
}

// JournalBudget bounds the text the journal holds (UTF-16 code units are
// what localStorage counts; a deck past it is kept by IndexedDB alone, as
// before, and the status says so once).
const JournalBudget = 1 << 20

// journalEntry is what the journal holds. Version 1 held the deck alone;
// version 2 holds the project's text files too (slides.yaml, themes, kinds,
// data: everything but the media), all of them, so that a file deleted
// since the last save is deleted again when the journal is replayed.
type journalEntry struct {
	V      int               `json:"v"`
	Source string            `json:"source"`
	Files  map[string]string `json:"files,omitempty"`
	Saved  int64             `json:"saved"` // Unix milliseconds
}

// projectTexts are the project's text files by their project paths: what
// the journal holds besides the deck.
func (vm *ViewModel) projectTexts() map[string]string {
	p := vm.project()
	out := map[string]string{}
	for _, k := range p.Paths() {
		if k != store.DeckPath && !IsMedia(k) {
			out[k] = p.Text(k)
		}
	}
	return out
}

// journalSig is what the journal would hold now, to tell whether it changed:
// the deck and the text files, in path order.
func (vm *ViewModel) journalSig() string { return sigOf(vm.Source.Get(), vm.project()) }

// JournalDeck writes the deck's text and the project's text files to the
// journal, when they changed since the last write. The host calls it
// before SaveDraft, on the save tick (of the deck and of the files) and
// when the page is hidden or left.
func (vm *ViewModel) JournalDeck() {
	d := vm.drafts
	if d.jr == nil {
		return
	}
	sig := vm.journalSig()
	if sig == d.journaled {
		return
	}
	if len(sig) > JournalBudget {
		if !d.jrWarned {
			d.jrWarned = true
			vm.Status.Set(T(KStatusJournalLarge))
		}
		d.jr.JournalClear() // an older, shorter text must not come back
		d.journaled = ""
		return
	}
	b, _ := json.Marshal(journalEntry{V: 2, Source: vm.Source.Get(), Files: vm.projectTexts(), Saved: d.now().UnixMilli()})
	if err := d.jr.JournalWrite(string(b)); err != nil {
		if !d.jrWarned {
			d.jrWarned = true
			vm.Status.Set(T(KStatusJournalRefused, ErrText(err)))
		}
		return
	}
	d.journaled = sig
}

// journalSaved clears the journal once the stored draft holds what it
// holds (sig, a journalSig).
func (vm *ViewModel) journalSaved(sig string) {
	d := vm.drafts
	if d.jr != nil && d.journaled == sig {
		d.jr.JournalClear()
		d.journaled = ""
	}
}

// readJournal returns the journal's entry, if one is there and sound.
func (d *drafts) readJournal() (journalEntry, bool) {
	if d.jr == nil {
		return journalEntry{}, false
	}
	s, ok := d.jr.JournalRead()
	if !ok {
		return journalEntry{}, false
	}
	var e journalEntry
	err := json.Unmarshal([]byte(s), &e)
	size := len(e.Source)
	for k, v := range e.Files {
		size += len(k) + len(v)
	}
	if err != nil || (e.V != 1 && e.V != 2) || size > JournalBudget || e.Source == "" || !journalPaths(e.Files) {
		d.jr.JournalClear()
		return journalEntry{}, false
	}
	return e, true
}

// replayJournal, at load, puts a journal newer than the stored draft into
// the draft project and saves it, clearing the journal once it is stored;
// an older or equal journal is cleared. It returns the journal's text when
// there is no store at all (the journal is then all there is).
func (vm *ViewModel) replayJournal() (text string, at time.Time, ok bool) {
	d := vm.drafts
	e, found := d.readJournal()
	if !found {
		return "", time.Time{}, false
	}
	at = time.UnixMilli(e.Saved)
	if d.st == nil {
		// no store: the journal is all there is, its files with it
		d.mine = store.NewProject(DraftProject)
		for k, v := range e.Files {
			_ = d.mine.Write(k, []byte(v), at)
		}
		return e.Source, at, true
	}
	p := d.st.Cached(DraftProject)
	if p != nil && (e.holds(p) || !at.After(p.Saved)) {
		d.jr.JournalClear()
		return "", time.Time{}, false
	}
	if p == nil {
		p = store.NewProject(DraftProject)
	}
	if err := p.Write(store.DeckPath, []byte(e.Source), at); err != nil {
		return "", time.Time{}, false
	}
	if e.V == 2 {
		// the text files as they were: written, and those gone since removed
		for k, v := range e.Files {
			if err := p.Write(k, []byte(v), at); err != nil {
				return "", time.Time{}, false
			}
		}
		for _, k := range p.Paths() {
			if _, kept := e.Files[k]; !kept && k != store.DeckPath && !IsMedia(k) {
				p.Remove(k)
			}
		}
	}
	// what the journal holds, as journalSig counts it once the draft is loaded
	d.journaled = sigOf(e.Source, p)
	sig := d.journaled
	d.st.Save(p, at, func(err error) {
		if err == nil {
			vm.journalSaved(sig)
		}
	})
	return "", time.Time{}, false
}

// holds reports whether the stored project holds what the entry does: the
// deck, and (version 2) exactly its text files.
func (e journalEntry) holds(p *store.Project) bool {
	if p.Text(store.DeckPath) != e.Source {
		return false
	}
	if e.V == 1 {
		return true
	}
	n := 0
	for _, k := range p.Paths() {
		if k == store.DeckPath || IsMedia(k) {
			continue
		}
		if v, ok := e.Files[k]; !ok || v != p.Text(k) {
			return false
		}
		n++
	}
	return n == len(e.Files)
}

// journalPaths reports whether a journal's file paths are project paths of
// text files (a garbled or hostile entry is dropped whole).
func journalPaths(files map[string]string) bool {
	for k := range files {
		if c, err := store.CleanPath(k); err != nil || c != k || k == store.DeckPath || IsMedia(k) {
			return false
		}
	}
	return true
}

// sigOf is journalSig for a deck's text and a project.
func sigOf(src string, p *store.Project) string {
	var b strings.Builder
	b.WriteString(src)
	for _, k := range p.Paths() { // sorted
		if k != store.DeckPath && !IsMedia(k) {
			b.WriteString("\x00" + k + "\x00" + p.Text(k))
		}
	}
	return b.String()
}
