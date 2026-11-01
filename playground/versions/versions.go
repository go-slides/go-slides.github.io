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
	Git    Kind = "git"    // loaded from a forge (Name says from where)
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
	// Deleted are versions deleted lately: another tab's copy of the index,
	// merged in, does not bring them back.
	Deleted []int `json:"deleted,omitempty"`
}

// maxDeleted bounds the deleted versions remembered.
const maxDeleted = 500

// OrphanAge is how old a blob no version lists must be before a load
// deletes it: younger, it may be another tab's snapshot still being written.
var OrphanAge = 10 * time.Minute

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
	// gone are the blobs this page deleted: another tab's index, merged in,
	// must not list them again.
	gone map[string]bool
	// fresh are the blobs this page wrote that no stored index lists yet.
	fresh map[string]bool
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
		// blobs no version lists, left by a tab that died, then (if a snapshot
		// or a pruning was cut short) what no version uses
		v.sweep(func(err error) {
			if err != nil || (len(v.idx.Pending) == 0 && len(v.idx.Trash) == 0) {
				done(v, err)
				return
			}
			v.idx.Trash = append(v.idx.Trash, v.idx.Pending...)
			v.idx.Pending = nil
			v.collect(func(err error) { done(v, err) })
		})
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
	// an id two tabs are unlikely to share: the time in milliseconds, or
	// past the last id when the clock is behind it
	e := Entry{ID: max(v.idx.Next, int(now.UnixMilli())), Kind: k, Name: name, Time: now, Files: map[string]string{}, Dirs: sortedDirs(s.Dirs), Open: s.Open}
	all := map[string][]byte{} // every distinct content of the version
	for path, b := range s.Files {
		id := BlobID(b)
		e.Files[path] = id
		all[id] = b
	}
	size := 0
	for _, b := range all {
		size += len(b)
	}
	// one version alone must fit; the older automatic ones make room
	if size > v.max() {
		done(Entry{}, ErrTooLarge)
		return
	}
	v.write(e, all, now, true, done)
}

