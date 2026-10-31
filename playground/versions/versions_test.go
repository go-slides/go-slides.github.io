package versions

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

var t0 = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

func open(t *testing.T, b store.Backend) *Versions {
	t.Helper()
	var v *Versions
	var err error
	Open(store.New(b, "tab"), func(x *Versions, e error) { v, err = x, e })
	if err != nil || v == nil {
		t.Fatalf("open: %v", err)
	}
	return v
}

func snap(deck string, extra ...string) Snapshot {
	s := Snapshot{Files: map[string][]byte{"deck.md": []byte(deck)}, Open: "deck.md"}
	for i := 0; i+1 < len(extra); i += 2 {
		s.Files[extra[i]] = []byte(extra[i+1])
	}
	return s
}

func save(t *testing.T, v *Versions, s Snapshot, k Kind, name string, at time.Time) Entry {
	t.Helper()
	var e Entry
	var err error
	v.Save(s, k, name, at, func(x Entry, y error) { e, err = x, y })
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	return e
}

func load(t *testing.T, v *Versions, id int) Snapshot {
	t.Helper()
	var s Snapshot
	var err error
	v.Load(id, func(x Snapshot, _ Entry, y error) { s, err = x, y })
	if err != nil {
		t.Fatalf("load %d: %v", id, err)
	}
	return s
}

// consistent checks the store against the index: every version's blobs are
// there with the right bytes, and no blob is there that no version uses.
func consistent(t *testing.T, m *store.Memory, v *Versions) {
	t.Helper()
	used := v.used()
	for _, e := range v.idx.Entries {
		s := load(t, v, e.ID)
		for p, b := range s.Files {
			if BlobID(b) != e.Files[p] {
				t.Fatalf("version %d: %s has the wrong bytes", e.ID, p)
			}
		}
	}
	for _, n := range m.Names() {
		if id, ok := strings.CutPrefix(n, "v:"); ok && id != "index" && !used[id] {
			t.Fatalf("blob %s is stored and no version uses it", id[:8])
		}
	}
}

func TestSaveLoadDedupDelete(t *testing.T) {
	m := store.NewMemory()
	v := open(t, m)
	pic := strings.Repeat("P", 5000)
	a := save(t, v, snap("one", "media/p.png", pic), Auto, "", t0)
	// the same project again: not kept twice
	if again := save(t, v, snap("one", "media/p.png", pic), Auto, "", t0.Add(time.Minute)); again.ID != a.ID {
		t.Fatal("an unchanged project was kept again")
	}
	b := save(t, v, snap("two", "media/p.png", pic), Named, "Draft for Ada", t0.Add(2*time.Minute))
	if len(v.Entries()) != 2 || v.Entries()[0].ID != b.ID {
		t.Fatal("Entries are not newest first")
	}
	// the picture is kept once
	if got := v.Bytes(); got != 5000+3+3 {
		t.Fatalf("bytes %d", got)
	}
	if v.Size(b) != 5003 {
		t.Fatalf("size %d", v.Size(b))
	}
	if s := load(t, v, a.ID); string(s.Files["deck.md"]) != "one" || string(s.Files["media/p.png"]) != pic {
		t.Fatal("version one")
	}
	// another page load reads the same
	w := open(t, m)
	if len(w.Entries()) != 2 || w.Entries()[0].Name != "Draft for Ada" {
		t.Fatal("not kept across loads")
	}
	if !Same(b, snap("two", "media/p.png", pic)) || Same(b, snap("two")) {
		t.Fatal("Same")
	}
	var err error
	w.Delete(a.ID, func(e error) { err = e })
	if err != nil || len(w.Entries()) != 1 {
		t.Fatal("delete")
	}
	consistent(t, m, w)
	w.Delete(a.ID, func(e error) { err = e })
	if err != ErrNotFound {
		t.Fatal("deleting twice")
	}
	w.Load(99, func(_ Snapshot, _ Entry, e error) { err = e })
	if err != ErrNotFound {
		t.Fatal("loading nothing")
	}
}

