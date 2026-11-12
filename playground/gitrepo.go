package playground

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-widgets/toolkit"
	"github.com/go-widgets/vcs"
)

// The project as a version-controlled working tree.
//
// What the file explorer marks (toolkit.FileTree.SetRepository) is asked of a
// vcs.Repository, the one interface a tree needs, and the playground has no tree
// of its own to keep current: the folder roll-up (a folder is Modified when
// anything under it is, Conflicted wins), the staleness and the threading are the
// tree's. This file answers the question for a project that is not on a disk:
//
//   - the working copy is the project (captureState);
//   - the reference is the base the pane loaded: the hash of every file as it was
//     when Load (or Take theirs) read it from the forge, kept in the UI state so
//     a reload still knows. No content is kept for it: the versions hold that;
//   - a path the base does not have is Added, one whose content differs is
//     Modified, one the base has and the project lost is Deleted;
//   - Check for changes (refreshGit) leaves what the branch holds now in
//     vm.git.theirs: a path that changed there AND here, to different content, is
//     Conflicted (what Take theirs would overwrite). Where both sides made the
//     same change there is nothing to resolve, and the path keeps its local state.
//
// With no base (nothing loaded from a forge) the project is not under version
// control and nothing is marked.

// GitRepository is the project as a vcs.Repository, for toolkit.FileTree.SetRepository.
// It is one repository for the view-model's life: the hashes it has computed are
// kept in it.
func (vm *ViewModel) GitRepository() vcs.Repository {
	if vm.repo == nil {
		vm.repo = &projectRepo{vm: vm, cache: map[sliceID]string{}}
		vm.repo.capture()
	}
	return vm.repo
}

// captureRepo says the project, the base or theirs changed: the repository takes
// a new snapshot of them (UI goroutine), before anything asks for the status.
func (vm *ViewModel) captureRepo() {
	if vm.repo != nil {
		vm.repo.capture()
	}
}

// projectRepo is the project and the base it was loaded from.
//
// ⛔ Threading. The tree may ask Status off the UI goroutine (it does when it is
// given a PostUI, scene.go), while the view-model is edited on it. So Status never
// reads the view-model: it reads a snapshot, taken on the UI goroutine by capture
// whenever the project, the base or theirs changes, and published atomically. A
// snapshot is never written after it is published: its maps are fresh, and the
// file contents in them are the project's stored slices, which the store replaces
// and never changes in place.
//
// That is also what makes the hashes cacheable. A stored file is a slice that is
// replaced (a fresh copy) on every write, so the slice itself is the file's
// version: hashes are kept by the identity of the slice's array and its length,
// and a refresh hashes only what was written since the last one (the deck, which
// is rebuilt from the editor's text on each capture, among them). Entries the
// latest status did not use are dropped, so the cache holds the project's files
// and no old copy.
type projectRepo struct {
	vm   *ViewModel
	snap atomic.Pointer[repoSnap]

	mu     sync.Mutex
	cache  map[sliceID]string
	hashed atomic.Int64 // contents hashed so far, for the tests
}

// repoSnap is what Status reads.
type repoSnap struct {
	base   map[string]string // the base's hashes; nil: not under version control
	work   map[string][]byte
	theirs remote.Files
}

// sliceID is a stored file's version: where its content is, and how long.
type sliceID struct {
	p *byte
	n int
}

func idOf(b []byte) sliceID { return sliceID{unsafe.SliceData(b), len(b)} }

// capture takes the snapshot (UI goroutine only).
func (r *projectRepo) capture() {
	g := &r.vm.git
	if g.base == nil || g.baseSums == nil {
		r.snap.Store(&repoSnap{})
		return
	}
	r.snap.Store(&repoSnap{base: g.baseSums, work: r.vm.captureState().Files, theirs: g.theirs})
}

// Root is "/": the project lives in the page, not under a directory.
func (*projectRepo) Root() string { return "/" }