// write stores e: its new blobs, then the index. retry allows one more try
// after dropping automatic versions when the store refuses. all are the
// version's contents; those the index holds are not written again.
func (v *Versions) write(e Entry, all map[string][]byte, now time.Time, retry bool, done func(Entry, error)) {
	v.idx.Next = max(v.idx.Next, e.ID+1)
	missing := func() map[string][]byte {
		out := map[string][]byte{}
		for id, b := range all {
			if _, held := v.idx.Blobs[id]; !held {
				out[id] = b
			}
		}
		return out
	}
	blobs := missing()
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
				v.write(e, all, now, false, done)
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
		// the merge may have learnt that another tab deleted a blob thought
		// held: write it too (not listed as pending: if this tab dies, a
		// later load sweeps it by age)
		blobs = missing()
		ids = slices.Sorted(maps.Keys(blobs))
		v.putBlobs(ids, blobs, func(err error) {
			if err != nil {
				failed(err)
				return
			}
			if v.fresh == nil {
				v.fresh = map[string]bool{}
			}
			for _, id := range ids {
				v.idx.Blobs[id] = len(blobs[id])
				delete(v.gone, id) // written again
				v.fresh[id] = true // stored, not yet listed by the stored index
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
	keep, drop := v.Policy.Keep(v.idx.Entries, now)
	v.idx.Entries = keep
	for _, e := range drop {
		v.idx.Deleted = append(v.idx.Deleted, e.ID)
	}
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
			v.idx.Deleted = append(v.idx.Deleted, e.ID)
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
//
// What may be deleted is decided after the index is written, merged with
// another tab's: a blob that tab's versions still use leaves the trash.
func (v *Versions) collect(done func(error)) {
	v.saveIndex(func(err error) {
		if err != nil {
			done(err)
			return
		}
		u := v.used()
		var trash []string
		for _, b := range v.idx.Trash {
			if !u[b] {
				trash = append(trash, b)
			}
		}
		if len(trash) == 0 {
			if len(v.idx.Trash) > 0 {
				v.idx.Trash = nil // all used again
				v.saveIndex(done)
				return
			}
			done(nil)
			return
		}
		v.deleteBlobs(trash, func(err error) {
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
		if v.gone == nil {
			v.gone = map[string]bool{}
		}
		v.gone[ids[0]] = true
		delete(v.idx.Blobs, ids[0]) // a later version that needs it writes it again
		v.deleteBlobs(ids[1:], done)
	})
}

// saveIndex writes the index, merged first with the one stored: another tab
// of the playground keeps versions in the same store, and writing ours over
// its would lose them (and leave their blobs unlisted).
func (v *Versions) saveIndex(done func(error)) {
	v.st.Load(indexName, func(p *store.Project, err error) {
		if err != nil {
			done(err)
			return
		}
		var other *index
		if p != nil {
			if b, ok := p.Read(indexFile); ok {
				var o index
				if json.Unmarshal(b, &o) == nil {
					other = &o
				}
			}
		} else {
			p = store.NewProject(indexName)
		}
		if other == nil {
			v.writeIndex(p, done)
			return
		}
		// a blob deleted here that the stored copy lists: written again by
		// another tab, or a stale listing (ours, or a tab's that read the
		// index before the deletion)? The store says.
		var check []string
		for id := range other.Blobs {
			if v.gone[id] {
				check = append(check, id)
			}
		}
		v.present(check, func() {
			v.merge(*other)
			v.writeIndex(p, done)
		})
	})
}

// present forgets, among ids (blobs deleted here), those the store holds
// again, so that the merge lists them.
func (v *Versions) present(ids []string, done func()) {
	if len(ids) == 0 {
		done()
		return
	}
	v.st.Load(blobName(ids[0]), func(p *store.Project, err error) {
		if err == nil && p != nil {
			if b, ok := p.Read(blobFile); ok && BlobID(b) == ids[0] {
				delete(v.gone, ids[0])
			}
		}
		v.present(ids[1:], done)
	})
}

func (v *Versions) writeIndex(p *store.Project, done func(error)) {
	b, err := json.Marshal(v.idx)
	if err != nil {
		done(err)
		return
	}
	if err := p.Write(indexFile, b, time.Time{}); err != nil {
		done(err)
		return
	}
	v.st.Save(p, time.Now(), func(err error) {
		if err == nil {
			v.fresh = nil // the stored index lists them now
		}
		done(err)
	})
}

// merge takes into the index what another tab's copy has that ours does
// not: its versions (unless deleted here or there), its blobs (unless
// deleted here), its pending and trash lists.
func (v *Versions) merge(o index) {
	dead := map[int]bool{}
	for _, id := range append(slices.Clone(v.idx.Deleted), o.Deleted...) {
		dead[id] = true
	}
	// a blob we hold that the stored copy does not list was deleted by the
	// tab that wrote it (every tab merges what it found before writing),
	// unless we wrote it and have not listed it yet
	for id := range v.idx.Blobs {
		if _, ok := o.Blobs[id]; !ok && !v.fresh[id] {
			delete(v.idx.Blobs, id)
		}
	}
	// what the stored copy lists, its writer found or wrote: a blob deleted
	// here and written again there is listed again. (The one window left: a
	// tab that read the index before a deletion and writes it after lists a
	// deleted blob; Load then refuses that version, ErrDamaged, and says so.)
	for id, n := range o.Blobs {
		if _, ok := v.idx.Blobs[id]; !ok && !v.gone[id] {
			v.idx.Blobs[id] = n
		}
	}
	have := map[int]bool{}
	for _, e := range v.idx.Entries {
		have[e.ID] = true
	}
	for _, e := range o.Entries {
		if !have[e.ID] && !dead[e.ID] {
			v.idx.Entries = append(v.idx.Entries, e)
		}
	}
	kept := v.idx.Entries[:0]
	for _, e := range v.idx.Entries {
		if !dead[e.ID] {
			kept = append(kept, e)
		}
	}
	v.idx.Entries = kept
	sort.SliceStable(v.idx.Entries, func(i, j int) bool {
		a, b := v.idx.Entries[i], v.idx.Entries[j]
		if !a.Time.Equal(b.Time) {
			return a.Time.Before(b.Time)
		}
		return a.ID < b.ID
	})
	// the other copy's pending and trash lists are its writer's work in
	// progress: not taken (a blob it leaves unlisted is swept by age, see
	// sweep), since a stale list would delete blobs used since
	v.idx.Deleted = slices.Sorted(maps.Keys(dead))
	if len(v.idx.Deleted) > maxDeleted {
		v.idx.Deleted = v.idx.Deleted[len(v.idx.Deleted)-maxDeleted:]
	}
	v.idx.Next = max(v.idx.Next, o.Next)
	v.sanitize()
	// a blob in our trash that a version now listed uses stays
	u := v.used()
	trash := v.idx.Trash[:0]
	for _, b := range v.idx.Trash {
		if !u[b] {
			trash = append(trash, b)
		}
	}
	v.idx.Trash = trash
}

// sweep deletes the blobs the store holds that no version lists and that are
// older than OrphanAge: what a tab that died before listing them left.
func (v *Versions) sweep(done func(error)) {
	v.st.List("v:", func(ls []store.Listed, err error) {
		if err != nil {
			done(err)
			return
		}
		var orphans []string
		for _, l := range ls {
			id := l.Name[len("v:"):]
			if l.Name == indexName || v.used()[id] || slices.Contains(v.idx.Pending, id) {
				continue
			}
			if _, held := v.idx.Blobs[id]; held || time.Since(l.Saved) < OrphanAge {
				continue
			}
			orphans = append(orphans, id)
		}
		v.deleteBlobs(orphans, done)
	})
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
		v.idx.Deleted = append(v.idx.Deleted, id)
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
		for _, e := range v.idx.Entries {
			v.idx.Deleted = append(v.idx.Deleted, e.ID)
		}
		v.idx.Entries = nil
		end := func(err error) {
			finish()
			v.changed()
			done(err)
		}
		// every blob the store holds, this page's and any other's, then the
		// index: nothing of the deck is left
		v.st.List("v:", func(ls []store.Listed, err error) {
			if err != nil {
				end(err)
				return
			}
			var all []string
			for _, l := range ls {
				if l.Name != indexName {
					all = append(all, l.Name[len("v:"):])
				}
			}
			for b := range v.idx.Blobs {
				all = append(all, b)
			}
			slices.Sort(all)
			v.deleteBlobs(slices.Compact(all), func(err error) {
				if err != nil {
					end(err)
					return
				}
				v.idx = index{Next: v.idx.Next, Blobs: map[string]int{}}
				v.st.Delete(indexName, end)
			})
		})
	})
}

func (v *Versions) changed() {
	if v.Changed != nil {
		v.Changed()
	}
}