func TestPolicyKeepsRecentHourlyDailyAndNamed(t *testing.T) {
	var es []Entry
	now := t0
	// one automatic version every 10 minutes for 40 days, and a named one a week
	for i := 0; i < 40*24*6; i++ {
		at := now.Add(-time.Duration(i) * 10 * time.Minute)
		k := Auto
		if i%(7*24*6) == 0 {
			k = Named
		}
		es = append(es, Entry{ID: i, Kind: k, Time: at})
	}
	keep, drop := DefaultPolicy.Keep(es, now)
	if len(keep)+len(drop) != len(es) {
		t.Fatal("lost entries")
	}
	named, auto := 0, 0
	for _, e := range keep {
		if e.Kind == Named {
			named++
		} else {
			auto++
		}
	}
	if named != 6 {
		t.Fatalf("named kept %d", named)
	}
	// 20 recent (3h20), then hourly to 24 h (about 21 more), daily to 30 days (about 29)
	if auto < 65 || auto > 75 {
		t.Fatalf("automatic kept %d", auto)
	}
	for i := 1; i < len(keep); i++ {
		if keep[i].Time.Before(keep[i-1].Time) {
			t.Fatal("keep not oldest first")
		}
	}
	for _, e := range drop {
		if e.Kind == Named {
			t.Fatal("a named version dropped")
		}
	}
}

func TestPruningAndByteBound(t *testing.T) {
	m := store.NewMemory()
	v := open(t, m)
	v.Max = 20_000
	v.Policy = Policy{Recent: 3}
	for i := 0; i < 10; i++ {
		save(t, v, snap(fmt.Sprintf("deck %d %s", i, strings.Repeat("x", 3000))), Auto, "", t0.Add(time.Duration(i)*time.Second))
	}
	if n := len(v.Entries()); n != 3 {
		t.Fatalf("kept %d, want 3", n)
	}
	consistent(t, m, v)
	// the byte bound drops the oldest automatic ones first, never a named one
	v.Policy = Policy{Recent: 100}
	save(t, v, snap("named "+strings.Repeat("n", 9000)), Named, "keep me", t0.Add(time.Hour))
	for i := 0; i < 5; i++ {
		save(t, v, snap(fmt.Sprintf("big %d %s", i, strings.Repeat("y", 4000))), Auto, "", t0.Add(2*time.Hour+time.Duration(i)*time.Second))
	}
	if v.referencedBytes() > v.Max {
		t.Fatalf("%d bytes over %d", v.referencedBytes(), v.Max)
	}
	if v.Entries()[len(v.Entries())-1].Kind != Named && !hasNamed(v) {
		t.Fatal("the named version went")
	}
	consistent(t, m, v)
	// one version larger than all may take is refused
	var err error
	v.Save(snap(strings.Repeat("z", 30_000)), Auto, "", t0.Add(3*time.Hour), func(_ Entry, e error) { err = e })
	if err != ErrTooLarge {
		t.Fatalf("err %v", err)
	}
}

func hasNamed(v *Versions) bool {
	for _, e := range v.Entries() {
		if e.Kind == Named {
			return true
		}
	}
	return false
}

func TestNamedLimit(t *testing.T) {
	v := open(t, store.NewMemory())
	for i := 0; i < MaxNamed; i++ {
		save(t, v, snap(fmt.Sprint(i)), Named, fmt.Sprint(i), t0.Add(time.Duration(i)*time.Second))
	}
	var err error
	v.Save(snap("one more"), Named, "x", t0.Add(time.Hour), func(_ Entry, e error) { err = e })
	if err != ErrNamed {
		t.Fatalf("err %v", err)
	}
}

// A full store: the oldest automatic versions go and the write is tried
// again; a store with nothing left to drop says it is full.
func TestQuota(t *testing.T) {
	m := store.NewMemory()
	v := open(t, m)
	for i := 0; i < 8; i++ {
		save(t, v, snap(fmt.Sprintf("%d %s", i, strings.Repeat("q", 2000))), Auto, "", t0.Add(time.Duration(i)*time.Minute))
	}
	m.Quota = m.Bytes() + 1500 // room for less than one more
	before := len(v.Entries())
	e := save(t, v, snap("new "+strings.Repeat("r", 2000)), Named, "after the quota", t0.Add(time.Hour))
	if len(v.Entries()) >= before+1 || e.Name != "after the quota" {
		t.Fatalf("entries %d -> %d", before, len(v.Entries()))
	}
	consistent(t, m, v)
	// nothing automatic left to drop, and still no room
	w := open(t, m)
	for _, x := range w.Entries() {
		if x.Kind.Automatic() {
			w.Delete(x.ID, func(error) {})
		}
	}
	m.Quota = m.Bytes() + 10
	var err error
	w.Save(snap(strings.Repeat("s", 5000)), Named, "no room", t0.Add(2*time.Hour), func(_ Entry, e error) { err = e })
	if err != store.ErrQuota {
		t.Fatalf("err %v", err)
	}
	m.Quota = 0
	x := open(t, m)
	consistent(t, m, x)
}

