package playground

import (
	"encoding/json"
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

type journalEntry struct {
	V      int    `json:"v"`
	Source string `json:"source"`
	Saved  int64  `json:"saved"` // Unix milliseconds
}

// JournalDeck writes the deck's text to the journal, when it changed since
// the last write. The host calls it before SaveDraft, on the save tick and
// when the page is hidden or left.
func (vm *ViewModel) JournalDeck() {
	d := vm.drafts
	if d.jr == nil {
		return
	}
	src := vm.Source.Get()
	if src == d.journaled {
		return
	}
	if len(src) > JournalBudget {
		if !d.jrWarned {
			d.jrWarned = true
			vm.Status.Set("This deck is too large for the quick save: the edits of the last half-second before you close the page may not be kept.")
		}
		d.jr.JournalClear() // an older, shorter text must not come back
		d.journaled = ""
		return
	}
	b, _ := json.Marshal(journalEntry{V: 1, Source: src, Saved: d.now().UnixMilli()})
	if err := d.jr.JournalWrite(string(b)); err != nil {
		if !d.jrWarned {
			d.jrWarned = true
			vm.Status.Set("⚠ The quick save is refused (" + err.Error() + "): the edits of the last half-second before you close the page may not be kept.")
		}
		return
	}
	d.journaled = src
}

// journalSaved clears the journal once the stored draft holds its text.
func (vm *ViewModel) journalSaved(text string) {
	d := vm.drafts
	if d.jr != nil && d.journaled == text {
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
	if json.Unmarshal([]byte(s), &e) != nil || e.V != 1 || len(e.Source) > JournalBudget || e.Source == "" {
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
		return e.Source, at, true
	}
	p := d.st.Cached(DraftProject)
	if p != nil && (p.Text(store.DeckPath) == e.Source || !at.After(p.Saved)) {
		d.jr.JournalClear()
		return "", time.Time{}, false
	}
	if p == nil {
		p = store.NewProject(DraftProject)
	}
	if err := p.Write(store.DeckPath, []byte(e.Source), at); err != nil {
		return "", time.Time{}, false
	}
	d.journaled = e.Source
	d.st.Save(p, at, func(err error) {
		if err == nil {
			vm.journalSaved(e.Source)
		}
	})
	return "", time.Time{}, false
}
