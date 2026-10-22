// Package store keeps the playground's projects in the browser: a project
// is a set of files (path → bytes) with a little metadata, the deck being
// one of them (deck.md), and go-slides reads it through a vfs.FS. The
// playground keeps the reader's draft in it across reloads; the file panel
// (themes, kinds, data files, media) builds on the same projects.
//
// Persistence is a Backend: IndexedDB in the page (cmd/playground-wasm),
// Memory in tests and where IndexedDB is refused. A Backend's calls answer
// through callbacks, because IndexedDB answers through events and Go's
// js/wasm cannot block in a JavaScript callback; Memory answers at once.
// The Store keeps the projects it loaded in memory, so the view-model reads
// them synchronously; writes go through to the Backend, the changed files
// only, and their outcome comes back to a callback.
//
// Why IndexedDB and not localStorage: a project's media (images, fonts)
// can outgrow localStorage's ~5 MB an origin, which also stores only
// strings; IndexedDB stores bytes, writes several records in one
// transaction, and its quota is a share of the disk. Limits here (MaxFiles,
// MaxFileSize, MaxProjectSize) keep a project far inside it.
package store

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-slides/slides/vfs"
)

// The bounds of a project.
const (
	MaxFiles       = 256
	MaxFileSize    = 8 << 20
	MaxProjectSize = 64 << 20
	MaxPath        = 255
)

// DeckPath is the deck's file in a project.
const DeckPath = "deck.md"

// File is one file of a project.
type File struct {
	Data     []byte
	Modified time.Time
}

// Project is a named set of files and its metadata.
type Project struct {
	Name string
	// Saved is when the project was last written, and Tab the page (one
	// random id per page load) that wrote it: two tabs of the playground
	// tell their drafts apart by it.
	Saved time.Time
	Tab   string
	// UI is small state kept with the project (the preview mode, the
	// panels shown); keys and values are the caller's.
	UI map[string]string

	files   map[string]File
	dirty   map[string]bool // written since the last save
	removed map[string]bool // removed since the last save
	meta    bool            // the metadata changed since the last save
}

// NewProject returns an empty project.
func NewProject(name string) *Project {
	return &Project{Name: name, UI: map[string]string{}, files: map[string]File{}, dirty: map[string]bool{}, removed: map[string]bool{}, meta: true}
}

// CleanPath checks and cleans a project path: relative, slash-separated,
// within the project (no "..", no leading "/"), of printable characters,
// at most MaxPath bytes. "a//b/./c.md" is "a/b/c.md".
func CleanPath(p string) (string, error) {
	if p == "" || len(p) > MaxPath {
		return "", fmt.Errorf("%q: a path is 1 to %d bytes", p, MaxPath)
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return "", fmt.Errorf("%q: a project path is relative and uses /", p)
	}
	for _, r := range p {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("%q: a control character in a path", p)
		}
	}
	c := path.Clean(p)
	if c == "." || c == ".." || strings.HasPrefix(c, "../") {
		return "", fmt.Errorf("%q: the path leaves the project", p)
	}
	return c, nil
}

// Read returns a file's contents.
func (p *Project) Read(name string) ([]byte, bool) {
	c, err := CleanPath(name)
	if err != nil {
		return nil, false
	}
	f, ok := p.files[c]
	return f.Data, ok
}

// Text returns a file as text ("" when absent).
func (p *Project) Text(name string) string {
	b, _ := p.Read(name)
	return string(b)
}

// Write sets a file, within the bounds; writing what is already there
// changes nothing.
func (p *Project) Write(name string, data []byte, now time.Time) error {
	c, err := CleanPath(name)
	if err != nil {
		return err
	}
	if len(data) > MaxFileSize {
		return fmt.Errorf("%s is %d KiB: a file is at most %d KiB", c, len(data)>>10, MaxFileSize>>10)
	}
	old, had := p.files[c]
	if had && string(old.Data) == string(data) {
		return nil
	}
	if !had && len(p.files) >= MaxFiles {
		return fmt.Errorf("a project holds at most %d files", MaxFiles)
	}
	if size := p.Size() - len(old.Data) + len(data); size > MaxProjectSize {
		return fmt.Errorf("the project would be %d MiB: at most %d MiB", size>>20, MaxProjectSize>>20)
	}
	p.files[c] = File{Data: append([]byte(nil), data...), Modified: now}
	p.dirty[c] = true
	delete(p.removed, c)
	return nil
}

// Remove deletes a file.
func (p *Project) Remove(name string) {
	c, err := CleanPath(name)
	if err != nil {
		return
	}
	if _, ok := p.files[c]; !ok {
		return
	}
	delete(p.files, c)
	delete(p.dirty, c)
	p.removed[c] = true
}

// SetUI sets a UI value.
func (p *Project) SetUI(key, value string) {
	if p.UI[key] != value {
		p.UI[key] = value
		p.meta = true
	}
}

