package playground

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

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
func (vm *ViewModel) GitRepository() vcs.Repository { return projectRepo{vm} }

// projectRepo is the project and the base it was loaded from.
type projectRepo struct{ vm *ViewModel }

// Root is "/": the project lives in the page, not under a directory.
func (projectRepo) Root() string { return "/" }

// Status is the working copy against the base, and against theirs when Check for
// changes found some. It runs on the UI goroutine (the tree is given no PostUI),
// so it reads the view-model without a lock.
func (r projectRepo) Status(ctx context.Context) (map[string]vcs.State, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	g := &r.vm.git
	if g.base == nil || g.baseSums == nil {
		return map[string]vcs.State{}, nil
	}
	return diffStatus(ctx, g.baseSums, r.vm.captureState().Files, g.theirs)
}

// diffStatus is the status of work against base, and of the conflicts with theirs
// (nil: none found). base holds a hash per path; work and theirs hold content.
func diffStatus(ctx context.Context, base map[string]string, work map[string][]byte, theirs remote.Files) (map[string]vcs.State, error) {
	out := map[string]vcs.State{}
	for p, b := range work {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if s, ok := base[p]; !ok {
			out[p] = vcs.Added
		} else if sumOf(b) != s {
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
		if inTheirs == inBase && (!inTheirs || sumOf(t) == s) {
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
