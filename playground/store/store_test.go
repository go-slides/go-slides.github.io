package store

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)

func TestCleanPath(t *testing.T) {
	for in, want := range map[string]string{"deck.md": "deck.md", "themes//corp.css": "themes/corp.css", "a/./b/../c.png": "a/c.png"} {
		if got, err := CleanPath(in); err != nil || got != want {
			t.Errorf("%q: %q %v", in, got, err)
		}
	}
	for _, bad := range []string{"", "/etc/passwd", "../x", "a/../../x", ".", "a\\b", "a\x00b", "a\nb", strings.Repeat("a", MaxPath+1)} {
		if _, err := CleanPath(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestProjectFilesAndBounds(t *testing.T) {
	p := NewProject("draft")
	if err := p.Write("deck.md", []byte("# x"), t0); err != nil {
		t.Fatal(err)
	}
	if err := p.Write("media/a.png", []byte{1, 2, 3}, t0); err != nil {
		t.Fatal(err)
	}
	if p.Text("deck.md") != "# x" || !slices.Equal(p.Paths(), []string{"deck.md", "media/a.png"}) || p.Size() != 6 {
		t.Fatal(p.Paths(), p.Size())
	}
	if err := p.Write("../out", nil, t0); err == nil {
		t.Fatal("a path out of the project")
	}
	if err := p.Write("big", make([]byte, MaxFileSize+1), t0); err == nil {
		t.Fatal("an oversize file")
	}
	q := NewProject("many")
	for i := 0; i < MaxFiles; i++ {
		if err := q.Write(strings.Repeat("f", 1+i%50)+"/"+string(rune('a'+i%26))+strings.Repeat("x", i/26), nil, t0); err != nil {
			t.Fatal(i, err)
		}
	}
	if err := q.Write("one-more", nil, t0); err == nil {
		t.Fatal("past MaxFiles")
	}
	// the vfs view: the files under the deck's directory, nothing outside
	fs := p.FS("/work")
	if b, err := fs.ReadFile("/work/media/a.png"); err != nil || len(b) != 3 {
		t.Fatal(b, err)
	}
	if _, err := fs.ReadFile("/etc/passwd"); err == nil {
		t.Fatal("the vfs reads outside the project")
	}
	if es, err := fs.ReadDir("/work/media"); err != nil || len(es) != 1 {
		t.Fatal(es, err)
	}
	p.Remove("media/a.png")
	if _, ok := p.Read("media/a.png"); ok {
		t.Fatal("removed")
	}
	if w, r := p.Changes(); !slices.Equal(w, []string{"deck.md"}) || !slices.Equal(r, []string{"media/a.png"}) {
		t.Fatal(w, r)
	}
}

func TestStoreSavesTheChangesOnly(t *testing.T) {
	m := NewMemory()
	s := New(m, "tab1")
	p := NewProject("draft")
	p.Write(DeckPath, []byte("# one"), t0)
	p.Write("media/a.png", []byte("png"), t0)
	p.SetUI("mode", "1")
	var err error
	s.Save(p, t0, func(e error) { err = e })
	if err != nil || p.Dirty() {
		t.Fatal(err, p.Dirty())
	}
	if at, ok := m.Saved("draft"); !ok || !at.Equal(t0) {
		t.Fatal(at, ok)
	}
	// a second page reads it back
	var got *Project
	New(m, "tab2").Load("draft", func(q *Project, e error) { got, err = q, e })
	if err != nil || got.Text(DeckPath) != "# one" || got.Tab != "tab1" || got.UI["mode"] != "1" || got.Dirty() {
		t.Fatalf("%+v %v", got, err)
	}
	// one change, one file written; a removal is carried out
	p.Write(DeckPath, []byte("# two"), t0)
	p.Remove("media/a.png")
	if w, r := p.Changes(); len(w) != 1 || len(r) != 1 {
		t.Fatal(w, r)
	}
	s.Save(p, t0.Add(time.Minute), func(e error) { err = e })
	New(m, "tab2").Load("draft", func(q *Project, e error) { got = q })
	if got.Text(DeckPath) != "# two" || len(got.Paths()) != 1 {
		t.Fatal(got.Paths())
	}
	// an unchanged project is not written
	before, _ := m.Saved("draft")
	s.Save(p, t0.Add(time.Hour), func(error) {})
	if after, _ := m.Saved("draft"); !after.Equal(before) {
		t.Fatal("an unchanged project was written")
	}
}

func TestStoreFailuresKeepTheChanges(t *testing.T) {
	m := NewMemory()
	m.Quota = 10
	s := New(m, "tab")
	p := NewProject("draft")
	p.Write(DeckPath, []byte(strings.Repeat("x", 20)), t0)
	var err error
	s.Save(p, t0, func(e error) { err = e })
	if !errors.Is(err, ErrQuota) || !p.Dirty() {
		t.Fatal(err, p.Dirty())
	}
	m.Quota = 0
	s.Save(p, t0, func(e error) { err = e })
	if err != nil || p.Dirty() {
		t.Fatal("the change was lost by the failed save")
	}
	m.Blocked = true
	p.Write(DeckPath, []byte("y"), t0)
	s.Save(p, t0, func(e error) { err = e })
	if err == nil {
		t.Fatal("a blocked store saved")
	}
	// no backend at all
	New(nil, "tab").Save(p, t0, func(e error) { err = e })
	if !errors.Is(err, ErrNoStore) {
		t.Fatal(err)
	}
}

func TestStoreCopyAndDelete(t *testing.T) {
	m := NewMemory()
	a := New(m, "a")
	p := NewProject("draft")
	p.Write(DeckPath, []byte("# A"), t0)
	a.Save(p, t0, func(error) {})
	// another tab copies a draft it never loaded: known, not cached
	b := New(m, "b")
	b.Copy("draft", "previous", t0, func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	})
	if at, ok := b.Exists("previous"); !ok || !at.Equal(t0) || b.Cached("previous") != nil {
		t.Fatal(at, ok)
	}
	var prev *Project
	b.Load("previous", func(q *Project, err error) { prev = q })
	if prev == nil || prev.Text(DeckPath) != "# A" || b.Cached("previous") == nil {
		t.Fatal("previous")
	}
	b.Delete("previous", func(error) {})
	if _, ok := b.Exists("previous"); ok {
		t.Fatal("deleted")
	}
	b.Load("previous", func(q *Project, err error) { prev = q })
	if prev != nil {
		t.Fatal("still stored")
	}
	// copying a project that does not exist deletes the target
	a.Copy("nothing", "draft", t0, func(error) {})
	a.Load("draft", func(q *Project, err error) { prev = q })
	if prev != nil {
		t.Fatal("copy of nothing")
	}
}