// Paths are the project's files, sorted.
func (p *Project) Paths() []string {
	out := make([]string, 0, len(p.files))
	for k := range p.files {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// File returns a file and its metadata.
func (p *Project) File(name string) (File, bool) {
	c, err := CleanPath(name)
	if err != nil {
		return File{}, false
	}
	f, ok := p.files[c]
	return f, ok
}

// Size is the bytes the files hold.
func (p *Project) Size() int {
	n := 0
	for _, f := range p.files {
		n += len(f.Data)
	}
	return n
}

// Dirty reports whether the project changed since it was last saved.
func (p *Project) Dirty() bool { return p.meta || len(p.dirty) > 0 || len(p.removed) > 0 }

// Changes are the files written and removed since the last save, sorted.
func (p *Project) Changes() (written, removed []string) {
	for k := range p.dirty {
		written = append(written, k)
	}
	for k := range p.removed {
		removed = append(removed, k)
	}
	sort.Strings(written)
	sort.Strings(removed)
	return written, removed
}

// clean marks the project saved.
func (p *Project) clean() {
	p.dirty, p.removed, p.meta = map[string]bool{}, map[string]bool{}, false
}

// Clone copies the project under another name, every file to be written.
func (p *Project) Clone(name string) *Project {
	q := NewProject(name)
	q.Saved, q.Tab = p.Saved, p.Tab
	for k, v := range p.UI {
		q.UI[k] = v
	}
	for k, f := range p.files {
		q.files[k] = File{Data: append([]byte(nil), f.Data...), Modified: f.Modified}
		q.dirty[k] = true
	}
	return q
}

// FS is the project as go-slides reads it: its files under dir (the
// directory the deck is built from, "" for the working directory), and
// nothing else: a deck that names ../x or /etc/passwd finds nothing.
func (p *Project) FS(dir string) vfs.FS {
	files := make(map[string][]byte, len(p.files))
	for k, f := range p.files {
		files[filepath.Join(dir, filepath.FromSlash(k))] = f.Data
	}
	return vfs.NewMap(files)
}

// Restore builds a project as a Backend read it: nothing to save.
func Restore(name string, saved time.Time, tab string, ui map[string]string, files map[string]File) *Project {
	p := NewProject(name)
	p.Saved, p.Tab = saved, tab
	for k, v := range ui {
		p.UI[k] = v
	}
	for k, f := range files {
		if c, err := CleanPath(k); err == nil && len(p.files) < MaxFiles && len(f.Data) <= MaxFileSize {
			p.files[c] = f
		}
	}
	p.clean()
	return p
}

// Backend persists projects. Each call answers once through its callback,
// at once or later; calls are carried out in the order they are made.
type Backend interface {
	// Load reads a project: nil and no error when there is none.
	Load(name string, done func(*Project, error))
	// Save writes a project's metadata and the files written, and deletes
	// the files removed, in one transaction.
	Save(p *Project, written, removed []string, done func(error))
	// Copy replaces project to with a copy of project from (to is deleted
	// when from does not exist).
	Copy(from, to string, done func(error))
	// Delete removes a project and its files.
	Delete(name string, done func(error))
}

// ErrNoStore is a save without a backend: a browser that refuses storage.
var ErrNoStore = errors.New("this browser keeps no storage for the page")

// Store is the projects of a page: those it loaded, kept in memory, over a
// Backend (nil: nothing is kept, and saves say so).
type Store struct {
	b     Backend
	tab   string
	cache map[string]*Project
	// known are projects known to exist, with when they were saved, whose
	// contents are not in the cache (copied by the backend).
	known map[string]time.Time
}

// New returns a store over b for the page tab, holding the projects loaded
// before (the page loads them as it starts, where it may wait).
func New(b Backend, tab string, loaded ...*Project) *Store {
	s := &Store{b: b, tab: tab, cache: map[string]*Project{}, known: map[string]time.Time{}}
	for _, p := range loaded {
		if p != nil {
			s.cache[p.Name] = p
		}
	}
	return s
}

// Tab is the page's id.
func (s *Store) Tab() string { return s.tab }

// Cached returns a loaded project, nil if none.
func (s *Store) Cached(name string) *Project { return s.cache[name] }

// Exists reports whether a project exists, and when it was saved.
func (s *Store) Exists(name string) (time.Time, bool) {
	if p, ok := s.cache[name]; ok {
		return p.Saved, true
	}
	t, ok := s.known[name]
	return t, ok
}

// Put makes p the cached project of its name (not saved yet).
func (s *Store) Put(p *Project) {
	s.cache[p.Name] = p
	delete(s.known, p.Name)
}

// Save writes p's changes, stamped with now and this page's tab.
func (s *Store) Save(p *Project, now time.Time, done func(error)) {
	s.Put(p)
	if s.b == nil {
		done(ErrNoStore)
		return
	}
	if !p.Dirty() {
		done(nil)
		return
	}
	p.Saved, p.Tab, p.meta = now, s.tab, true
	written, removed := p.Changes()
	// marked saved now: a change made while the write is out is a new change
	p.clean()
	s.b.Save(p, written, removed, func(err error) {
		if err != nil {
			// what was not written is written next time
			for _, w := range written {
				if _, ok := p.files[w]; ok {
					p.dirty[w] = true
				}
			}
			for _, r := range removed {
				if _, ok := p.files[r]; !ok {
					p.removed[r] = true
				}
			}
			p.meta = true
		}
		done(err)
	})
}

// Copy copies project from to to, in the backend and in the cache.
func (s *Store) Copy(from, to string, saved time.Time, done func(error)) {
	if p, ok := s.cache[from]; ok {
		q := p.Clone(to)
		q.clean()
		s.cache[to] = q
	} else {
		delete(s.cache, to)
		s.known[to] = saved
	}
	if s.b == nil {
		done(ErrNoStore)
		return
	}
	s.b.Copy(from, to, done)
}

// Load reads a project from the backend into the cache.
func (s *Store) Load(name string, done func(*Project, error)) {
	if s.b == nil {
		done(s.cache[name], nil)
		return
	}
	s.b.Load(name, func(p *Project, err error) {
		if err == nil {
			delete(s.known, name)
			if p != nil {
				s.cache[name] = p
			} else {
				delete(s.cache, name)
			}
		}
		done(p, err)
	})
}

// Delete removes a project.
func (s *Store) Delete(name string, done func(error)) {
	delete(s.cache, name)
	delete(s.known, name)
	if s.b == nil {
		done(nil)
		return
	}
	s.b.Delete(name, done)
}