// killer is a Backend that stops answering, and stops writing, after n
// calls: a tab killed in the middle of a snapshot.
type killer struct {
	m *store.Memory
	n int
}

func (k *killer) alive() bool { k.n--; return k.n >= 0 }
func (k *killer) Load(name string, done func(*store.Project, error)) {
	if k.alive() {
		k.m.Load(name, done)
	}
}
func (k *killer) Save(p *store.Project, w, r []string, done func(error)) {
	if k.alive() {
		k.m.Save(p, w, r, done)
	}
}
func (k *killer) Copy(f, t string, done func(error)) {
	if k.alive() {
		k.m.Copy(f, t, done)
	}
}
func (k *killer) Delete(name string, done func(error)) {
	if k.alive() {
		k.m.Delete(name, done)
	}
}

// Chaos: the tab is killed after every possible number of store calls in
// the middle of snapshots and deletions; the next load always finds a
// consistent store, and the versions it lists load whole.
func TestKilledMidSnapshotLeavesTheStoreConsistent(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for kill := 0; kill < 60; kill++ {
		m := store.NewMemory()
		// some history first, intact
		v := open(t, m)
		v.Policy = Policy{Recent: 4}
		for i := 0; i < 5; i++ {
			save(t, v, snap(fmt.Sprint("base ", i), "media/a.png", fmt.Sprint("pic ", i%2)), Auto, "", t0.Add(time.Duration(i)*time.Minute))
		}
		// then a tab that dies after `kill` calls, while it saves and deletes
		k := &killer{m: m, n: kill}
		var w *Versions
		Open(store.New(k, "dying"), func(x *Versions, _ error) { w = x })
		if w != nil {
			w.Policy = Policy{Recent: 4}
			for i := 0; i < 6; i++ {
				extra := []string{"media/a.png", fmt.Sprint("pic ", r.Intn(4)), "themes/t.css", fmt.Sprint("css ", r.Intn(3))}
				w.Save(snap(fmt.Sprint("dying ", i), extra...), Auto, "", t0.Add(time.Hour+time.Duration(i)*time.Minute), func(Entry, error) {})
				if i == 3 {
					if es := w.Entries(); len(es) > 1 {
						w.Delete(es[len(es)-1].ID, func(error) {})
					}
				}
			}
		}
		// the next page load
		next := open(t, m)
		consistent(t, m, next)
		if len(next.Entries()) == 0 {
			t.Fatalf("kill %d: every version lost", kill)
		}
	}
}

// A damaged index (cut, or not JSON) is started again; an entry that names
// a blob the index does not hold is dropped; a blob with the wrong bytes is
// reported, not shown.
func TestDamagedIndexAndBlob(t *testing.T) {
	m := store.NewMemory()
	st := store.New(m, "tab")
	p := store.NewProject(indexName)
	_ = p.Write(indexFile, []byte("{not json"), t0)
	st.Save(p, t0, func(error) {})
	v := open(t, m)
	if len(v.Entries()) != 0 {
		t.Fatal("entries from a damaged index")
	}
	e := save(t, v, snap("ok"), Auto, "", t0)
	// an entry naming a blob that is not held
	v.idx.Entries = append(v.idx.Entries, Entry{ID: 50, Kind: Auto, Time: t0, Files: map[string]string{"deck.md": "feedbeef"}})
	v.saveIndex(func(error) {})
	w := open(t, m)
	if len(w.Entries()) != 1 || w.idx.Next <= e.ID {
		t.Fatalf("entries %d next %d", len(w.Entries()), w.idx.Next)
	}
	// the blob rewritten with other bytes
	bp := store.NewProject(blobName(e.Files["deck.md"]))
	_ = bp.Write(blobFile, []byte("tampered"), t0)
	st.Save(bp, t0, func(error) {})
	var err error
	w.Load(e.ID, func(_ Snapshot, _ Entry, x error) { err = x })
	if err != ErrDamaged {
		t.Fatalf("err %v", err)
	}
	if _, ok := w.Latest(); !ok {
		t.Fatal("latest")
	}
	var nilErr error
	Open(nil, func(_ *Versions, x error) { nilErr = x })
	if nilErr != ErrNoStore {
		t.Fatal("no store")
	}
	if !bytes.Equal([]byte(BlobID(nil)), []byte(BlobID([]byte{}))) {
		t.Fatal("BlobID")
	}
}

