// Package versions keeps snapshots of a playground project in the page's
// store (IndexedDB), with no server: automatic ones as the reader works,
// and named ones ("Save version…"), the way an editor's local history (VS
// Code's Timeline) and a document's version history (Google Docs) do.
//
// # Layout in the store
//
// One project, "v:index", holds index.json: the versions (each a map of
// path → blob, the folders and the file shown) and the blobs stored. Each
// distinct file content is one project "v:<sha-256>" holding one file,
// "data": two versions that share a file (the media, a style sheet that did
// not change) share its blob, so a hundred versions of a deck with a 5 MiB
// picture keep the picture once.
//
// # Consistency when the tab dies
//
// Every write is one IndexedDB transaction (a store.Backend call), and the
// calls run in order. A snapshot is written so that a tab killed between any
// two of them leaves the store consistent, or repairable on the next load:
//
//  1. the index, listing the new blobs as pending;
//  2. each new blob;
//  3. the index with the new version (pending cleared), and the blobs no
//     version uses any more listed as trash;
//  4. the trash deleted, then the index without it.
//
// On load, a pending or trashed blob that no version uses is deleted. An
// index is never written that names a blob not yet written, so a version
// that is listed is whole.
//
// # Bounds
//
// Policy keeps the newest automatic versions, then the newest of each hour
// for a day and of each day for a month (VS Code keeps 50 entries a file,
// Google Docs merges old revisions); named versions are kept until deleted,
// at most MaxNamed. All versions together take at most MaxBytes (blobs
// counted once): past it the oldest automatic ones go first, and a store
// that refuses a write (QuotaExceededError) makes the snapshot drop the
// oldest automatic versions and try once more.
package versions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// Kind says why a version was kept.
type Kind string

const (
	Auto   Kind = "auto"   // a while after the typing stopped
	Named  Kind = "named"  // Save version…
	Before Kind = "before" // before an operation on the whole project (Name says which)
	Export Kind = "export" // at an export (Name says the format)
	Share  Kind = "share"  // at a share link
)

// Automatic reports whether the policy may drop versions of this kind.
func (k Kind) Automatic() bool { return k != Named }

// Entry is one version.
type Entry struct {
	ID    int               `json:"id"`
	Kind  Kind              `json:"kind"`
	Name  string            `json:"name,omitempty"`
	Time  time.Time         `json:"time"`
	Files map[string]string `json:"files"` // path → blob
	Dirs  []string          `json:"dirs,omitempty"`
	Open  string            `json:"open,omitempty"`
}

// Snapshot is a project to keep: its files, folders and the file shown.
type Snapshot struct {
	Files map[string][]byte
	Dirs  []string
	Open  string
}

// index is index.json.
type index struct {
	Next    int            `json:"next"`
	Entries []Entry        `json:"entries"` // oldest first
	Blobs   map[string]int `json:"blobs"`   // blob → bytes
	Pending []string       `json:"pending,omitempty"`
	Trash   []string       `json:"trash,omitempty"`
}

// The bounds.
const (
	MaxNamed = 50
	MaxBytes = 96 << 20
)

// Policy is how many automatic versions are kept.
type Policy struct {
	Recent int // the newest, whatever their age
	Hourly int // then the newest of each of the last Hourly hours
	Daily  int // then the newest of each of the last Daily days
}

// DefaultPolicy keeps the 20 newest, one an hour for a day, one a day for a
// month: about 70 automatic versions at most.
var DefaultPolicy = Policy{Recent: 20, Hourly: 24, Daily: 30}

// Keep splits entries (in any order) into those the policy keeps and those
// it drops, both oldest first. Named versions are always kept.
func (p Policy) Keep(entries []Entry, now time.Time) (keep, drop []Entry) {
	es := slices.Clone(entries)
	// newest first; at the same time, the later one first
	sort.SliceStable(es, func(i, j int) bool {
		if !es[i].Time.Equal(es[j].Time) {
			return es[i].Time.After(es[j].Time)
		}
		return es[i].ID > es[j].ID
	})
	recent := 0
	hours, days := map[int64]bool{}, map[string]bool{}
	for _, e := range es {
		if !e.Kind.Automatic() {
			keep = append(keep, e)
			continue
		}
		age := now.Sub(e.Time)
		hour := e.Time.Unix() / 3600
		day := e.Time.Format("2006-01-02")
		switch {
		case recent < p.Recent:
			recent++
		case age < time.Duration(p.Hourly)*time.Hour && !hours[hour]:
		case age < time.Duration(p.Daily)*24*time.Hour && !days[day]:
		default:
			drop = append(drop, e)
			continue
		}
		hours[hour], days[day] = true, true
		keep = append(keep, e)
	}
	slices.Reverse(keep)
	slices.Reverse(drop)
	return keep, drop
}