// Status is the working copy against the base, and against theirs when Check for
// changes found some, as of the latest snapshot.
func (r *projectRepo) Status(ctx context.Context) (map[string]vcs.State, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s := r.snap.Load()
	if s == nil || s.base == nil {
		return map[string]vcs.State{}, nil
	}
	used := map[sliceID]bool{}
	out, err := diffStatusWith(ctx, s.base, s.work, s.theirs, func(b []byte) string {
		id := idOf(b)
		used[id] = true
		r.mu.Lock()
		sum, ok := r.cache[id]
		r.mu.Unlock()
		if !ok {
			sum = sumOf(b)
			r.hashed.Add(1)
			r.mu.Lock()
			r.cache[id] = sum
			r.mu.Unlock()
		}
		return sum
	})
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	for id := range r.cache {
		if !used[id] {
			delete(r.cache, id)
		}
	}
	r.mu.Unlock()
	return out, nil
}

// diffStatus is diffStatusWith, hashing every content.
func diffStatus(ctx context.Context, base map[string]string, work map[string][]byte, theirs remote.Files) (map[string]vcs.State, error) {
	return diffStatusWith(ctx, base, work, theirs, sumOf)
}

// diffStatusWith is the status of work against base, and of the conflicts with
// theirs (nil: none found). base holds a hash per path; work and theirs hold
// content, hashed by sum.
func diffStatusWith(ctx context.Context, base map[string]string, work map[string][]byte, theirs remote.Files, sum func([]byte) string) (map[string]vcs.State, error) {
	out := map[string]vcs.State{}
	for p, b := range work {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if s, ok := base[p]; !ok {
			out[p] = vcs.Added
		} else if sum(b) != s {
			out[p] = vcs.Modified
		}
	}
	for p := range base {
		if _, ok := work[p]; !ok {
			out[p] = vcs.Deleted
		}
	}
	if theirs == nil {
		return out, nil
	}
	for p := range out {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		t, inTheirs := theirs[p]
		s, inBase := base[p]
		if inTheirs == inBase && (!inTheirs || sum(t) == s) {
			continue // the branch did not touch it: only we did
		}
		if w, inWork := work[p]; inWork == inTheirs && (!inWork || bytes.Equal(w, t)) {
			continue // both made the same change
		}
		out[p] = vcs.Conflicted
	}
	return out, nil
}

// sumOf is the hash recorded for a file's content.
func sumOf(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// sumsOf is the hash of each file of a folder read from the forge.
func sumsOf(files remote.Files) map[string]string {
	out := make(map[string]string, len(files))
	for p, b := range files {
		out[p] = sumOf(b)
	}
	return out
}

// parseSums reads the hashes back from the UI state. Anything that is not a
// bounded map of hex hashes is dropped (nil: the base is then unknown, and
// nothing is marked), as a base that does not parse is.
func parseSums(j string) map[string]string {
	if j == "" {
		return nil
	}
	var m map[string]string
	if json.Unmarshal([]byte(j), &m) != nil || len(m) > 4096 {
		return nil
	}
	for _, s := range m {
		if raw, err := hex.DecodeString(s); err != nil || len(raw) != sha256.Size {
			return nil
		}
	}
	return m
}

// deletedFromBase are the paths the base had and the project lost, sorted: the
// explorer lists them (struck through, by the tree's own Deleted mark) because a
// tree shows the rows it is given and invents none.
func (vm *ViewModel) deletedFromBase() []string {
	if vm.git.base == nil || vm.git.baseSums == nil {
		return nil
	}
	have := map[string]bool{}
	for _, f := range vm.Files.Get() {
		have[f.Path] = true
	}
	var out []string
	for p := range vm.git.baseSums {
		if !have[p] {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// dropBase forgets what the project was loaded from.
func (vm *ViewModel) dropBase() {
	vm.git.base, vm.git.baseSums, vm.git.theirs, vm.git.theirsBase = nil, nil, nil, nil
	vm.baseChanged()
}

// baseChanged says the reference of the status (the base, or what Check for
// changes found) changed: the explorer re-reads the status.
func (vm *ViewModel) baseChanged() {
	vm.captureRepo()
	vm.GitRev.Set(vm.GitRev.Get() + 1)
}

// treeEntries are the view-model's files as the toolkit's, and the files the base
// had and the project lost (deleted lists their paths).
func treeEntries(fs []ProjectFile, deleted []string) []toolkit.FileEntry {
	out := make([]toolkit.FileEntry, 0, len(fs)+len(deleted))
	for _, f := range fs {
		out = append(out, toolkit.FileEntry{Path: f.Path, Dir: f.Dir, Size: f.Size})
	}
	for _, p := range deleted {
		out = append(out, toolkit.FileEntry{Path: p})
	}
	return out
}