// Clear leaves nothing of the versions in the store, and the versions go on.
func TestClear(t *testing.T) {
	m := store.NewMemory()
	v := open(t, m)
	save(t, v, snap("a", "media/p.png", "pic"), Named, "a", t0)
	save(t, v, snap("b"), Auto, "", t0.Add(time.Minute))
	var err error
	v.Clear(func(e error) { err = e })
	if err != nil || len(v.Entries()) != 0 || len(m.Names()) != 0 {
		t.Fatalf("err %v entries %d left %v", err, len(v.Entries()), m.Names())
	}
	e := save(t, v, snap("c"), Auto, "", t0.Add(time.Hour))
	if w := open(t, m); len(w.Entries()) != 1 || w.Entries()[0].ID != e.ID {
		t.Fatal("not usable after Clear")
	}
}

// A store that refuses everything: each operation says so, nothing panics.
func TestBlockedStore(t *testing.T) {
	m := store.NewMemory()
	v := open(t, m)
	e := save(t, v, snap("kept"), Named, "kept", t0)
	changed := 0
	v.Changed = func() { changed++ }
	m.Blocked = true
	var err error
	Open(store.New(m, "tab"), func(_ *Versions, x error) { err = x })
	if err == nil {
		t.Fatal("open")
	}
	v.Save(snap("more"), Auto, "", t0.Add(time.Minute), func(_ Entry, x error) { err = x })
	if err == nil {
		t.Fatal("save")
	}
	v.Delete(e.ID, func(x error) { err = x })
	if err == nil || changed == 0 {
		t.Fatal("delete")
	}
	v.Load(e.ID, func(_ Snapshot, _ Entry, x error) { err = x })
	if err == nil {
		t.Fatal("load")
	}
}

// Operations started while one is out run after it, in order.
func TestOperationsAreSerialised(t *testing.T) {
	m := store.NewMemory()
	slow := &deferred{m: m}
	var v *Versions
	Open(store.New(slow, "tab"), func(x *Versions, _ error) { v = x })
	slow.flush()
	var order []string
	v.Save(snap("a"), Auto, "", t0, func(e Entry, _ error) { order = append(order, "a") })
	v.Save(snap("b"), Named, "b", t0.Add(time.Second), func(e Entry, _ error) { order = append(order, "b") })
	if len(order) != 0 {
		t.Fatal("answered before the store")
	}
	slow.flush()
	if strings.Join(order, "") != "ab" || len(v.Entries()) != 2 {
		t.Fatalf("order %v entries %d", order, len(v.Entries()))
	}
}

// deferred answers each call only when flushed, as IndexedDB answers later.
type deferred struct {
	m       *store.Memory
	pending []func()
}

func (d *deferred) flush() {
	for len(d.pending) > 0 {
		f := d.pending[0]
		d.pending = d.pending[1:]
		f()
	}
}
func (d *deferred) Load(n string, done func(*store.Project, error)) {
	d.pending = append(d.pending, func() { d.m.Load(n, done) })
}
func (d *deferred) Save(p *store.Project, w, r []string, done func(error)) {
	d.pending = append(d.pending, func() { d.m.Save(p, w, r, done) })
}
func (d *deferred) Copy(f, t string, done func(error)) {
	d.pending = append(d.pending, func() { d.m.Copy(f, t, done) })
}
func (d *deferred) Delete(n string, done func(error)) {
	d.pending = append(d.pending, func() { d.m.Delete(n, done) })
}