// Errors.
var (
	ErrNoStore  = errors.New("versions: no store")
	ErrNotFound = errors.New("versions: no such version")
	ErrTooLarge = errors.New("versions: the project is larger than the versions may take")
	ErrNamed    = fmt.Errorf("versions: at most %d named versions", MaxNamed)
	ErrDamaged  = errors.New("versions: a version's file is missing from the store")
)

const (
	indexName = "v:index"
	indexFile = "index.json"
	blobFile  = "data"
)

func blobName(id string) string { return "v:" + id }

// BlobID is the name of a content.
func BlobID(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// Versions is the versions of the page's project.
type Versions struct {
	st     *store.Store
	idx    index
	Policy Policy
	Max    int // the bytes all versions may take (MaxBytes when 0)
	// Changed is called after the list changed.
	Changed func()

	// busy serialises the operations: a store answers later (IndexedDB),
	// and two snapshots must not interleave their index writes.
	busy  bool
	queue []func()
	// pinned are the blobs of the version being written: not yet listed,
	// they must not go to the trash when room is made for it.
	pinned map[string]bool
}

// run starts op now or after the operations before it; op calls finish
// when it is done.
func (v *Versions) run(op func(finish func())) {
	if v.busy {
		v.queue = append(v.queue, func() { v.run(op) })
		return
	}
	v.busy = true
	op(func() {
		v.busy = false
		if len(v.queue) > 0 {
			next := v.queue[0]
			v.queue = v.queue[1:]
			next()
		}
	})
}

// Open loads the versions from st, repairing what a tab that died left.
func Open(st *store.Store, done func(*Versions, error)) {
	v := &Versions{st: st, idx: index{Blobs: map[string]int{}}, Policy: DefaultPolicy}
	if st == nil {
		done(nil, ErrNoStore)
		return
	}
	st.Load(indexName, func(p *store.Project, err error) {
		if err != nil {
			done(nil, err)
			return
		}
		if p != nil {
			if b, ok := p.Read(indexFile); ok {
				if err := json.Unmarshal(b, &v.idx); err != nil {
					// a damaged index: start again, the blobs it named are lost
					v.idx = index{}
				}
			}
		}
		if v.idx.Blobs == nil {
			v.idx.Blobs = map[string]int{}
		}
		v.sanitize()
		if len(v.idx.Pending) == 0 && len(v.idx.Trash) == 0 {
			done(v, nil)
			return
		}
		// a snapshot or a pruning was cut short: drop what no version uses
		v.idx.Trash = append(v.idx.Trash, v.idx.Pending...)
		v.idx.Pending = nil
		v.collect(func(err error) { done(v, err) })
	})
}

// sanitize drops the entries that name a blob the index does not hold (a
// hand-edited or damaged index): a listed version is always whole.
func (v *Versions) sanitize() {
	ok := v.idx.Entries[:0]
	for _, e := range v.idx.Entries {
		whole := e.Files != nil
		for _, b := range e.Files {
			if _, held := v.idx.Blobs[b]; !held {
				whole = false
			}
		}
		if whole {
			ok = append(ok, e)
		}
		v.idx.Next = max(v.idx.Next, e.ID+1)
	}
	v.idx.Entries = ok
}

// Entries are the versions, newest first.
func (v *Versions) Entries() []Entry {
	out := slices.Clone(v.idx.Entries)
	slices.Reverse(out)
	return out
}

// Entry returns a version by id.
func (v *Versions) Entry(id int) (Entry, bool) {
	for _, e := range v.idx.Entries {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}

// Size is the bytes a version's files take.
func (v *Versions) Size(e Entry) int {
	n := 0
	for _, b := range e.Files {
		n += v.idx.Blobs[b]
	}
	return n
}

// Bytes is what all the versions take (each blob once).
func (v *Versions) Bytes() int {
	n := 0
	for _, s := range v.idx.Blobs {
		n += s
	}
	return n
}

func (v *Versions) max() int {
	if v.Max > 0 {
		return v.Max
	}
	return MaxBytes
}

// Same reports whether s holds exactly the files and folders of version e.
func Same(e Entry, s Snapshot) bool {
	if len(e.Files) != len(s.Files) || !slices.Equal(e.Dirs, sortedDirs(s.Dirs)) {
		return false
	}
	for k, b := range s.Files {
		if e.Files[k] != BlobID(b) {
			return false
		}
	}
	return true
}

func sortedDirs(d []string) []string {
	d = slices.Clone(d)
	slices.Sort(d)
	return d
}

// Latest is the newest version, if any.
func (v *Versions) Latest() (Entry, bool) {
	if len(v.idx.Entries) == 0 {
		return Entry{}, false
	}
	return v.idx.Entries[len(v.idx.Entries)-1], true
}

// Save keeps s as a new version of kind k (name: a named version's name, or
// what an automatic one was kept for). An automatic version identical to the
// newest one is not kept again: done gets the newest one. A store that is
// full makes the oldest automatic versions go, and the write is tried once
// more.
func (v *Versions) Save(s Snapshot, k Kind, name string, now time.Time, done func(Entry, error)) {
	v.run(func(finish func()) {
		v.save(s, k, name, now, func(e Entry, err error) {
			finish()
			done(e, err)
		})
	})
}

func (v *Versions) save(s Snapshot, k Kind, name string, now time.Time, done func(Entry, error)) {
	if last, ok := v.Latest(); ok && k.Automatic() && Same(last, s) {
		done(last, nil)
		return
	}
	if k == Named {
		named := 0
		for _, e := range v.idx.Entries {
			if e.Kind == Named {
				named++
			}
		}
		if named >= MaxNamed {
			done(Entry{}, ErrNamed)
			return
		}
	}
	e := Entry{ID: v.idx.Next, Kind: k, Name: name, Time: now, Files: map[string]string{}, Dirs: sortedDirs(s.Dirs), Open: s.Open}
	blobs := map[string][]byte{}
	distinct := map[string]int{}
	for path, b := range s.Files {
		id := BlobID(b)
		e.Files[path] = id
		distinct[id] = len(b)
		if _, held := v.idx.Blobs[id]; !held {
			blobs[id] = b
		}
	}
	size := 0
	for _, n := range distinct {
		size += n
	}
	// one version alone must fit; the older automatic ones make room
	if size > v.max() {
		done(Entry{}, ErrTooLarge)
		return
	}
	v.write(e, blobs, now, true, done)
}

// write stores e: its new blobs, then the index. retry allows one more try
// after dropping automatic versions when the store refuses.
func (v *Versions) write(e Entry, blobs map[string][]byte, now time.Time, retry bool, done func(Entry, error)) {
	v.idx.Next = max(v.idx.Next, e.ID+1)
	ids := slices.Sorted(maps.Keys(blobs))
	v.idx.Pending = ids
	v.pinned = map[string]bool{}
	for _, b := range e.Files {
		v.pinned[b] = true
	}
	failed := func(err error) {
		if retry && v.dropOldestAuto(now) {
			// make the room first (delete what they alone used), then try again
			v.collect(func(cerr error) {
				if cerr != nil {
					v.pinned = nil
					done(Entry{}, err)
					return
				}
				v.write(e, blobs, now, false, done)
			})
			return
		}
		// the blobs written for nothing go with the next collection (or,
		// if the index listing them as pending was written, the next load)
		v.pinned = nil
		v.idx.Pending = nil
		v.idx.Trash = append(v.idx.Trash, ids...)
		done(Entry{}, err)
	}
	v.saveIndex(func(err error) {
		if err != nil {
			failed(err)
			return
		}
		v.putBlobs(ids, blobs, func(err error) {
			if err != nil {
				failed(err)
				return
			}
			for _, id := range ids {
				v.idx.Blobs[id] = len(blobs[id])
			}
			v.idx.Pending = nil
			v.idx.Entries = append(v.idx.Entries, e)
			v.pinned = nil
			v.prune(now)
			v.collect(func(err error) {
				v.changed()
				done(e, err)
			})
		})
	})
}

// putBlobs writes the blobs in order.
func (v *Versions) putBlobs(ids []string, blobs map[string][]byte, done func(error)) {
	if len(ids) == 0 {
		done(nil)
		return
	}
	p := store.NewProject(blobName(ids[0]))
	if err := p.Write(blobFile, blobs[ids[0]], time.Time{}); err != nil {
		done(err)
		return
	}
	v.st.Save(p, time.Now(), func(err error) {
		if err != nil {
			done(err)
			return
		}
		v.putBlobs(ids[1:], blobs, done)
	})
}

// prune applies the policy and the byte bound: the entries dropped, and the
// blobs no entry uses any more go to the trash.
func (v *Versions) prune(now time.Time) {
	keep, _ := v.Policy.Keep(v.idx.Entries, now)
	v.idx.Entries = keep
	for v.referencedBytes() > v.max() && v.dropOneAuto() {
	}
	v.trashUnused()
}

// dropOldestAuto drops a quarter of the automatic versions (at least one),
// oldest first, for a store that refused a write; it reports whether it
// dropped any.
func (v *Versions) dropOldestAuto(now time.Time) bool {
	n := 0
	for _, e := range v.idx.Entries {
		if e.Kind.Automatic() {
			n++
		}
	}
	if n == 0 {
		return false
	}
	for i := 0; i < max(n/4, 1); i++ {
		v.dropOneAuto()
	}
	v.trashUnused()
	return true
}

// dropOneAuto drops the oldest automatic version.
func (v *Versions) dropOneAuto() bool {
	for i, e := range v.idx.Entries {
		if e.Kind.Automatic() {
			v.idx.Entries = slices.Delete(v.idx.Entries, i, i+1)
			return true
		}
	}
	return false
}

func (v *Versions) used() map[string]bool {
	u := maps.Clone(v.pinned)
	if u == nil {
		u = map[string]bool{}
	}
	for _, e := range v.idx.Entries {
		for _, b := range e.Files {
			u[b] = true
		}
	}
	return u
}

func (v *Versions) referencedBytes() int {
	n := 0
	for b := range v.used() {
		n += v.idx.Blobs[b]
	}
	return n
}

// trashUnused moves the blobs no entry uses to the trash.
func (v *Versions) trashUnused() {
	u := v.used()
	for b := range v.idx.Blobs {
		if !u[b] {
			delete(v.idx.Blobs, b)
			v.idx.Trash = append(v.idx.Trash, b)
		}
	}
	slices.Sort(v.idx.Trash)
	v.idx.Trash = slices.Compact(v.idx.Trash)
}

// collect writes the index (with its trash), deletes the trash, and writes
// the index again without it.
func (v *Versions) collect(done func(error)) {
	u := v.used()
	trash := v.idx.Trash[:0]
	for _, b := range v.idx.Trash {
		if !u[b] {
			trash = append(trash, b)
		}
	}
	v.idx.Trash = trash
	v.saveIndex(func(err error) {
		if err != nil || len(trash) == 0 {
			done(err)
			return
		}
		v.deleteBlobs(slices.Clone(trash), func(err error) {
			if err != nil {
				done(err)
				return
			}
			v.idx.Trash = nil
			v.saveIndex(done)
		})
	})
}

func (v *Versions) deleteBlobs(ids []string, done func(error)) {
	if len(ids) == 0 {
		done(nil)
		return
	}
	v.st.Delete(blobName(ids[0]), func(err error) {
		if err != nil {
			done(err)
			return
		}
		v.deleteBlobs(ids[1:], done)
	})
}

func (v *Versions) saveIndex(done func(error)) {
	b, err := json.Marshal(v.idx)
	if err != nil {
		done(err)
		return
	}
	p := v.st.Cached(indexName)
	if p == nil {
		p = store.NewProject(indexName)
	}
	if err := p.Write(indexFile, b, time.Time{}); err != nil {
		done(err)
		return
	}
	v.st.Save(p, time.Now(), done)
}

// Load reads a version's files.
func (v *Versions) Load(id int, done func(Snapshot, Entry, error)) {
	e, ok := v.Entry(id)
	if !ok {
		done(Snapshot{}, Entry{}, ErrNotFound)
		return
	}
	s := Snapshot{Files: map[string][]byte{}, Dirs: slices.Clone(e.Dirs), Open: e.Open}
	paths := slices.Sorted(maps.Keys(e.Files))
	var next func(i int)
	next = func(i int) {
		if i == len(paths) {
			done(s, e, nil)
			return
		}
		v.st.Load(blobName(e.Files[paths[i]]), func(p *store.Project, err error) {
			if err != nil {
				done(Snapshot{}, e, err)
				return
			}
			b, ok := []byte(nil), false
			if p != nil {
				b, ok = p.Read(blobFile)
			}
			if !ok || BlobID(b) != e.Files[paths[i]] {
				done(Snapshot{}, e, ErrDamaged)
				return
			}
			s.Files[paths[i]] = b
			next(i + 1)
		})
	}
	next(0)
}

// Delete removes a version (and the blobs only it used).
func (v *Versions) Delete(id int, done func(error)) {
	v.run(func(finish func()) {
		i := slices.IndexFunc(v.idx.Entries, func(e Entry) bool { return e.ID == id })
		if i < 0 {
			finish()
			done(ErrNotFound)
			return
		}
		v.idx.Entries = slices.Delete(v.idx.Entries, i, i+1)
		v.trashUnused()
		v.collect(func(err error) {
			finish()
			v.changed()
			done(err)
		})
	})
}

// Clear removes every version and every blob: what "Forget my drafts" does,
// so that nothing of the deck stays in the browser. The index goes last, so
// a tab killed on the way leaves blobs the next load lists as unused and
// deletes.
func (v *Versions) Clear(done func(error)) {
	v.run(func(finish func()) {
		v.idx.Entries = nil
		v.trashUnused()
		end := func(err error) {
			finish()
			v.changed()
			done(err)
		}
		v.collect(func(err error) {
			if err != nil {
				end(err)
				return
			}
			v.idx = index{Next: v.idx.Next, Blobs: map[string]int{}}
			v.st.Delete(indexName, end)
		})
	})
}

func (v *Versions) changed() {
	if v.Changed != nil {
		v.Changed()
	}
}
